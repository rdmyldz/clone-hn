package sqlitedb

import (
	"fmt"
	"time"

	"github.com/rdmyldz/clone-hn/models"
)

func (s *SqliteHN) InsertUser(username, password string) (int, error) {
	stmt, err := s.db.Prepare("INSERT INTO users (user_name, password, created_at) values(?, ?, ?)")
	if err != nil {
		return -1, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(username, password, time.Now())
	if err != nil {
		return -1, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}
	return int(id), nil
}

func (s *SqliteHN) GetUser(query, username string) (*models.User, error) {
	row := s.db.QueryRow(query, username)
	var u models.User
	err := row.Scan(&u.ID, &u.Username, &u.Password, &u.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (s *SqliteHN) Authenticate(email, password string) (int, error) {
	var id int
	var pswd string
	stmt := "SELECT user_id, password FROM users WHERE user_name = ?"
	row := s.db.QueryRow(stmt, email)
	err := row.Scan(&id, &pswd)
	if err != nil {
		return -1, err
	}

	if password != pswd {
		return -1, fmt.Errorf("password didn't match")
	}

	return id, nil
}