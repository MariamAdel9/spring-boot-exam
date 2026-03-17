package main

import (
	"database/sql"
	"log"
	"myapp/authentication"
	"myapp/chats"
	"myapp/posts"
	"myapp/tables"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB // shared database connection

// func main() {
// 	var err error
// 	DB, err = sql.Open("sqlite3", "./mydb.db")
// 	if err != nil {
// 		log.Fatal("Failed to open DB:", err)
// 	}
// 	defer DB.Close()

// 	tables.InitDB(DB) // set DB globally in tables

// 	// Serve all files from ../static/ at root path "/"
// 	fs := http.FileServer(http.Dir("../static"))
// 	http.Handle("/", fs)

// 	// API endpoint for user registration
// 	http.HandleFunc("/api/checksession", authentication.Checksession)
// 	http.HandleFunc("/api/login", authentication.LoginHandler)
//http.HandleFunc("/api/upload", posts.UploadHandler)

//		log.Println("Server running at http://localhost:8080")
//		log.Fatal(http.ListenAndServe(":8080", nil))
//	}
func main() {
	var err error
	DB, err = sql.Open("sqlite3", "./mydb.db")
	if err != nil {
		log.Fatal("Failed to open DB:", err)
	}
	defer DB.Close()

	tables.InitDB(DB)

	// API endpoint for user registration
	http.HandleFunc("/api/register", authentication.RegisterHandler)
	http.HandleFunc("/api/login", authentication.LoginHandler)
	//http.HandleFunc("/api/upload", posts.UploadHandler)
	http.HandleFunc("/api/fetch", posts.FetchHandler)
	http.HandleFunc("/api/upload", posts.UploadHandler)
	http.HandleFunc("/api/comment", posts.CommentHandler)
	http.HandleFunc("/api/commentfetch", posts.ViewComment)
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))
	http.HandleFunc("/api/chat", chats.All)
	// http.HandleFunc("/api/chat", chats.DisplayChat)
	http.HandleFunc("/api/logout", authentication.Logout)
	http.HandleFunc("/api/checksession", authentication.Checksession)
	http.HandleFunc("/api/onlineuser", authentication.ChatHandler)

	fs := http.FileServer(http.Dir("../static"))
	http.Handle("/", fs)
	http.HandleFunc("/ws", authentication.WsHandler)

	log.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
