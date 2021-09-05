package main

import (
	"fmt"
	"html/template"
	"time"

	"github.com/rdmyldz/clone-hn/models"
)

type TmplData struct {
	Post        *models.Post
	Posts       []models.Post
	Comments    []models.Comment
	Comment     *models.Comment
	Username    string
	User        *models.User
	Indentation map[int]int
}

var tmplFunc = template.FuncMap{
	"formatDate": formatDate,
	"timeSince":  timeSince,
	"replyLink":  replyLink,
	"jsToggle":   jsToggle,
	"incIndex":   incIndex,
}

func formatDate(t time.Time) string {
	return t.Format("02 Jun 2006 at 15:04")
}

func timeSince(t time.Time) string {
	return time.Since(t).Round(time.Second).String()
}

func replyLink(pid, cid int) string {
	return fmt.Sprintf("/item/%d#%d", pid, cid)
}

func jsToggle(cid int) template.JS {
	return template.JS(fmt.Sprintf("return toggle(event, %d)", cid))
}

func incIndex(i int) int {
	return i + 1
}
