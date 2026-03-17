package authentication

import (
	"encoding/json"
	"fmt"
	"sync"

	"log"
	"myapp/tables"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type User struct {
	Nickname  string `json:"nickname"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Age       int    `json:"age"`
	Gender    int    `json:"gender"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

func CreateUser(nickname string, firstname string, lastname string, gender int, email string, password string, age int) error {
	db := tables.GetDB()

	_, err := db.Exec(`
		INSERT INTO users (nickname, firstname, lastname, gender,age, email, password)
		VALUES (?, ?, ?, ?, ?, ?, ?)`, nickname, firstname, lastname, gender, age, email, password,
	)

	return err
}

type tosend struct {
	Message string `json:"message"`
}

type User2 struct {
	Nickname string `json:"nickname"`
	Password string `json:"password"`
}

func Checksession(w http.ResponseWriter, r *http.Request) {

	type Verify struct {
		Valid    bool   `json:"valid"`
		Username string `json:"username,omitempty"`
	}

	db := tables.GetDB()
	valid := false
	var username string

	// Try to read session cookie
	cookie, err := r.Cookie("session_token")
	if err == nil {
		sessionToken := cookie.Value
		err := db.QueryRow("SELECT user FROM sessions WHERE session = ?", sessionToken).Scan(&username)
		if err == nil && username != "" {
			valid = true
		} else {
			http.SetCookie(w, &http.Cookie{
				Name:     "session_token",
				Value:    "",
				Path:     "/",
				Expires:  time.Unix(0, 0),
				MaxAge:   -1, // 👈 tells browser to delete it immediately
				HttpOnly: true,
				Secure:   false, // set to true if using https
			})

		}
	}

	// Build response
	response := Verify{Valid: valid, Username: username}

	// Headers
	w.Header().Set("Content-Type", "application/json")
	// If your frontend is same origin, no CORS needed. If not:
	// w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	// w.Header().Set("Access-Control-Allow-Credentials", "true")

	json.NewEncoder(w).Encode(response)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	db := tables.GetDB()
	var u User2
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	name := u.Nickname
	password := u.Password

	if name == "" || password == "" {
		http.Error(w, "Nickname and password cannot be empty", http.StatusBadRequest)
		return
	}

	// Check if user exists with correct credentials
	var namedb string
	err := db.QueryRow(`SELECT nickname FROM users WHERE nickname = ? AND password = ?`, name, password).Scan(&namedb)
	if err != nil {
		http.Error(w, "User does not exist or password is invalid", http.StatusUnauthorized)
		return
	}

	// Remove any existing sessions for this user
	_, err = db.Exec("DELETE FROM sessions WHERE user = ?", name)
	if err != nil {
		log.Printf("Failed to delete old session: %v", err)
	}

	// Create a new session token
	sessionToken := uuid.NewString()

	// Insert new session into DB
	row := db.QueryRow("SELECT 1 FROM sessions WHERE user = ? LIMIT 1", namedb)
	err = row.Err()
	if err == nil {
		_, err = db.Exec("DELETE FROM sessions WHERE user = ?", namedb)
		if err != nil {
			log.Printf("Failed to delete old session: %v", err)
		}
	}

	_, err = db.Exec("INSERT INTO sessions (user, session) VALUES (?, ?)", name, sessionToken)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	if cookie, err := r.Cookie("session_token"); err == nil && cookie.Value != "" {
		sessionToken := cookie.Value
		_, err := db.Exec("DELETE FROM sessions WHERE session = ?", sessionToken)
		if err != nil {
			log.Printf("Failed to delete session: %v", err)
		}
	}

	// Set the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Path:     "/", // This is important!
	})

	// Return response
	sending := tosend{Message: namedb}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sending)

}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)

		return
	}

	var u User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err := CreateUser(
		u.Nickname,
		u.Firstname,
		u.Lastname,
		u.Gender,
		u.Email,
		u.Password,
		u.Age,
	)

	if err != nil {
		http.Error(w, "User creation failed", http.StatusInternalServerError)
		log.Println("CreateUser error:", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User created successfully"))
}

func Logout(w http.ResponseWriter, r *http.Request) {

	fmt.Print("✅ Logged out successfully")
	fmt.Print("✅ Logged out successfully")
	fmt.Print("✅ Logged out successfully")
	fmt.Print("✅ Logged out successfully")

	db := tables.GetDB()

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		http.Error(w, "Your already logged out", http.StatusForbidden)
		return
	}

	t, err := db.Exec("DELETE FROM sessions WHERE session = ?", cookie.Value)
	if err != nil {
		http.Error(w, "failed to delete your session", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1, // 👈 tells browser to delete it immediately
		HttpOnly: true,
		Secure:   false, // set to true if using https
	})

	rows, err := t.RowsAffected()
	if err != nil {
		http.Error(w, "could not check rows affected", http.StatusInternalServerError)
		return
	}

	if rows > 0 {
		w.WriteHeader(http.StatusOK)

	} else {
		http.Error(w, "no session found to delete", http.StatusNotFound)
	}

	w.WriteHeader(http.StatusOK)
}

