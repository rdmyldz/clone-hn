package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (app *application) routes() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/", app.handleHome).Methods("GET")
	router.HandleFunc("/news", app.handleNews).Methods("GET")
	router.HandleFunc("/newest", app.handleNewest).Methods("GET")
	router.HandleFunc("/from/{site}", app.handleFrom)
	router.HandleFunc("/item/{id}", app.handleItem).Methods("GET")
	router.HandleFunc("/submit", authenticationMiddleware(app.handleSubmit)).Methods("GET")
	router.HandleFunc("/r", app.handleR).Methods("POST")
	router.HandleFunc("/login", app.handleLogin).Methods("GET", "POST")
	router.HandleFunc("/logout", app.handleLogout).Methods("GET")
	router.HandleFunc("/comment", authenticationMiddleware(app.handleComment)).Methods("POST")
	router.HandleFunc("/user/{username}", app.handleUser).Methods("GET")
	router.HandleFunc("/reply/{id}", authenticationMiddleware(app.handleReply)).Methods("GET")
	router.HandleFunc("/vote/{id}", authenticationMiddleware(app.handleVote)).Methods("GET")
	router.HandleFunc("/newcomments", app.handleNewComments).Methods("GET")
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./ui/static"))))

	return router
}
