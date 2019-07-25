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

	r.With(checkJWTtoken).Get("/", indexHandler)
	r.Get("/login", loginHandler)
	r.With(jsonContentType).Post("/login", loginPostHandler)

	r.Handle("/assets/*", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))

	r.Route("/users", func(r chi.Router) {
		r.Route("/{userID}", func(r chi.Router) {
			r.Use(jsonContentType)
			r.Use(setDBtoCtx)
			r.Use(userCtx)
			r.Get("/", getUser)
		})
	})

	return r
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	b, _ := ioutil.ReadFile("templates/index.html")
	w.Write(b)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	b, _ := ioutil.ReadFile("templates/login.html")
	w.Write(b)
}

func loginPostHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var Auth struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	var J struct {
		Answer string `json:"answer"`
		Error  string `json:"error,omitempty"`
	}

	err := json.NewDecoder(r.Body).Decode(&Auth)
	if err != nil {
		log.Println(err)
		J.Answer = "bad"
		J.Error = "Ошибка парсинга json."
		json.NewEncoder(w).Encode(J)
		return
	}

	err = ldapAuth(Auth.Login, Auth.Password)
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

	token, err := createJWTtoken(Auth.Login)
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
		HttpOnly: false, // for js interaction
		Expires:  time.Now().AddDate(0, 1, 0),
		SameSite: 3,
	}
	http.SetCookie(w, &c)
	J.Answer = "ok"
	json.NewEncoder(w).Encode(J)
}
