package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
)

func newRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Handle("/assets/*", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))

	r.Get("/login", loginHandler)
	r.With(jsonContentType).Post("/login", loginPostHandler)
	r.With(checkJWTtoken).With(jsonContentType).Post("/payment", paymentPostHandler)
	r.With(checkJWTtoken).Get("/", indexHandler)
	r.With(checkJWTtoken).Get("/logout", logoutHandler)
	r.With(checkJWTtoken).Get("/add-user", addUserHandler)
	r.With(checkJWTtoken).Get("/user", userHandler)
	r.With(checkJWTtoken).With(jsonContentType).Post("/add-user", addUserPostHandler)

	r.Route("/users", func(r chi.Router) {
		r.Use(checkJWTtoken)
		r.Use(jsonContentType)
		r.Use(setDBtoCtx)
		r.Route("/{userID}", func(r chi.Router) {
			r.Use(userCtx)
			r.Get("/", getUser)
		})
		r.Get("/", getAllUsers)
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

func addUserHandler(w http.ResponseWriter, r *http.Request) {
	b, _ := ioutil.ReadFile("templates/add-user.html")
	w.Write(b)
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	b, _ := ioutil.ReadFile("templates/user.html")
	w.Write(b)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	c := http.Cookie{
		Name:    "jwt",
		Expires: time.Now().Add(-1 * time.Minute),
	}
	http.SetCookie(w, &c)
	http.Redirect(w, r, "/login", 303)
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

func addUserPostHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	agreement := r.FormValue("agreement")
	login := r.FormValue("login")
	phone := r.FormValue("phone")
	room := r.FormValue("room")
	connectionPlace := r.FormValue("connectionPlace")
	tariff, _ := strconv.Atoi(r.FormValue("tariff"))

	user := User{
		Name:            name,
		Agreement:       agreement,
		Login:           login,
		Tariff:          Tariff{ID: tariff},
		Phone:           phone,
		Room:            room,
		ConnectionPlace: connectionPlace,
	}

	mysql := MySQL{db: initializeDB()}
	id, err := mysql.AddUser(user)
	if err != nil {
		log.Printf("could not add user to db with id=%v: %v", id, err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", 303)
}

func paymentPostHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var payment struct {
		UserID int `json:"id"`
		Sum    int `json:"sum"`
	}

	err := json.NewDecoder(r.Body).Decode(&payment)
	if err != nil {
		log.Println(err)
		return
	}

	mysql := MySQL{db: initializeDB()}
	err = mysql.ProcessPayment(payment.UserID, payment.Sum)
	if err != nil {
		log.Println(err)
		return
	}

	user, err := mysql.GetUserByID(payment.UserID)
	if err != nil {
		log.Println(err)
		return
	}

	if user.Activity == false {
		tryToRenewPayment(mysql, user)
	}
}

func createTryToRenewPaymentFunc(mysql MySQL, u User) func() {
	user, _ := mysql.GetUserByID(int(u.ID))
	return func() {
		tryToRenewPayment(mysql, user)
	}
}

func tryToRenewPayment(mysql MySQL, user User) {
	if user.hasEnoughMoneyForPayment() {
		expirationDate, err := mysql.PayForNextMonth(user)
		fmt.Printf("%v\n", expirationDate.String())
		if err != nil {
			log.Println(err)
			return
		}

		f := createTryToRenewPaymentFunc(mysql, user)
		time.AfterFunc(expirationDate, f)
	}
}
