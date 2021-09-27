package models

import "time"

type Post struct {
	ID           int
	Link         string
	Title        string
	Domain       string
	CreatedAt    time.Time
	Owner        string
	Points       int
	ParentID     int
	MainPostID   int
	CommentNum   int
	TitleSummary string
	Text         string
}

type User struct {
	ID        int
	Username  string
	Password  string
	CreatedAt time.Time
}
