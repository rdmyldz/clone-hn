package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/rdmyldz/clone-hn/models"
)

const timeout = 500 * time.Millisecond
const limit = 3

func (app *application) handleHome(w http.ResponseWriter, r *http.Request) {
	app.handleNews(w, r)
}

func pagination(r *http.Request) (int, int, error) {
	page := strings.TrimSpace(r.URL.Query().Get("p"))
	if page == "" {
		page = "1"
	}
	p, err := strconv.Atoi(page)
	if err != nil {
		return -1, p, fmt.Errorf("error while converting %s : %w", page, err)
	}
	return (limit * (p - 1)), p + 1, nil

}

func (app *application) handleNews(w http.ResponseWriter, r *http.Request) {
	log.Println("in handleNews")
	uname, err := getUsername(w, r)
	if err != nil {
		log.Printf("error returned from getusername: %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	offset, page, err := pagination(r)
	if err != nil {
		log.Printf("in handleNews: %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	log.Printf("offset: %d -- page: %d\n", offset, page)
	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	query := fmt.Sprintf(`SELECT post_id, link, title, domain, owner, points, parent_id, main_post_id,
			comment_num, title_summary, created_at
			FROM posts
			WHERE parent_id = 0 AND
			title NOT LIKE "Ask HN:%%" AND
			title NOT LIKE "Show HN:%%"
			LIMIT %d OFFSET %d`, limit, offset)
	posts, err := app.db.GetPosts(ctx, query)
	if err != nil {
		log.Printf("in handleNews, error while getting posts. err: %v\n", err)
		if errors.Is(err, context.Canceled) {
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	data := &TmplData{
		Posts:    posts,
		Username: uname,
		Page:     page,
	}
	app.tmpl.ExecuteTemplate(w, "home.html", data)
}

func (app *application) handleNewest(w http.ResponseWriter, r *http.Request) {
	uname, err := getUsername(w, r)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	offset, page, err := pagination(r)
	if err != nil {
		log.Printf("in handleNews: %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	log.Printf("offset: %d -- page: %d\n", offset, page)

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	query := fmt.Sprintf(`SELECT post_id, link, title, domain, owner, points, parent_id, main_post_id,
		comment_num, title_summary, created_at 
		FROM posts 
		Where parent_id = 0 ORDER BY created_at DESC
		LIMIT %d OFFSET %d`,
		limit, offset)
	posts, err := app.db.GetPosts(ctx, query)
	if err != nil {
		log.Printf("in handleNewest, error while getting posts. err: %v\n", err)
		if errors.Is(err, context.Canceled) {
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	data := &TmplData{
		Posts:    posts,
		Username: uname,
		Page:     page,
	}
	app.tmpl.ExecuteTemplate(w, "newest.html", data)
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

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	queryPost := "SELECT post_id, link, title, domain, owner, points, parent_id, main_post_id, comment_num, title_summary, created_at FROM posts WHERE post_id = ?"
	p, err := app.db.GetPost(ctx, queryPost, id)
	if err != nil {
		log.Printf("in handleItem, error while getting post. err: %v\n", err)
		if errors.Is(err, context.Canceled) {
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// queryComment := "SELECT post_id, link, domain, owner, points, parent_id, created_at FROM posts WHERE parent_id = ?"
	// SELECT * FROM posts WHERE parent_id = 1 OR parent_id IN (SELECT post_id FROM posts WHERE parent_id = 1);
	// query := fmt.Sprintf(`SELECT post_id, link, title, domain, owner, points, parent_id, created_at FROM posts where parent_id = "%d" OR parent_id IN (SELECT post_id FROM posts WHERE parent_id = "%d")`, p.ID, p.ID)
	query := fmt.Sprintf(`
	WITH RECURSIVE temp_posts (post_id, link, title, domain, owner, points, parent_id, main_post_id, comment_num, title_summary, created_at) AS (
    SELECT p.post_id, p.link, p.title, p.domain, p.owner, p.points, p.parent_id, p.main_post_id, p.comment_num, p.title_summary, p.created_at
    FROM posts AS p
    WHERE p.parent_id = %d

    UNION ALL

    SELECT p.post_id, p.link, p.title, p.domain, p.owner, p.points, p.parent_id, p.main_post_id, p.comment_num, p.title_summary, p.created_at
    FROM posts AS p
    JOIN temp_posts tp ON tp.post_id = p.parent_id ORDER BY p.parent_id DESC, p.created_at DESC
)

SELECT * FROM temp_posts;
	`, p.ID)
	comments, err := app.db.GetPosts(ctx, query)
	if err != nil {
		log.Printf("in handleItem, error while getting comments. err: %v\n", err)
		if errors.Is(err, context.Canceled) {
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	indentations := getDepth(comments, p.ID)
	csrfToken := map[string]interface{}{csrf.TemplateTag: csrf.TemplateField(r)}
	data := &TmplData{
		Post:        p,
		Posts:       comments,
		Username:    uname,
		Indentation: indentations,
		CsrfToken:   csrfToken,
	}

	err = app.tmpl.ExecuteTemplate(w, "item.html", data)
	if err != nil {
		log.Printf("error while executing item.html: %v\n", err)
	}
}

func getDepth(data []models.Post, pid int) map[int]int {
	depth := make(map[int]int)
	findDepth(data, pid, 0, depth)
	return depth
}

func findDepth(data []models.Post, pid int, depth int, indMap map[int]int) {
	for i, d := range data {
		if d.ParentID == pid {
			indMap[d.ID] = depth * 40
			findDepth(data[i+1:], d.ID, depth+1, indMap)
		}
	}
}

func (app *application) handleSubmit(w http.ResponseWriter, r *http.Request) {
	_, err := getUsername(w, r)
	if err != nil {
		http.Redirect(w, r, "/submit", http.StatusSeeOther)
		return
	}
	csrfToken := map[string]interface{}{csrf.TemplateTag: csrf.TemplateField(r)}
	app.tmpl.ExecuteTemplate(w, "submit.html", csrfToken)
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
	post.CreatedAt = time.Now().UTC()
	post.Owner = username

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	_, err = app.db.CreatePost(ctx, &post)
	if err != nil {
		log.Printf("in handleR, error while creating post. err: %v\n", err)
		if errors.Is(err, context.Canceled) {
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("in handleR, post created. id: %#v\n", post)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) handleFrom(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	site := vars["site"]

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	query := fmt.Sprintf(`SELECT post_id, link, title, domain, owner, points, parent_id, main_post_id, comment_num, title_summary, created_at FROM posts where domain = "%s"`, site)
	posts, err := app.db.GetPosts(ctx, query)
	if err != nil {
		log.Printf("in handleFrom, error while getting posts. err: %v\n", err)
		if errors.Is(err, context.Canceled) {
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	data := TmplData{
		Posts: posts,
	}

	app.tmpl.ExecuteTemplate(w, "home.html", data)
}

func (app *application) handleLogin(w http.ResponseWriter, r *http.Request) {
	csrfToken := map[string]interface{}{csrf.TemplateTag: csrf.TemplateField(r)}
	if r.Method == "GET" {
		app.tmpl.ExecuteTemplate(w, "login.html", csrfToken)
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
			if errors.Is(err, context.Canceled) {
				return
			}
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
		if errors.Is(err, context.Canceled) {
			return
		}
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
	p.MainPostID, err = getMainPostID(redirectTo)
	if err != nil {
		log.Printf("in handleComment, error while getting main_post_id: %s - %v\n", redirectTo, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	p.Owner = username
	p.CreatedAt = time.Now().UTC()
	log.Printf("form values: %#v\n", r.Form)
	log.Printf("comment: %#v\n", p)

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	_, err = app.db.CreatePost(ctx, &p)
	if err != nil {
		log.Printf("in handleComment, error while inserting comment: %#v - %v\n", p, err)
		if errors.Is(err, context.Canceled) {
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = app.db.UpdateCommentNum(ctx, p.MainPostID)
	if err != nil {
		log.Printf("in handleComment, error while updating comment_num: %#v - %v\n", p, err)
		if errors.Is(err, context.Canceled) {
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	log.Printf("after update - comment: %#v\n", p)
	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
}

func getMainPostID(p string) (int, error) {
	pp := strings.Split(p, "/")[2]
	mpi, err := strconv.Atoi(strings.Split(pp, "#")[0])
	if err != nil {
		return mpi, err
	}
	return mpi, nil
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
		if errors.Is(err, context.Canceled) {
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	log.Printf("%#v\n", u)
	data := TmplData{User: u, Username: uname}
	app.tmpl.ExecuteTemplate(w, "user.html", data)
}

func (app *application) handleReply(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	log.Printf("vars: %v\n", vars)

	id := vars["id"]
	log.Printf("id: %v\n", id)

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	queryPost := "SELECT post_id, link, title, domain, owner, points, parent_id, main_post_id, comment_num, title_summary, created_at FROM posts WHERE post_id = ?"
	p, err := app.db.GetPost(ctx, queryPost, id)
	if err != nil {
		log.Printf("in handleReply, error while getting comment: %v\n", err)
		if errors.Is(err, context.Canceled) {
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	log.Printf("comment: %#v\n", p)
	csrfToken := map[string]interface{}{csrf.TemplateTag: csrf.TemplateField(r)}
	data := &TmplData{
		Post:      p,
		CsrfToken: csrfToken,
	}

	app.tmpl.ExecuteTemplate(w, "addcomment.html", data)
}

func (app *application) handleVote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	log.Printf("vars: %v\n", vars)

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Printf("in handleVote, error while converting id: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	log.Printf("id: %v\n", id)
	redirectTo := r.FormValue("goto")

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	err = app.db.UpdatePoints(ctx, id)
	if err != nil {
		log.Printf("in handleVote, error while updating points: %v\n", err)
		if errors.Is(err, context.Canceled) {
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, redirectTo, http.StatusSeeOther)

}

func getDomain(link string) string {
	u, err := url.Parse(link)
	if err != nil {
		return link
	}

	return u.Hostname()
}

func (app *application) handleNewComments(w http.ResponseWriter, r *http.Request) {
	log.Println("in handleNewComments")
	uname, err := getUsername(w, r)
	if err != nil {
		log.Printf("error returned from getusername: %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	query := `SELECT post_id, link, title, domain, owner, points, parent_id, main_post_id,
			comment_num, title_summary, created_at 
			FROM posts 
			WHERE parent_id != 0 ORDER BY created_at DESC`
	posts, err := app.db.GetPosts(ctx, query)
	if err != nil {
		log.Printf("in handleNewcomments, error while getting posts. err: %v\n", err)
		if errors.Is(err, context.Canceled) {
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	data := &TmplData{
		Posts:    posts,
		Username: uname,
	}
	err = app.tmpl.ExecuteTemplate(w, "newcomments.html", data)
	if err != nil {
		log.Printf("error while executing newcomments.html: %v\n", err)
	}
}

func (app *application) handleAsk(w http.ResponseWriter, r *http.Request) {
	log.Println("in handleAsk")
	uname, err := getUsername(w, r)
	if err != nil {
		log.Printf("error returned from getusername: %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	offset, page, err := pagination(r)
	if err != nil {
		log.Printf("in handleAsk: %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	log.Printf("offset: %d -- page: %d\n", offset, page)

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	query := fmt.Sprintf(`SELECT post_id, link, title, domain, owner, points, parent_id, main_post_id, 
			comment_num, title_summary, created_at 
			FROM posts 
			WHERE parent_id = 0 AND title LIKE "Ask HN:%%"
			LIMIT %d OFFSET %d`, limit, offset)
	posts, err := app.db.GetPosts(ctx, query)
	if err != nil {
		log.Printf("in handleAsk, error while getting posts. err: %v\n", err)
		if errors.Is(err, context.Canceled) {
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	data := &TmplData{
		Posts:    posts,
		Username: uname,
		Page:     page,
	}
	app.tmpl.ExecuteTemplate(w, "ask.html", data)
}

func (app *application) handleShow(w http.ResponseWriter, r *http.Request) {
	log.Println("in handleShow")
	uname, err := getUsername(w, r)
	if err != nil {
		log.Printf("error returned from getusername: %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	offset, page, err := pagination(r)
	if err != nil {
		log.Printf("in handleShow: %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	log.Printf("offset: %d -- page: %d\n", offset, page)

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	query := fmt.Sprintf(`SELECT post_id, link, title, domain, owner, points, parent_id, main_post_id, 
			comment_num, title_summary, created_at 
			FROM posts 
			WHERE parent_id = 0 AND title LIKE "Show HN:%%"
			LIMIT %d OFFSET %d`, limit, offset)
	posts, err := app.db.GetPosts(ctx, query)
	if err != nil {
		log.Printf("in handleShow, error while getting posts. err: %v\n", err)
		if errors.Is(err, context.Canceled) {
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	data := &TmplData{
		Posts:    posts,
		Username: uname,
		Page:     page,
	}
	app.tmpl.ExecuteTemplate(w, "show.html", data)
}

func (app *application) handleJobs(w http.ResponseWriter, r *http.Request) {
	log.Println("in handleJobs")
	uname, err := getUsername(w, r)
	if err != nil {
		log.Printf("error returned from getusername: %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := &TmplData{
		Username: uname,
	}
	app.tmpl.ExecuteTemplate(w, "jobs.html", data)
}

func (app *application) handleFront(w http.ResponseWriter, r *http.Request) {
	log.Println("in handleFront")
	uname, err := getUsername(w, r)
	if err != nil {
		log.Printf("error returned from getusername: %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	offset, page, err := pagination(r)
	if err != nil {
		log.Printf("in handleNews: %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	log.Printf("offset: %d -- page: %d\n", offset, page)

	day := getDate(time.Now().UTC().AddDate(0, 0, -1))
	val, ok := r.URL.Query()["day"]
	if ok {
		day = val[0]
	}

	dates, err := getDatesForFrontPage(day)
	if err != nil {
		log.Printf("in handleFront, error while getting dates: %v\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	fmt.Printf("in handleFront, day: %q\n", day)

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	query := fmt.Sprintf(`SELECT post_id, link, title, domain, owner, points, parent_id, main_post_id,
		comment_num, title_summary, created_at
		FROM posts
		WHERE parent_id = 0 AND
		%q >= date(created_at) AND
		title NOT LIKE "Ask HN:%%" AND
		title NOT LIKE "Show HN:%%"
		ORDER BY created_at DESC
		LIMIT %d OFFSET %d`, day, limit, offset)
	posts, err := app.db.GetPosts(ctx, query)
	if err != nil {
		log.Printf("in handleFront, error while getting posts. err: %v\n", err)
		if errors.Is(err, context.Canceled) {
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	data := &TmplData{
		Username:   uname,
		Posts:      posts,
		FrontDates: dates,
		Page:       page,
	}
	err = app.tmpl.ExecuteTemplate(w, "front.html", data)
	if err != nil {
		log.Printf("err while executing template: %v\n", err)
	}
}
