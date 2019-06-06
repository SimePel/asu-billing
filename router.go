package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
)

func newRouter() *chi.Mux {
	r := chi.NewRouter()

	r.With(jwtToken).Get("/", indexHandler)
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
		log.Println(err)
		J.Answer = "bad"
		switch err.(type) {
		case *loginLDAPerror:
			J.Error = "Неверный логин или пароль."
		default:
			J.Error = "Проблемы на стороне сервера. Повторите попытку через несколько минут."
		}
		json.NewEncoder(w).Encode(J)
		return
	}

	id, err := dbGetIDbyLogin(L.Login)
	if err != nil {
		log.Println(err)
		J.Answer = "bad"
		J.Error = "Проблемы с базой данных."
		json.NewEncoder(w).Encode(J)
		return
	}

	token, err := createJWTtoken(id)
	if err != nil {
		log.Println(err)
		J.Answer = "bad"
		J.Error = "Проблемы с jwt токеном."
		json.NewEncoder(w).Encode(J)
		return
	}
	c := http.Cookie{
		Name:     "jwt",
		Value:    token,
		HttpOnly: true,
		Expires:  time.Now().AddDate(0, 1, 0),
		SameSite: 3,
	}
	http.SetCookie(w, &c)
	J.Answer = "ok"
	json.NewEncoder(w).Encode(J)
}
