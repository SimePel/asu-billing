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
	r.With(checkJWTtoken).Post("/add-user", addUserPostHandler)
	r.With(checkJWTtoken).Post("/edit-user", editUserPostHandler)
	r.With(checkJWTtoken).With(jsonContentType).Post("/payment", paymentPostHandler)

	r.With(checkJWTtoken).Get("/", indexHandler)
	r.With(checkJWTtoken).Get("/logout", logoutHandler)
	r.With(checkJWTtoken).Get("/add-user", addUserHandler)
	r.With(checkJWTtoken).Get("/edit-user", editUserHandler)
	r.With(checkJWTtoken).Get("/user", userHandler)
	r.With(checkJWTtoken).Get("/notification-status", notificationStatusHandler)
	r.With(checkJWTtoken).Post("/change-notification-status", changeNotificationStatusHandler)
	r.With(checkJWTtoken).With(jsonContentType).Get("/stats", getStatsAboutUsers)

	r.Route("/users", func(r chi.Router) {
		r.Use(checkJWTtoken)
		r.Use(jsonContentType)
		r.Use(setDBtoCtx)
		r.Route("/{userID}", func(r chi.Router) {
			r.Use(userCtx)
			r.Get("/", getUser)
			r.Delete("/", archiveUser)
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
	_, err := getJWTtokenFromCookies(r.Cookies())
	if err == nil {
		http.Redirect(w, r, "/", 303)
		return
	}
	b, _ := ioutil.ReadFile("templates/login.html")
	w.Write(b)
}

func addUserHandler(w http.ResponseWriter, r *http.Request) {
	b, _ := ioutil.ReadFile("templates/add-user.html")
	w.Write(b)
}

func editUserHandler(w http.ResponseWriter, r *http.Request) {
	b, _ := ioutil.ReadFile("templates/edit-user.html")
	w.Write(b)
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	b, _ := ioutil.ReadFile("templates/user.html")
	w.Write(b)
}

func notificationStatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(strconv.FormatBool(smsNotificationStatus)))
}

func changeNotificationStatusHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	status, err := strconv.ParseBool(string(body))
	if err != nil {
		log.Println(err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	if status == smsNotificationStatus {
		return
	}

	smsNotificationStatus = !smsNotificationStatus
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
		Secure:   true,
		Expires:  time.Now().AddDate(0, 1, 0),
		SameSite: 3,
	}
	http.SetCookie(w, &c)
	J.Answer = "ok"
	json.NewEncoder(w).Encode(J)
}

func addUserPostHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	login := r.FormValue("login")
	phone := r.FormValue("phone")
	room := r.FormValue("room")
	comment := r.FormValue("comment")
	connectionPlace := r.FormValue("connectionPlace")
	tariff, _ := strconv.Atoi(r.FormValue("tariff"))

	user := User{
		Name:            name,
		Login:           login,
		Tariff:          Tariff{ID: tariff},
		Phone:           phone,
		Room:            room,
		Comment:         comment,
		ConnectionPlace: connectionPlace,
	}

	mysql := MySQL{db: initializeDB()}
	id, err := mysql.AddUser(user)
	if err != nil {
		log.Printf("cannot add user with id=%v: %v", id, err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", 303)
}

func editUserPostHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.FormValue("id"))
	name := r.FormValue("name")
	agreement := r.FormValue("agreement")
	login := r.FormValue("login")
	phone := r.FormValue("phone")
	room := r.FormValue("room")
	comment := r.FormValue("comment")
	connectionPlace := r.FormValue("connectionPlace")
	tariff, _ := strconv.Atoi(r.FormValue("tariff"))

	user := User{
		ID:              uint(id),
		Name:            name,
		Agreement:       agreement,
		Login:           login,
		Tariff:          Tariff{ID: tariff},
		Phone:           phone,
		Room:            room,
		Comment:         comment,
		ConnectionPlace: connectionPlace,
	}

	mysql := MySQL{db: initializeDB()}
	err := mysql.UpdateUser(user)
	if err != nil {
		log.Printf("cannot edit user with id=%v: %v", id, err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/user?id=%v", id), 303)
}

func paymentPostHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var payment struct {
		UserID    int `json:"id"`
		ReceiptID int `json:"receipt_id"`
		Sum       int `json:"sum"`
	}

	err := json.NewDecoder(r.Body).Decode(&payment)
	if err != nil {
		log.Println(err)
		return
	}

	mysql := MySQL{db: initializeDB()}
	err = mysql.ProcessPayment(payment.UserID, payment.ReceiptID, payment.Sum)
	if err != nil {
		log.Println(err)
		return
	}

	user, err := mysql.GetUserByID(payment.UserID)
	if err != nil {
		log.Println(err)
		return
	}

	if user.Paid == false {
		tryToRenewPayment(mysql, user)
	}
}

func createSendNotificationFunc(u User) func() {
	return func() {
		err := sendNotification(u)
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func sendNotification(user User) error {
	message := fmt.Sprintf("На ЛС: %v %vр. Пополните счет за проводное подключение к сети АГУ", user.Agreement, user.Balance)
	err := sendSMS(user.Phone, message)
	if err != nil {
		return fmt.Errorf("cannot send sms: %v", err)
	}

	return nil
}

func createTryToRenewPaymentFunc(mysql MySQL, u User) func() {
	return func() {
		tryToRenewPayment(mysql, u)
	}
}

func tryToRenewPayment(mysql MySQL, user User) {
	if user.hasEnoughMoneyForPayment() {
		expirationDate, err := mysql.PayForNextMonth(user)
		if err != nil {
			log.Println(err)
			return
		}

		user, err = mysql.GetUserByID(int(user.ID))
		if err != nil {
			log.Println(err)
			return
		}

		paymentFunc := createTryToRenewPaymentFunc(mysql, user)
		time.AfterFunc(time.Until(expirationDate), paymentFunc)

		notificationDate := expirationDate.AddDate(0, 0, -3)
		notificationFunc := createSendNotificationFunc(user)
		time.AfterFunc(time.Until(notificationDate), notificationFunc)
	}
}

func getStatsAboutUsers(w http.ResponseWriter, r *http.Request) {
	mysql := MySQL{db: initializeDB()}
	activeUsersCount, err := mysql.GetCountOfActiveUsers()
	if err != nil {
		log.Printf("cannot get count of active users: %v", err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	inactiveUsersCount, err := mysql.GetCountOfInactiveUsers()
	if err != nil {
		log.Printf("cannot get count of inactive users: %v", err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	allMoney, err := mysql.GetAllMoneyWeHave()
	if err != nil {
		log.Printf("cannot get sum of all money we have: %v", err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	J := struct {
		ActiveUsersCount   int `json:"active_users_count"`
		InactiveUsersCount int `json:"inactive_users_count"`
		AllMoney           int `json:"all_money"`
	}{
		ActiveUsersCount:   activeUsersCount,
		InactiveUsersCount: inactiveUsersCount,
		AllMoney:           allMoney,
	}
	json.NewEncoder(w).Encode(&J)
}