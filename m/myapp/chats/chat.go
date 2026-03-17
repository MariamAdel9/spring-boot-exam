package chats

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"myapp/tables"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func Chat(w http.ResponseWriter, r *http.Request) {
	fmt.Println("function called meow")

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
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "WebSocket upgrade failed", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	for {

		fmt.Println("function started")

		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("WebSocket read error:", err)
			break
		}

		fmt.Println("reading somting")
		var data struct {
			To   string `json:"to"`
			Chat string `json:"text"`
		}

		if err := json.Unmarshal(message, &data); err != nil {
			conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"Invalid JSON data"}`))
			continue
		}
		fmt.Println(data.Chat)

		if data.Chat == "" {
			conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"Empty chat message"}`))
			continue
		}

		// Validate recipient exists in users table (not sessions)
		var to string
		err = db.QueryRow("SELECT nickname FROM users WHERE nickname = ?", data.To).Scan(&to)
		if err != nil {
			fmt.Println("error Recipient not found")
			continue
		}

		// Check if recipient online - your existing logic

		// Save chat
		_, err = db.Exec("INSERT INTO chat (user1, user2, content) VALUES (?, ?, ?)", username, to, data.Chat)
		if err != nil {
			fmt.Println("error Failed to save message")
			continue
		}
		fmt.Println("inserted meoooooooooowwwwwwwwww")

		// Optionally broadcast message to recipient if online
	}
}

var upgrader2 = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type fetchChat struct {
	ID      int    `json:"id"`
	User1   string `json:"user1"`
	User2   string `json:"user2"`
	Content string `json:"content"`
}

func DisplayChat(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader2.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "WebSocket upgrade failed", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	var currentChatNick string
	lastSeenID := 0

	for {
		// Read client message
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("WebSocket read error:", err)
			break
		}

		// Detect if message is string (nick) or JSON (chat message)
		if len(message) > 0 && message[0] == '"' {
			// Assume string = nick to fetch history
			err := json.Unmarshal(message, &currentChatNick)
			if err != nil {
				log.Println("JSON unmarshal nick error:", err)
				continue
			}

			// Fetch all previous chat messages between username and currentChatNick
			rows, err := db.Query(`
                SELECT id, user1, user2, content FROM chat
                WHERE ((user1 = ? AND user2 = ?) OR (user1 = ? AND user2 = ?))
                ORDER BY id ASC`, username, currentChatNick, currentChatNick, username)
			if err != nil {
				log.Println("DB query error:", err)
				continue
			}

			chats := []fetchChat{}
			lastSeenID = 0
			for rows.Next() {
				var c fetchChat
				if err := rows.Scan(&c.ID, &c.User1, &c.User2, &c.Content); err != nil {
					log.Println("Row scan error:", err)
					continue
				}
				chats = append(chats, c)
				if c.ID > lastSeenID {
					lastSeenID = c.ID
				}
			}
			rows.Close()

			respData := struct {
				Type     string      `json:"type"`
				Messages []fetchChat `json:"messages"`
			}{
				Type:     "history",
				Messages: chats,
			}

			resp, err := json.Marshal(respData)
			if err != nil {
				log.Println("JSON marshal error:", err)
				continue
			}
			conn.WriteMessage(websocket.TextMessage, resp)

		} else {
			// Otherwise, parse as new message JSON
			var incoming struct {
				To   string `json:"to"`
				Text string `json:"text"`
			}
			if err := json.Unmarshal(message, &incoming); err != nil {
				log.Println("JSON unmarshal message error:", err)
				continue
			}
			if incoming.Text == "" {
				// Ignore empty messages
				continue
			}

			// Insert new message into DB
			res, err := db.Exec("INSERT INTO chat (user1, user2, content) VALUES (?, ?, ?)", username, incoming.To, incoming.Text)
			if err != nil {
				log.Println("DB insert error:", err)
				continue
			}

			lastInsertID, err := res.LastInsertId()
			if err != nil {
				log.Println("Getting last insert ID failed:", err)
				continue
			}

			// Send back the new message to client so they see it instantly
			newMsg := struct {
				Type string `json:"type"`
				From string `json:"from"`
				To   string `json:"to"`
				Text string `json:"text"`
				ID   int64  `json:"id"`
			}{
				Type: "new_message",
				From: username,
				To:   incoming.To,
				Text: incoming.Text,
				ID:   lastInsertID,
			}

			resp, err := json.Marshal(newMsg)
			if err != nil {
				log.Println("JSON marshal error:", err)
				continue
			}
			conn.WriteMessage(websocket.TextMessage, resp)
		}
	}
	fmt.Println("Client disconnected")
}

