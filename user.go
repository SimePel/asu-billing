package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
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
	pieces := strings.Split(login, "\\")
	if len(pieces) == 1 {
		pieces = []string{"stud", login}
	}
	searchRequest := ldap.NewSearchRequest(
		fmt.Sprintf("dc=%s,dc=asu,dc=ru", pieces[0]),
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(samAccountName=%s)", pieces[1]),
		[]string{"dn"},
		nil,
	)

	err := ldapAuth(w, r, searchRequest)
	if err != nil {
		url := fmt.Sprint("/user-login?err=", err.Error())
		http.Redirect(w, r, url, http.StatusFound)
		return
	}

	session.Values["user_logged"] = "true"
	session.AddFlash(pieces[1] + getRightPostfix(pieces[0]))
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func getRightPostfix(domain string) string {
	if domain == "stud" {
		return "@stud.asu.ru"
	}
	if domain == "mc" {
		return "@mc.asu.ru"
	}
	return "Неизвестный домен"
}

func userLogout(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := store.Get(r, "user")
	session.Values["user_logged"] = "false"
	session.Save(r, w)
	http.Redirect(w, r, "/user-login", http.StatusFound)
}
