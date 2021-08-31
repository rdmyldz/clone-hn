package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/rdmyldz/clone-hn/models"
)

func (app *application) handleHome(w http.ResponseWriter, r *http.Request) {
	log.Println("in handleHome")
	uname, err := getUsername(w, r)
	if err != nil {
		log.Printf("error returned from getusername: %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	query := "SELECT post_id, link, title, domain, owner, points, parent_id, created_at FROM posts WHERE parent_id = 0"
	posts, err := app.db.GetPosts(query)
	if err != nil {
		log.Printf("in handleHome, error while getting posts. err: %v\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	data := &TmplData{
		Posts:    posts,
		Username: uname,
	}
	app.tmpl.ExecuteTemplate(w, "home.html", data)
}

func (app *application) handleNews(w http.ResponseWriter, r *http.Request) {
	app.handleHome(w, r)
}

func (app *application) handleNewest(w http.ResponseWriter, r *http.Request) {
	uname, err := getUsername(w, r)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	query := "SELECT post_id, link, title, domain, owner, points, parent_id, created_at FROM posts Where parent_id = 0 ORDER BY created_at DESC"
	posts, err := app.db.GetPosts(query)
	if err != nil {
		log.Printf("in handleNewest, error while getting posts. err: %v\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	data := &TmplData{
		Posts:    posts,
		Username: uname,
	}
	app.tmpl.ExecuteTemplate(w, "home.html", data)
}

func (app *application) handleItem(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	log.Printf("vars: %v\n", vars)

	id := vars["id"]
	uname, err := getUsername(w, r)
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("/item/%s", id), http.StatusSeeOther)
		return
	}
	log.Printf("id: %v\n", id)
	queryPost := "SELECT post_id, link, title, domain, owner, points, parent_id, created_at FROM posts WHERE post_id = ?"
	p, err := app.db.GetPost(queryPost, id)
	if err != nil {
		log.Printf("in handleItem, error while getting post. err: %v\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// queryComment := "SELECT post_id, link, domain, owner, points, parent_id, created_at FROM posts WHERE parent_id = ?"
	// SELECT * FROM posts WHERE parent_id = 1 OR parent_id IN (SELECT post_id FROM posts WHERE parent_id = 1);
	query := fmt.Sprintf(`SELECT post_id, link, title, domain, owner, points, parent_id, created_at FROM posts where parent_id = "%d" OR parent_id IN (SELECT post_id FROM posts WHERE parent_id = "%d")`, p.ID, p.ID)
	comments, err := app.db.GetPosts(query)
	if err != nil {
		log.Printf("in handleItem, error while getting comments. err: %v\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	data := &TmplData{
		Post:     p,
		Posts:    comments,
		Username: uname,
	}

	app.tmpl.ExecuteTemplate(w, "item.html", data)
}

func (app *application) handleSubmit(w http.ResponseWriter, r *http.Request) {
	_, err := getUsername(w, r)
	if err != nil {
		http.Redirect(w, r, "/submit", http.StatusSeeOther)
		return
	}
	app.tmpl.ExecuteTemplate(w, "submit.html", nil)
}

func (app *application) handleR(w http.ResponseWriter, r *http.Request) {
	username, err := getUsername(w, r)
	log.Printf("username is: %v\n", username)
	if err != nil {
		http.Redirect(w, r, "/submit", http.StatusSeeOther)
		return
	}

	var post models.Post
	post.Title = r.PostFormValue("title")
	post.Link = r.PostFormValue("url")
	text := r.PostFormValue("text")
	if text != "" {
		post.Title = text
	}
	post.Domain = getDomain(post.Link)
	post.CreatedAt = time.Now()
	if err != nil {
		log.Printf("in handleR, error while parsing time: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	post.Owner = username
	log.Printf("owner is: %v\n", post.Owner)

	id, err := app.db.CreatePost(&post)
	if err != nil {
		log.Printf("in handleR, error while creating post. err: %v\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("post created. id: %v\n", id)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) handleFrom(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	site := vars["site"]

	query := fmt.Sprintf(`SELECT post_id, link, title, domain, owner, points, parent_id, created_at FROM posts where domain = "%s"`, site)
	posts, err := app.db.GetPosts(query)
	if err != nil {
		log.Printf("in handleFrom, error while getting posts. err: %v\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	data := TmplData{
		Posts: posts,
	}

	app.tmpl.ExecuteTemplate(w, "home.html", data)
}

func (app *application) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		app.tmpl.ExecuteTemplate(w, "login.html", nil)
		return
	}

	username := r.PostFormValue("acct")
	password := r.PostFormValue("pw")
	creating := r.PostFormValue("creating")
	log.Printf("u: %q - p: %q - creat: %q", username, password, creating)
	if creating == "t" {
		_, err := app.db.InsertUser(username, password)
		if err != nil {
			log.Printf("in handleLogin, error while inserting user: %v\n", err)
			http.Redirect(w, r, "/login", http.StatusPermanentRedirect)
			return
		}
		log.Printf("user inserted successfully!!- %s -%s\n", username, password)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	uid, err := app.db.Authenticate(username, password)
	if err != nil {
		log.Printf("in handleLogin, error while authenticating the user: %s - %v\n", username, err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	session, err := store.Get(r, "newsession")
	if err != nil {
		log.Printf("in handleLogin, error while getting session: %v\n", err)
		clearSession(session, w, r)
		return
	}

	session.Values["username"] = username
	session.Values["authenticated"] = true
	err = session.Save(r, w)
	if err != nil {
		log.Printf("in handleLogin, error while saving session: %v\n", err)
		http.Redirect(w, r, "/login", http.StatusPermanentRedirect)
		return
	}
	log.Printf("in handleLogin, session.Values: %#v\n", session.Values)
	log.Printf("authenticated - uid: %d, username: %s, password: %s\n", uid, username, password)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) handleLogout(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "newsession")
	if err != nil {
		log.Printf("in handleLogout, error while getting session: %v\n", err)
		clearSession(session, w, r)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	clearSession(session, w, r)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func clearSession(s *sessions.Session, w http.ResponseWriter, r *http.Request) {
	log.Println("clearSession just ran")
	s.Options.MaxAge = -1
	err := s.Save(r, w)
	if err != nil {
		log.Printf("in clearSession(), error while saving session: %v\n", err)
	}
}

func getUsername(w http.ResponseWriter, r *http.Request) (string, error) {
	session, err := store.Get(r, "newsession")
	if err != nil {
		log.Printf("in getUsername, error while getting session: %v\n", err)
		clearSession(session, w, r)
		return "", err
	}
	log.Printf("in getUsername, %#v\n", session.Values)
	un, ok := session.Values["username"].(string)
	if !ok {
		log.Printf("in getUsername, un:%v - ok: %v\n", un, ok)
		return "", nil
	}

	return un, nil
}

func (app *application) handleComment(w http.ResponseWriter, r *http.Request) {
	username, err := getUsername(w, r)
	log.Printf("username is: %v\n", username)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var p models.Post
	redirectTo := r.FormValue("goto")
	p.Title = r.FormValue("text")
	parent := r.FormValue("parent")
	pid, err := strconv.Atoi(parent)
	if err != nil {
		log.Printf("in handleComment, error while converting parent id: %s - %v\n", parent, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	p.ParentID = pid
	p.Owner = username
	log.Printf("form values: %#v\n", r.Form)
	log.Printf("comment: %#v\n", p)

	id, err := app.db.CreatePost(&p)
	if err != nil {
		log.Printf("in handleComment, error while inserting comment: %#v - %v\n", p, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	log.Printf("comment successfully inserted with the id: %d\n", id)
	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
}

func (app *application) handleUser(w http.ResponseWriter, r *http.Request) {
	uname, err := getUsername(w, r)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	vars := mux.Vars(r)
	username := vars["username"]
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	query := "SELECT * FROM users WHERE user_name = ?"
	u, err := app.db.GetUser(query, username)
	if err != nil {
		log.Printf("in handleUser, error while getting user: %v\n", err)
	}
	log.Printf("%#v\n", u)
	data := TmplData{User: u, Username: uname}
	app.tmpl.ExecuteTemplate(w, "user.html", data)
}

func (app *application) handleReply(w http.ResponseWriter, r *http.Request) {
	// postid:=28360987, commentid=28361157,
	vars := mux.Vars(r)
	log.Printf("vars: %v\n", vars)

	id := vars["id"]
	log.Printf("id: %v\n", id)
	queryPost := "SELECT post_id, link, title, domain, owner, points, parent_id, created_at FROM posts WHERE post_id = ?"
	p, err := app.db.GetPost(queryPost, id)
	if err != nil {
		log.Printf("in handleReply, error while getting comment: %v\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("comment: %#v\n", p)
	data := &TmplData{
		Post: p,
	}

	app.tmpl.ExecuteTemplate(w, "addcomment.html", data)
}

func getDomain(link string) string {
	u, err := url.Parse(link)
	if err != nil {
		return link
	}

	return u.Hostname()
}
