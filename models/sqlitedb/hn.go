package sqlitedb

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3" // sql behavior modified
)

type SqliteHN struct {
	db *sql.DB
}

func NewDB(dbFile string) (*SqliteHN, error) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, fmt.Errorf("in NewDB, error while opening db: %w", err)
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
			"parent_id" INTEGER NOT NULL,
			"main_post_id" INTEGER NOT NULL,
			"comment_num" INTEGER NOT NULL,
			"title_summary" TEXT NOT NULL,
			"created_at" DATETIME NOT NULL,
			"text" TEXT NOT NULL
			);

		CREATE TABLE IF NOT EXISTS users ( 
			"user_id" INTEGER PRIMARY KEY AUTOINCREMENT, 
			"user_name" TEXT NOT NULL UNIQUE, 
			"password" TEXT NOT NULL, 
			"created_at" DATETIME NOT NULL
			);
		
	`)
	if err != nil {
		return nil, fmt.Errorf("in NewDB, error while creating tables: %w", err)
	}

	return &SqliteHN{db: db}, nil
}
