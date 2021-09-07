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
}

type Comment struct {
	ID        int
	Text      string
	CreatedAt time.Time
	Owner     string
	ParentID  int
	PostID    int
}

type User struct {
	ID        int
	Username  string
	Password  string
	CreatedAt time.Time
}
