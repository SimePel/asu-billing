package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
)

func userCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var user *User
		if userID := chi.URLParam(r, "userID"); userID != "" {
			user = dbGetUser(userID)
		} else {
			http.Error(w, http.StatusText(404), 404)
			return
		}

		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*User)
	json.NewEncoder(w).Encode(user)
}
