package main

import (
	"fmt"
	"html/template"
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

// User is document in "users" mongodb collection
type User struct {
	ID           int
	Money        int
	Active       bool
	Agreement    string
	Name         string
	Login        string
	InIP         string
	ExtIP        string
	Tariff       Tariff
	Payments     []Payment
	Phone        string
	Comment      string
	PaymentsEnds time.Time
}

func userIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := store.Get(r, "user")
	if session.Values["user_logged"] == "false" || session.Values["user_logged"] == nil {
		http.Redirect(w, r, "/user-login", http.StatusFound)
		return
	}

	flashes := session.Flashes()
	login := flashes[len(flashes)-1].(string)
	user, err := getUserByLogin(login)
	if err != nil {
		log.Printf("could not get user by login=%v: %v", login, err)
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

	session.Values["user_logged"] = "true"
	session.AddFlash(login)
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func userLogout(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := store.Get(r, "user")
	session.Values["user_logged"] = "false"
	session.Save(r, w)
	http.Redirect(w, r, "/user-login", http.StatusFound)
}