const (
	pingPeriod = 40 * time.Second
	pongWait   = 40 * time.Second
)

var (
	upgrader   = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	sessions   = make(map[string]*websocket.Conn)
	sessionsMu sync.Mutex
)

func WsHandler(w http.ResponseWriter, r *http.Request) {
	db := tables.GetDB()

	// 1) Authenticate via cookie
	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		http.Error(w, "No session token", http.StatusUnauthorized)
		return
	}
	token := cookie.Value

	var user string
	if err := db.QueryRow("SELECT user FROM sessions WHERE session = ?", token).Scan(&user); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2) Upgrade to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Upgrade error:", err)
		return
	}

	// Track connection
	sessionsMu.Lock()
	sessions[user] = conn
	sessionsMu.Unlock()
	fmt.Printf("→ %s connected\n", user)

	// 3) Configure pong handler to reset read deadline
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(appData string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// 4) Ping loop
	go func() {
		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()

		for range ticker.C {
			// send ping
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				fmt.Printf("Ping error for %s: %v\n", user, err)

				return
			}
		}
	}()

	// 5) Read loop → will time out if no pong or message within pongWait
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			fmt.Printf("Read error %s: %v\n", user, err)
			Logout(w, r)
			return
		}
		// any received message also resets the deadline via SetReadDeadline above
	}
}

func ChatHandler(w http.ResponseWriter, r *http.Request) {
	db := tables.GetDB()

	// 1) Authenticate via cookie
	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	token := cookie.Value

	var user string
	if err := db.QueryRow("SELECT user FROM sessions WHERE session = ?", token).Scan(&user); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2) Upgrade to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Upgrade error:", err)
		return
	}

	// 3) Register the connection
	sessionsMu.Lock()
	sessions[user] = conn
	sessionsMu.Unlock()

	// 4) Broadcast updated list
	broadcastOnlineUsers()

	// 5) Ensure we clean up on exit
	defer func() {
		// Remove from map
		sessionsMu.Lock()
		delete(sessions, user)
		sessionsMu.Unlock()

		// Notify everyone
		broadcastOnlineUsers()

		// Perform your logout steps

	}()

	// 6) Read loop just blocks until client disconnects
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			// client disconnected or error
			return
		}
	}
}

// // Helper to send current user list to every client
// func broadcastOnlineUsers() {

// 	db := tables.GetDB()

// 	sessionsMu.Lock()
// 	defer sessionsMu.Unlock()

// 	// Build list of usernames
// 	userList := make([]string, 0, len(sessions))
// 	for u := range sessions {
// 		userList = append(userList, u)
// 	}

// 	// Message payload
// 	msg := map[string]interface{}{
// 		"type":  "online_users",
// 		"users": userList,
// 	}

// 	fmt.Println(userList)
// 		fmt.Println("------------------------------------")

// 		fmt.Println(msg)


// 	// Send to each client
// 	for _, conn := range sessions {
// 		_ = conn.WriteJSON(userList)
// 	}



// 	rows, err := db.Query(`SELECT nickname FROM users`)
// 	if err !=nil{
// 		fmt.Println(err)
// 	}
	
	
// 	var allusers []string

// 	for rows.Next(){
// 		var nickname string
// 		err = rows.Scan(&nickname)
// 		if err != nil {
// 			fmt.Println(err)
// 		}
// 		allusers = append(allusers, nickname)
// 	}

// 	fmt.Println(allusers)

	
	
// }

// Helper to send current user list (with online/offline) to every client


var payload map[string]interface{}


func broadcastOnlineUsers() {
    db := tables.GetDB()

    sessionsMu.Lock()
    defer sessionsMu.Unlock()

    // 1) Get all users from DB
    rows, err := db.Query(`SELECT nickname FROM users`)
    if err != nil {
        fmt.Println("Error fetching users:", err)
        return
    }
    defer rows.Close()

    // 2) Build slice with online status

	 type UserStatus struct {
        Nickname string `json:"nickname"`
        Online   bool   `json:"online"`
    }
   
    var allUsers []UserStatus
    for rows.Next() {
        var nickname string
        if err := rows.Scan(&nickname); err != nil {
            fmt.Println("Scan error:", err)
            continue
        }
        _, isOnline := sessions[nickname]
        allUsers = append(allUsers, UserStatus{
            Nickname: nickname,
            Online:   isOnline,
        })
    }

    // 3) Send full list to every connected client
    payload = map[string]interface{}{
        "type":  "users_status",
        "users": allUsers,
    }

    for _, conn := range sessions {
        if err := conn.WriteJSON(payload); err != nil {
            fmt.Println("Write error:", err)
        }
    }

    // Debug output
    fmt.Println("Full user list with status:", allUsers)
}


func GetMap()map[string]interface{} {
	return payload 
}