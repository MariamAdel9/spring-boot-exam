package posts

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"myapp/authentication"
	"myapp/tables"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// Socket for upload
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // Allow all origins
}

// Socket for fetching
var upgrader2 = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

var upgrader3 = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

var upgrader4 = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// Struct for uploading
type Post struct {
	Title      string `json:"title"`
	Content    string `json:"content"`
	Categories string `json:"categories"`
	Image      string `json:"image"` // base64 string
}

// Struct for fetching
type PostFetch struct {
	ID         int    `json:"id"`
	User       string `json:"user"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	Categories string `json:"categories"`
	Created    string `json:"created"`
	Image      string `json:"image"` // now string, not []byte
}

type CommentFetch struct {
	ID      int    `json:"id"`
	User    string `json:"user"`
	Content string `json:"content"`
	Created string `json:"created"`
}

// Decode and save base64 image to "uploads/" folder
func decodeAndSaveImage(base64Str string, filename string) error {
	if _, err := os.Stat("uploads"); os.IsNotExist(err) {
		if err := os.Mkdir("uploads", 0755); err != nil {
			return fmt.Errorf("failed to create uploads folder: %v", err)
		}
	}

	commaIdx := strings.Index(base64Str, ",")
	if commaIdx == -1 {
		return fmt.Errorf("invalid base64 string")
	}

	imageData := base64Str[commaIdx+1:]
	decoded, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		return err
	}

	return os.WriteFile("uploads/"+filename, decoded, 0644)
}

func Checksession(w http.ResponseWriter, r *http.Request) string {
	var username string
	db := tables.GetDB()

	for {

		cookiee, err := r.Cookie("session_token")
		if err != nil {
			return ""
		}

		sessionToken := cookiee.Value

		db.QueryRow("SELECT user FROM sessions WHERE session = ?", sessionToken).Scan(&username)
		return username

	}
}

// Upload handler (WebSocket)
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	db := tables.GetDB()

	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		http.Error(w, "No session token", http.StatusUnauthorized)
		return
	}

	sessionToken := cookie.Value

	var username string
	err = db.QueryRow("SELECT user FROM sessions WHERE session = ?", sessionToken).Scan(&username)
	if err != nil || username == "" {
		return
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "WebSocket upgrade failed", http.StatusInternalServerError)

	}

	defer conn.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket read error:", err)
			break
		}

		// Parse post from client
		var newPost Post
		if err := json.Unmarshal(message, &newPost); err != nil {
			conn.WriteMessage(websocket.TextMessage, []byte("Invalid JSON data"))
			continue
		}

		// Handle image if any
		imageFilename := ""
		if newPost.Image != "" {
			imageFilename = fmt.Sprintf("%d_%s.jpg", time.Now().Unix(), "upload")
			err = decodeAndSaveImage(newPost.Image, imageFilename)
			if err != nil {
				log.Println("Image decode error:", err)
				conn.WriteMessage(websocket.TextMessage, []byte("Image save failed"))
				continue
			}
		}

		cookie, err := r.Cookie("session_token")
		if err != nil || cookie.Value == "" {
			http.Error(w, "No session token", http.StatusUnauthorized)

			authentication.Logout(w, r)
			conn.Close()
			return
		}

		sessionToken := cookie.Value

		var username string
		err = db.QueryRow("SELECT user FROM sessions WHERE session = ?", sessionToken).Scan(&username)
		if err != nil || username == "" {
			conn.WriteMessage(websocket.TextMessage, []byte("session incorrect"))

			authentication.Logout(w, r)
			conn.Close()
			return

		}

		// Insert post into DB with the username from the session
		_, err = db.Exec(`INSERT INTO posts (user, title, content, categories, image) VALUES (?, ?, ?, ?, ?)`,
			username, newPost.Title, newPost.Content, newPost.Categories, imageFilename)

		if err != nil {
			log.Println("DB insert error:", err)
			conn.WriteMessage(websocket.TextMessage, []byte("Post insert failed"))
			continue
		}

		conn.WriteMessage(websocket.TextMessage, []byte("Post uploaded successfully"))
	}
}

func FetchHandler(w http.ResponseWriter, r *http.Request) {
	db := tables.GetDB()

	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		http.Error(w, "No session token", http.StatusUnauthorized)
		return
	}

	sessionToken := cookie.Value

	var username string
	err = db.QueryRow("SELECT user FROM sessions WHERE session = ?", sessionToken).Scan(&username)
	if err != nil || username == "" {
		return
	}

	conn, err := upgrader2.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade to WebSocket", http.StatusInternalServerError)
		return
	}

	lastSeenID := 0

	for {
		cookie, err := r.Cookie("session_token")
		if err != nil || cookie.Value == "" {
			http.Error(w, "No session token", http.StatusUnauthorized)

			authentication.Logout(w, r)
			conn.Close()
			return
		}

		sessionToken := cookie.Value

		var username string
		err = db.QueryRow("SELECT user FROM sessions WHERE session = ?", sessionToken).Scan(&username)
		if err != nil || username == "" {
			conn.WriteMessage(websocket.TextMessage, []byte("session incorrect"))

			authentication.Logout(w, r)
			conn.Close()
			return

		}
		rows, err := db.Query(`
			SELECT id, user, title, content, categories, dateCreation, image FROM posts WHERE id > ? ORDER BY id ASC`, lastSeenID)
		if err != nil {
			log.Println("DB query error:", err)
			time.Sleep(2 * time.Second)
			continue
		}

		var newPosts []PostFetch

		for rows.Next() {
			var p PostFetch
			var createdStr, filename string

			if err := rows.Scan(&p.ID, &p.User, &p.Title, &p.Content, &p.Categories, &createdStr, &filename); err != nil {
				log.Println("Row scan error:", err)
				continue
			}

			imgBytes, err := os.ReadFile("uploads/" + filename)
			if err == nil {
				p.Image = "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(imgBytes)
			}

			parsedTime, err := time.Parse(time.RFC3339, createdStr)
			if err == nil {
				p.Created = parsedTime.Local().Format("2006-01-02 15:04")
			} else {
				p.Created = createdStr
			}

			newPosts = append(newPosts, p)

			// ✅ Track the highest seen ID
			if p.ID > lastSeenID {
				lastSeenID = p.ID
			}
		}
		rows.Close()

		if len(newPosts) > 0 {
			if err := conn.WriteJSON(newPosts); err != nil {
				log.Println("WebSocket write error:", err)
				break
			}
		}

		time.Sleep(2 * time.Second)
	}
}

func CommentHandler(w http.ResponseWriter, r *http.Request) {
	db := tables.GetDB()

	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		http.Error(w, "No session token", http.StatusUnauthorized)
		return
	}

	sessionToken := cookie.Value

	var username string
	err = db.QueryRow("SELECT user FROM sessions WHERE session = ?", sessionToken).Scan(&username)
	if err != nil || username == "" {
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		ID      int    `json:"id"`
		Comment string `json:"comment"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if data.Comment == "" {
		http.Error(w, "Empty comment", http.StatusBadRequest)
		return
	}

	var user string
	err = db.QueryRow("SELECT user FROM sessions WHERE session = ?", cookie.Value).Scan(&user)
	if err != nil {
		http.Error(w, "Invalid session", http.StatusForbidden)
		return
	}

	_, err = db.Exec("INSERT INTO comments (post_id, user, content, dateCreation) VALUES (?, ?, ?, ?)",
		data.ID, user, data.Comment, time.Now().Format(time.RFC3339))
	if err != nil {
		http.Error(w, "Failed to save comment", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Comment added"))
}

func ViewComment(w http.ResponseWriter, r *http.Request) {
	db := tables.GetDB()

	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		http.Error(w, "No session token", http.StatusUnauthorized)
		return
	}

	sessionToken := cookie.Value

	var username string
	err = db.QueryRow("SELECT user FROM sessions WHERE session = ?", sessionToken).Scan(&username)
	if err != nil || username == "" {
		return
	}

	conn, err := upgrader4.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade to WebSocket", http.StatusInternalServerError)
		return
	}
	defer conn.Close()
	// Read first message to get the Post ID
	_, msg, err := conn.ReadMessage()

	if err != nil {
		fmt.Print(msg)
		log.Println("Failed to read initial post ID message:", err)
		return
	}

	var input struct {
		ID int `json:"id"`
	}

	if err := json.Unmarshal(msg, &input); err != nil || input.ID <= 0 {
		log.Println("Invalid post ID received:", err)
		conn.WriteMessage(websocket.TextMessage, []byte("Invalid post ID"))
		return
	}

	postID := input.ID
	log.Println(postID)
	lastSeenID := 0

	for {
		rows, err := db.Query(`
			SELECT id, user, content, dateCreation 
			FROM comments 
			WHERE post_id = ? AND id > ? 
			ORDER BY id ASC`, postID, lastSeenID)
		if err != nil {
			log.Println("DB query error:", err)
			time.Sleep(2 * time.Second)
			continue
		}

		var newComments []CommentFetch

		for rows.Next() {
			var c CommentFetch
			var createdStr string

			if err := rows.Scan(&c.ID, &c.User, &c.Content, &createdStr); err != nil {
				log.Println("Row scan error:", err)
				continue
			}

			if parsedTime, err := time.Parse(time.RFC3339, createdStr); err == nil {
				c.Created = parsedTime.Local().Format("2006-01-02 15:04")
			} else {
				c.Created = createdStr
			}

			newComments = append(newComments, c)
			if c.ID > lastSeenID {
				lastSeenID = c.ID
			}
		}
		rows.Close()

		if len(newComments) > 0 {
			if err := conn.WriteJSON(newComments); err != nil {
				log.Println("WebSocket write error:", err)
				break
			}
		}

		time.Sleep(2 * time.Second)
	}
}


