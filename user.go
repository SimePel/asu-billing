package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	ldap "gopkg.in/ldap.v3"
)

func newUserTemplate() *template.Template {
	funcMap := template.FuncMap{
		"formatTime": formatTime,
	}

	return template.Must(template.New("usr").Funcs(funcMap).ParseGlob("templates/usr/*.html"))
}

var (
	usrT = newUserTemplate()
)

// Payment type
type Payment struct {
	Amount int
	Last   time.Time
}

// Tariff type
type Tariff struct {
	ID    int
	Price int
	Name  string
}

// User is a table in mysql
type User struct {
	ID              int
	Money           int
	Active          bool
	Agreement       string
	Name            string
	Login           string
	InIP            string
	ExtIP           string
	Tariff          Tariff
	Payments        []Payment
	ConnectionPlace string
	Phone           string
	Comment         string
	PaymentsEnds    time.Time
}

func userIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := store.Get(r, "user")
	id := session.Values["user_id"].(int)
	user, err := getUserByID(id)
	if err != nil {
		log.Printf("could not get user by id=%v: %v", id, err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	usrT.ExecuteTemplate(w, "index", user)
}

func userLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := store.Get(r, "user")
	if session.Values["user_logged"] == "true" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	usrT.ExecuteTemplate(w, "login", nil)
}

func authUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := store.Get(r, "user")
	if session.Values["user_logged"] == "true" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	login := r.FormValue("login")
	searchRequest := ldap.NewSearchRequest(
		"dc=stud,dc=asu,dc=ru",
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(samAccountName=%s)", login),
		[]string{"dn"},
		nil,
	)

	err := ldapAuth(w, r, searchRequest)
	if err != nil {
		log.Println(err)
		url := fmt.Sprint("/user-login?err=1")
		http.Redirect(w, r, url, http.StatusFound)
		return
	}

	login += "@stud.asu.ru"

	if !userExistInDB(login) {
		url := fmt.Sprint("/user-login?err=2")
		http.Redirect(w, r, url, http.StatusFound)
		return
	}

	id, err := getUserIDByLogin(login)
	if err != nil {
		log.Println("Cannot get user id by login.", err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	session.Values["user_logged"] = "true"
	session.Values["user_id"] = id
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

// Settings that user can modify on its page
type Settings struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

func userSettings(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := store.Get(r, "user")
	id := session.Values["user_id"].(int)
	settings, err := getUserSettingsByID(id)
	if err != nil {
		log.Println("Cannot get user settings by id.", err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	usrT.ExecuteTemplate(w, "settings", settings)
}

func editUserSettings(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Cannot read body.", err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	var Email struct {
		V string `json:"email"`
	}

	err = json.Unmarshal(body, &Email)
	if err != nil {
		log.Println("Cannot unmarshal json.", err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	session, _ := store.Get(r, "user")
	settings := Settings{
		ID:    session.Values["user_id"].(int),
		Email: Email.V,
	}

	url, err := createConfirmationLink(settings)
	if err != nil {
		log.Println("Cannot create confirmation link.", err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	err = confirmEmail(settings.Email, url)
	if err != nil {
		log.Println("Cannot confirmEmail.", err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}
}

func confirmSettings(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	cipherData := r.FormValue("d")
	data, err := decrypt(cipherData)
	if err != nil {
		log.Println("Cannot decrypt cipher data.", err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	var settings Settings
	err = json.Unmarshal(data, &settings)
	if err != nil {
		log.Println("Cannot unmarshal json.", err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	err = updateUserSettings(settings)
	if err != nil {
		log.Println("Cannot update user settings.", err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	usrT.ExecuteTemplate(w, "confirm-settings", nil)
}

func userLogout(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := store.Get(r, "user")
	session.Values["user_logged"] = "false"
	session.Save(r, w)
	http.Redirect(w, r, "/user-login", http.StatusFound)
}
