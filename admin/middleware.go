package main

import (
	"context"
	"net/http"
)

func jsonContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		next.ServeHTTP(w, r)
	})
}

func checkJWTtoken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := getJWTtokenFromCookies(r.Cookies())
		if err != nil {
			http.Redirect(w, r, "/login", 303)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type dbCtxKey string

func setDBtoCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db := initializeDB()
		ctx := context.WithValue(r.Context(), dbCtxKey("db"), db)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
