package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rdmyldz/clone-hn/models/sqlitedb"
)

type application struct {
	db   *sqlitedb.SqliteHN
	tmpl *template.Template
}

var store = sessions.NewCookieStore([]byte("very-secret"))

func main() {
	store.Options.HttpOnly = true
	// store.Options.Secure = true
	port := ":8080"

	db, err := sqlitedb.NewDB("database/hn.db")
	if err != nil {
		log.Fatalln("error while opening the database. err: ", err)
	}

	tmpl, err := template.New("hn-tmpl").Funcs(tmplFunc).ParseGlob("ui/html/*.html")
	if err != nil {
		log.Fatalf("error while parsing templates: %v\n", err)
	}

	app := &application{db: db, tmpl: tmpl}

	srv := &http.Server{
		Addr:    port,
		Handler: app.routes(),
	}

	log.Printf("started on %s", port)
	log.Fatal(srv.ListenAndServe())
}
