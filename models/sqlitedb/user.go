package sqlitedb

import (
	"fmt"
	"time"

	"github.com/rdmyldz/clone-hn/models"
	"golang.org/x/crypto/bcrypt"
)

func (s *SqliteHN) InsertUser(username, password string) (int, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return -1, fmt.Errorf("in InsertUser, error while generating hashedPassword: %w", err)
	}

	stmt, err := s.db.Prepare("INSERT INTO users (user_name, password, created_at) values(?, ?, ?)")
	if err != nil {
		return -1, fmt.Errorf("in InsertUser, error while preparint statement: %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(username, string(hashedPassword), time.Now())
	if err != nil {
		return -1, fmt.Errorf("in InsertUser, error while executing statement: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return -1, fmt.Errorf("in InsertUser, error while getting lastInsertedID: %w", err)
	}
	return int(id), nil
}

func (s *SqliteHN) GetUser(query, username string) (*models.User, error) {
	row := s.db.QueryRow(query, username)
	var u models.User
	err := row.Scan(&u.ID, &u.Username, &u.Password, &u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("in GetUser, error while scanning row: %w", err)
	}

	return &u, nil
}

func (s *SqliteHN) Authenticate(email, password string) (int, error) {
	var id int
	var hashedPassword []byte
	stmt := "SELECT user_id, password FROM users WHERE user_name = ?"
	row := s.db.QueryRow(stmt, email)
	err := row.Scan(&id, &hashedPassword)
	if err != nil {
		return -1, fmt.Errorf("in Authenticate, error while scanning row: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(password)); err != nil {
		return -1, fmt.Errorf("in Authenticate, password didn't match: %w", err)
	}

	return id, nil
}
