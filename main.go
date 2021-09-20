package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/csrf"
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
	csrfMiddleware := csrf.Protect(
		[]byte("very-secret-fornow"),
		csrf.CookieName("token"),
		csrf.FieldName("token"),
	)

	srv := &http.Server{
		Addr:         port,
		Handler:      csrfMiddleware(app.routes()),
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	srvDone := make(chan os.Signal, 1)
	signal.Notify(srvDone, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	log.Printf("started on %s", port)

	<-srvDone

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("error while shutting down:%v", err)
	}

	log.Print("shut down gracefully")
}
