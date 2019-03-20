package main

import (
	"fmt"
	"html/template"
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
	Amount int       `bson:"amount"`
	Last   time.Time `bson:"last"`
}

// Tariff type
type Tariff struct {
	ID    int    `bson:"id"`
	Price int    `bson:"price"`
	Name  string `bson:"name"`
}

// User is document in "users" mongodb collection
type User struct {
	ID           int       `bson:"_id"`
	Money        int       `bson:"money"`
	Active       bool      `bson:"active"`
	Name         string    `bson:"name"`
	Login        string    `bson:"login"`
	InIP         string    `bson:"in_ip"`
	ExtIP        string    `bson:"ext_ip"`
	Tariff       Tariff    `bson:"tariff"`
	Payments     []Payment `bson:"payments,omitempty"`
	Phone        string    `bson:"phone,omitempty"`
	Comment      string    `bson:"comment,omitempty"`
	PaymentsEnds time.Time `bson:"payments_ends,omitempty"`
}

func userIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := store.Get(r, "user")
	if session.Values["user_logged"] == "false" || session.Values["user_logged"] == nil {
		http.Redirect(w, r, "/user-login", http.StatusFound)
		return
	}

	flashes := session.Flashes()
	usrT.ExecuteTemplate(w, "index", getUserByLogin(flashes[len(flashes)-1].(string)))
}

func userLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := store.Get(r, "user")
	if session.Values["user_logged"] == "true" {
		http.Redirect(w, r, "/user", http.StatusFound)
		return
	}
	usrT.ExecuteTemplate(w, "login", nil)
}

func authUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := store.Get(r, "user")
	if session.Values["user_logged"] == "true" {
		http.Redirect(w, r, "/user", http.StatusFound)
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
	http.Redirect(w, r, "/user", http.StatusFound)
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
