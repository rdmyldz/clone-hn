package sqlitedb

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3" // sql behavior modified
)

type SqliteHN struct {
	db *sql.DB
}

func NewDB(dbFile string) (*SqliteHN, error) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			db.Close()
		}
	}()
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS posts ( 
			"post_id" INTEGER PRIMARY KEY AUTOINCREMENT, 
			"link" TEXT NOT NULL, 
			"title" TEXT NOT NULL,
			"domain" TEXT NOT NULL,
			"owner" TEXT NOT NULL,
			"points" INTEGER NOT NULL,
			"created_at" DATETIME NOT NULL
			);

		CREATE TABLE IF NOT EXISTS comments ( 
			"comment_id" INTEGER PRIMARY KEY AUTOINCREMENT, 
			"text" TEXT NOT NULL, 
			"created_at" DATETIME NOT NULL,
			"owner" TEXT NOT NULL,
			"parent_id" int NOT NULL,
			"post_id" INTEGER NOT NULL,
			FOREIGN KEY ("post_id") REFERENCES "posts" ("post_id")
			);
		
		CREATE TABLE IF NOT EXISTS users ( 
			"user_id" INTEGER PRIMARY KEY AUTOINCREMENT, 
			"user_name" TEXT NOT NULL UNIQUE, 
			"password" TEXT NOT NULL, 
			"created_at" DATETIME NOT NULL
			);
		
	`)

	return &SqliteHN{db: db}, err
}
