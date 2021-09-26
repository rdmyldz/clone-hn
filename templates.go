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
	FrontDates  map[string]string
	CsrfToken   map[string]interface{}
	Page        int
}

var tmplFunc = template.FuncMap{
	"formatDatetime": formatDatetime,
	"timeSince":      timeSince,
	"replyLink":      replyLink,
	"jsToggle":       jsToggle,
	"incIndex":       incIndex,
	"getDate":        getDate,
	"htmlString":     htmlString,
}

func htmlString(text string) template.HTML {
	return template.HTML(text)
}

func formatDatetime(t time.Time) string {
	return t.Format("2006-01-02T15:04:05")
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

func incIndex(i, p int) int {
	return i + 1 + ((p - 2) * limit)
}

func getDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func getDatesForFrontPage(str string) (map[string]string, error) {
	dates := make(map[string]string)
	dates["titleDate"] = str
	tnow := time.Now().UTC()
	t, err := time.Parse("2006-01-02", str)
	if err != nil {
		return nil, fmt.Errorf("error while parsing %s: %v", str, err)
	}
	dates["storiesFrom"] = t.Format("January 02, 2006")
	dates["oneDayAgo"] = t.AddDate(0, 0, -1).Format("2006-01-02")
	dates["oneMonthAgo"] = t.AddDate(0, -1, 0).Format("2006-01-02")
	dates["oneYearAgo"] = t.AddDate(-1, 0, 0).Format("2006-01-02")

	oneDayAfter := t.AddDate(0, 0, +1)
	if oneDayAfter.After(tnow) {
		return dates, nil
	}
	dates["oneDayAfter"] = oneDayAfter.Format("2006-01-02")

	oneMonthAfter := t.AddDate(0, +1, 0)
	if oneMonthAfter.After(tnow) {
		return dates, nil
	}
	dates["oneMonthAfter"] = oneMonthAfter.Format("2006-01-02")

	oneYearAfter := t.AddDate(+1, 0, 0)
	if oneYearAfter.After(tnow) {
		return dates, nil
	}
	dates["oneYearAfter"] = oneYearAfter.Format("2006-01-02")
	return dates, nil
}
