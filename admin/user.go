package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

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

		id, _ := strconv.Atoi(userID)
		user, err := GetUserByID(id)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(500), 500)
			return
		}
		ctx := context.WithValue(r.Context(), userCtxKey("user"), &user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUser(w http.ResponseWriter, r *http.Request) {
	// тут нужно проверить, что в контексте юзер ест, а если нет, то упасть красиво
	user := r.Context().Value(userCtxKey("user")).(*User)
	err := json.NewEncoder(w).Encode(user)
	if err != nil {
		log.Println("cannot encode json. ", err)
		w.Write([]byte("{}"))
	}
}