func All(w http.ResponseWriter, r *http.Request){
	DisplayChat(w, r)
	Chat(w,r )
}
// func DisplayChat(w http.ResponseWriter, r *http.Request) {
// 	fmt.Println("displayChat called")

// 	db := tables.GetDB()
// 	cookie, err := r.Cookie("session_token")
// 	if err != nil || cookie.Value == "" {
// 		http.Error(w, "No session token", http.StatusUnauthorized)
// 		return
// 	}
// 	sessionToken := cookie.Value

// 	var username string
// 	err = db.QueryRow("SELECT user FROM sessions WHERE session = ?", sessionToken).Scan(&username)
// 	if err != nil || username == "" {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}

// 	conn, err := upgrader2.Upgrade(w, r, nil)
// 	if err != nil {
// 		http.Error(w, "WebSocket upgrade failed", http.StatusInternalServerError)
// 		return
// 	}
// 	defer conn.Close()

// 	for {

// 		lastSeenID := 0
// 		// Check session cookie on every loop to catch logout
// 		cookie, err := r.Cookie("session_token")
// 		if err != nil || cookie.Value == "" {
// 			authentication.Logout(w, r)
// 			conn.Close()
// 			return
// 		}
// 		sessionToken = cookie.Value

// 		err = db.QueryRow("SELECT user FROM sessions WHERE session = ?", sessionToken).Scan(&username)
// 		if err != nil || username == "" {
// 			// Notify client, then logout and close
// 			authentication.Logout(w, r)
// 			conn.Close()
// 			return
// 		}

// 		// Read message from client - expected to be a JSON string representing the nick to fetch chat with
// 		_, message, err := conn.ReadMessage()
// 		if err != nil {
// 			fmt.Println("WebSocket read error:", err)
// 			break
// 		}

// 		var nick string
// 		if err := json.Unmarshal(message, &nick); err != nil {
// 			log.Println("JSON unmarshal error:", err)
// 			continue
// 		}

// 		// Query chat messages between current user and nick where id > lastSeenID
// 		rows, err := db.Query(`
//             SELECT id, user1, user2, content FROM chat
//             WHERE id > ? AND ((user1 = ? AND user2 = ?) OR (user1 = ? AND user2 = ?))
//             ORDER BY id ASC`, lastSeenID, username, nick, nick, username)
// 		if err != nil {
// 			log.Println("DB query error:", err)
// 			time.Sleep(2 * time.Second)
// 			continue
// 		}

// 		var newChats []fetchChat

// 		for rows.Next() {
// 			var c fetchChat
// 			err := rows.Scan(&c.ID, &c.User1, &c.User2, &c.Content)
// 			if err != nil {
// 				log.Println("Row scan error:", err)
// 				continue
// 			}
// 			newChats = append(newChats, c)

// 			// Update lastSeenID to the highest seen
// 			if c.ID > lastSeenID {
// 				lastSeenID = c.ID
// 			}
// 		}

// 		// Send new chat messages to client as JSON
// 		if len(newChats) > 0 {
// 			fmt.Println("new chat are inwdencnwedb jfhcbgj")
// 			resp, err := json.Marshal(newChats)
// 			if err != nil {
// 				log.Println("JSON marshal error:", err)
// 				continue
// 			}
// 			err = conn.WriteMessage(websocket.TextMessage, resp)
// 			if err != nil {
// 				log.Println("WebSocket write error:", err)
// 				break
// 			}
// 		}

// 		// Optional: if no new messages, could send heartbeat or wait for new client message
// 	}
// }
