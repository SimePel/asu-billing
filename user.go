package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi"
)

type userCtxKey string

func userCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "userID")
		if userID == "" {
			http.Error(w, http.StatusText(404), 404)
			return
		}

		user, err := dbGetUser(userID)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(500), 500)
			return
		}
		ctx := context.WithValue(r.Context(), userCtxKey("user"), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userCtxKey("user")).(*User)
	err := json.NewEncoder(w).Encode(user)
	if err != nil {
		log.Println("cannot encode json. ", err)
		w.Write([]byte("{}"))
	}
}
