package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi"
)

func newRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Handle("/assets/*", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))
	r.Handle("/records/*", http.StripPrefix("/records/", http.FileServer(http.Dir("./records"))))
	r.Handle("/favicon.ico", http.StripPrefix("/", http.FileServer(http.Dir("./"))))

	r.Get("/login", loginHandler)
	r.With(jsonContentType).Post("/login", loginPostHandler)
	r.With(checkJWTtoken).Post("/add-user", addUserPostHandler)
	r.With(checkJWTtoken).Post("/edit-user", editUserPostHandler)
	r.With(checkJWTtoken).Post("/send-mass-sms", sendMassSMSPostHandler)
	r.With(checkJWTtoken).Post("/income-for-period", getIncomeForPeriodHandler)
	r.With(checkJWTtoken).Post("/generate-payments-report", generatePaymentsReportHandler)
	r.With(checkJWTtoken).With(jsonContentType).Post("/payment", paymentPostHandler)
	r.With(checkJWTtoken).With(jsonContentType).Post("/check-vacant-esockets", areThereSocketInTheRoomHandler)
	r.With(checkJWTtoken).Post("/change-notification-status", changeNotificationStatusHandler)

	r.With(checkJWTtoken).Get("/", indexHandler)
	r.With(checkJWTtoken).Get("/logout", logoutHandler)
	r.With(checkJWTtoken).Get("/add-user", addUserHandler)
	r.With(checkJWTtoken).Get("/edit-user", editUserHandler)
	r.With(checkJWTtoken).Get("/user", userHandler)
	r.With(checkJWTtoken).Get("/notification-status", notificationStatusHandler)
	r.With(checkJWTtoken).Get("/send-mass-sms", sendMassSMSHandler)
	r.With(checkJWTtoken).With(jsonContentType).Get("/stats", getStatsAboutUsers)
	r.With(checkJWTtoken).With(jsonContentType).Get("/next-agreement", getNextAgreementHandler)

	r.Route("/users", func(r chi.Router) {
		r.Use(checkJWTtoken)
		r.Use(jsonContentType)
		r.Use(setDBtoCtx)
		r.Route("/{userID}", func(r chi.Router) {
			r.Use(userCtx)
			r.Get("/", getUser)
			r.Put("/", restoreUser)
			r.Delete("/", archiveUser)
			r.Post("/deactivate", deactivateUser)
			r.Post("/activate", activateUser)
			r.Post("/limit", limitUser)
			r.Post("/unlimit", unlimitUser)
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

func sendMassSMSHandler(w http.ResponseWriter, r *http.Request) {
	b, _ := ioutil.ReadFile("templates/send-mass-sms.html")
	w.Write(b)
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
	name := strings.TrimSpace(r.FormValue("name"))
	isEmployee, err := strconv.ParseBool(r.FormValue("isEmployee"))
	if err != nil {
		log.Println(err)
	}
	agreement := strings.TrimSpace(r.FormValue("agreement"))
	login := strings.TrimSpace(r.FormValue("login")) + "@stud.asu.ru"
	phone := strings.TrimSpace(r.FormValue("phone"))
	room := strings.TrimSpace(r.FormValue("room"))
	comment := strings.TrimSpace(r.FormValue("comment"))
	connectionPlace := strings.TrimSpace(r.FormValue("connectionPlace"))
	agreementConclusionDate, _ := time.Parse("2006-01-02", r.FormValue("agreementConclusionDate"))
	tariff, _ := strconv.Atoi(r.FormValue("tariff"))

	user := User{
		Name:                    name,
		IsEmployee:              isEmployee,
		Agreement:               agreement,
		Login:                   login,
		Tariff:                  Tariff{ID: tariff},
		Phone:                   phone,
		Room:                    room,
		Comment:                 comment,
		ConnectionPlace:         connectionPlace,
		AgreementConclusionDate: agreementConclusionDate,
	}

	mysql := MySQL{db: initializeDB()}
	id, err := mysql.AddUser(user)
	if err != nil {
		log.Printf("cannot add user with id=%v: %v", id, err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	if user.IsEmployee {
		err = mysql.FreePaymentForOneYear(id)
		if err != nil {
			log.Printf("cannot make free payment for one year: %v", err)
		}
	}

	http.Redirect(w, r, "/", 303)
}

func editUserPostHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.FormValue("id"))
	name := strings.TrimSpace(r.FormValue("name"))
	agreement := strings.TrimSpace(r.FormValue("agreement"))
	isEmployee, _ := strconv.ParseBool(r.FormValue("isEmployee"))
	mac := strings.TrimSpace(r.FormValue("mac"))
	login := strings.TrimSpace(r.FormValue("login"))
	phone := strings.TrimSpace(r.FormValue("phone"))
	room := strings.TrimSpace(r.FormValue("room"))
	comment := strings.TrimSpace(r.FormValue("comment"))
	connectionPlace := strings.TrimSpace(r.FormValue("connectionPlace"))
	expiredDate, _ := time.Parse("2006-01-02", r.FormValue("expiredDate"))
	agreementConclusionDate, _ := time.Parse("2006-01-02", r.FormValue("agreementConclusionDate"))
	tariff, _ := strconv.Atoi(r.FormValue("tariff"))

	user := User{
		ID:                      uint(id),
		Name:                    name,
		Agreement:               agreement,
		IsEmployee:              isEmployee,
		Mac:                     mac,
		Login:                   login,
		Tariff:                  Tariff{ID: tariff},
		Phone:                   phone,
		Room:                    room,
		Comment:                 comment,
		ConnectionPlace:         connectionPlace,
		ExpiredDate:             expiredDate,
		AgreementConclusionDate: agreementConclusionDate,
	}

	mysql := MySQL{db: initializeDB()}
	oldUser, err := mysql.GetUserByID(id)
	if err != nil {
		log.Printf("cannot get user: %v", err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	err = mysql.UpdateUser(user)
	if err != nil {
		log.Printf("cannot edit user with id=%v: %v", id, err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	if !oldUser.IsEmployee && user.IsEmployee {
		err = mysql.FreePaymentForOneYear(id)
		if err != nil {
			log.Printf("cannot make free payment for one year: %v", err)
		}
	} else if oldUser.IsEmployee && !user.IsEmployee {
		err = mysql.ResetFreePaymentForOneYear(id)
		if err != nil {
			log.Printf("cannot reset free payment for one year: %v", err)
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/user?id=%v", id), 303)
}

func sendMassSMSPostHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var j struct {
		Message string `json:"message"`
		Phones  string `json:"phones"`
	}

	err := json.NewDecoder(r.Body).Decode(&j)
	if err != nil {
		log.Printf("cannot decode json: %v", err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	err = sendSMS(j.Phones, j.Message)
	if err != nil {
		log.Printf("cannot send mass sms: %v", err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}
}

func paymentPostHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var payment struct {
		UserID  int    `json:"id"`
		Sum     int    `json:"sum"`
		Method  string `json:"method"`
		Receipt string `json:"receipt"`
		Admin   string `json:"admin"`
	}

	token, err := getJWTtokenFromCookies(r.Cookies())
	if err != nil {
		log.Println(err)
		return
	}

	claims := token.Claims.(jwt.MapClaims)
	payment.Admin = claims["login"].(string)

	err = json.NewDecoder(r.Body).Decode(&payment)
	if err != nil {
		log.Println(err)
		return
	}

	mysql := MySQL{db: initializeDB()}
	err = mysql.ProcessPayment(payment.UserID, payment.Sum, payment.Method, payment.Receipt, payment.Admin)
	if err != nil {
		log.Println(err)
		return
	}

	user, err := mysql.GetUserByID(payment.UserID)
	if err != nil {
		log.Println(err)
		return
	}

	if !user.Paid {
		tryToRenewPayment(mysql, int(user.ID))
	}
}

func areThereSocketInTheRoomHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var j struct {
		Room string `json:"room"`
	}

	err := json.NewDecoder(r.Body).Decode(&j)
	if err != nil {
		log.Printf("cannot decode json: %v", err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	mysql := MySQL{db: initializeDB()}
	roomID, err := mysql.getRoomIDByName(j.Room)
	if err != nil {
		log.Printf("cannot get room id by name - %v: %v", j.Room, err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	var es int
	err = mysql.db.QueryRow(`SELECT vacant_esockets FROM rooms WHERE id=?`, roomID).Scan(&es)
	if err != nil {
		log.Printf("cannot get vacant_esockets from room with id - %v: %v", roomID, err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	J := struct {
		Answer bool `json:"answer"`
	}{
		Answer: es > 0,
	}
	json.NewEncoder(w).Encode(&J)
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

	archivedUsersCount, err := mysql.GetCountOfArchivedUsers()
	if err != nil {
		log.Printf("cannot get count of archived users: %v", err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	cash, err := mysql.GetAllMoneyWeHave()
	if err != nil {
		log.Printf("cannot get sum of all money we have: %v", err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	J := struct {
		ActiveUsersCount   int `json:"active_users_count"`
		InactiveUsersCount int `json:"inactive_users_count"`
		ArchivedUsersCount int `json:"archived_users_count"`
		Cash               int `json:"cash"`
	}{
		ActiveUsersCount:   activeUsersCount,
		InactiveUsersCount: inactiveUsersCount,
		ArchivedUsersCount: archivedUsersCount,
		Cash:               cash,
	}
	json.NewEncoder(w).Encode(&J)
}

func getIncomeForPeriodHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var j struct {
		From time.Time `json:"from"`
		To   time.Time `json:"to"`
	}

	err := json.NewDecoder(r.Body).Decode(&j)
	if err != nil {
		log.Printf("cannot decode json: %v", err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	mysql := MySQL{db: initializeDB()}
	income, err := mysql.GetIncomeForPeriod(j.From.Format("20060102"), j.To.Format("20060102"))
	if err != nil {
		log.Printf("cannot get income for period: %v", err)
		http.Error(w, "0", http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, income)
}

func generatePaymentsReportHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	fromDate, _ := time.Parse("2006-01-02", r.FormValue("from"))
	toDate, _ := time.Parse("2006-01-02", r.FormValue("to"))

	mysql := MySQL{db: initializeDB()}
	records, err := mysql.GetPaymentsRecords(fromDate.Format("20060102"), toDate.Format("20060102"))
	if err != nil {
		log.Printf("cannot get payments records for period: %v", err)
		http.Error(w, "0", http.StatusInternalServerError)
		return
	}

	f, err := WriteToFile(records)
	if err != nil {
		log.Printf("cannot create csv file: %v", err)
		http.Error(w, "Что-то пошло не так(", http.StatusInternalServerError)
	}

	fmt.Fprint(w, f.Name())
}

func getNextAgreementHandler(w http.ResponseWriter, r *http.Request) {
	mysql := MySQL{db: initializeDB()}
	agreement, err := mysql.GetNextAgreement()
	if err != nil {
		log.Printf("cannot get next agreement: %v", err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	J := struct {
		Agreement string `json:"agreement"`
	}{
		Agreement: agreement,
	}

	json.NewEncoder(w).Encode(&J)
}
