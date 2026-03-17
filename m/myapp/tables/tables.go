package tables

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

var err error

func InitDB(database *sql.DB) {
	db = database

	db, err = sql.Open("sqlite3", "./tables/appDB.db")
	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}

	Createtables()

}

func Createtables() {

	//db.Exec(`DROP TABLE sessions`)

	// db.Exec(`DROP TABLE comment`)

	_, err = db.Exec(`
   CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    nickname TEXT NOT NULL,
    firstname TEXT NOT NULL,
    lastname TEXT NOT NULL,
    age INTEGER NOT NULL,
    gender INTEGER NOT NULL,
    email TEXT NOT NULL,
    password TEXT NOT NULL
    );

	CREATE TABLE IF NOT EXISTS sessions (
		user TEXT PRIMARY KEY NOT NULL,
		session TEXT NOT NULL, 
    lastactive TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	
	CREATE TABLE IF NOT EXISTS posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
	user TEXT NOT NULL,
	categories VARCHAR(255) DEFAULT NULL, 
    title char(50) NOT NULL,
    content TEXT NOT NULL, 
    image TEXT NULL,
    dateCreation TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    userID INTEGER,
    FOREIGN KEY (userID) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE 
);

	CREATE TABLE IF NOT EXISTS comments (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  post_id INTEGER,
  user TEXT,
  content TEXT,
  dateCreation DATETIME
);


CREATE TABLE IF NOT EXISTS chat (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user1 TEXT NOT NULL ,
  user2 TEXT NOT NULL,
  content TEXT, 
dateCreation TIMESTAMP DEFAULT CURRENT_TIMESTAMP )





`)

	if err != nil {
		log.Fatal("Failed to create table:", err)
	}

}

func GetDB() *sql.DB {
	return db
}
