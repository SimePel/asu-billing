package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/go-chi/chi"
)

func newRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Get("/", indexHandler)
	r.Get("/login", loginHandler)
	r.With(jsonContentType).Post("/login", loginPostHandler)

	r.Handle("/assets/*", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))

	r.Route("/users", func(r chi.Router) {
		r.Route("/{userID}", func(r chi.Router) {
			r.Use(jsonContentType)
			r.Use(userCtx)
			r.Get("/", getUser)
		})
	})

	return r
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	b, _ := ioutil.ReadFile("templates/usr/index.html")
	w.Write(b)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	b, _ := ioutil.ReadFile("templates/usr/login.html")
	w.Write(b)
}

func loginPostHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var L struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&L)

	var J struct {
		Answer string `json:"answer"`
		Error  string `json:"error,omitempty"`
	}
	err := ldapAuth(L.Login, L.Password)
	if err != nil {
		J.Answer = "bad"
		switch err.(type) {
		case *loginLDAPerror:
			J.Error = "Неверный логин или пароль."
		default:
			J.Error = "Проблемы на стороне сервера. Попробуйте через несколько минут."
		}
		json.NewEncoder(w).Encode(J)
		return
	}

	J.Answer = "ok"
	json.NewEncoder(w).Encode(J)
}
