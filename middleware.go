package main

import (
	"log"
	"net/http"
)

func authenticationMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, "newsession")
		if err != nil {
			log.Printf("in authenticatonMiddleware, error while getting session: %v\n", err)
			clearSession(session, w, r)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		auth, ok := session.Values["authenticated"].(bool)
		if !ok || !auth {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	}
}
