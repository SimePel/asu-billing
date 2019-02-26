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

var (
	usrT = template.Must(template.New("usr").ParseGlob("templates/usr/*.html"))
)

type Tariff struct {
	ID    int    `bson:"id"`
	Price int    `bson:"price"`
	Name  string `bson:"name"`
}

// User is an instance of users collection from mongodb
type User struct {
	ID           int       `bson:"_id"`
	Money        int       `bson:"money"`
	Active       bool      `bson:"active"`
	Name         string    `bson:"name"`
	Login        string    `bson:"login"`
	InIP         string    `bson:"in_ip"`
	ExtIP        string    `bson:"ext_ip"`
	Tariff       Tariff    `bson:"tariff"`
	Phone        string    `bson:"phone,omitempty"`
	Comment      string    `bson:"comment,omitempty"`
	PaymentsEnds time.Time `bson:"payments_ends,omitempty"`
}

// CorrectedUser needs to print appropriate information about user
type CorrectedUser struct {
	ID           int
	Money        int
	Active       bool
	Name         string
	Login        string
	Tariff       Tariff
	InIP         string
	ExtIP        string
	Phone        string
	Comment      string
	PaymentsEnds string
}

func userIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := store.Get(r, "user")
	if session.Values["user_logged"] == "false" || session.Values["user_logged"] == nil {
		http.Redirect(w, r, "/user-login", http.StatusFound)
		return
	}

	flashes := session.Flashes()
	usrT.ExecuteTemplate(w, "index", getUserDataByLogin(flashes[len(flashes)-1].(string)))
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
	searchRequest := ldap.NewSearchRequest(
		fmt.Sprintf("dc=%s,dc=asu,dc=ru", pieces[0]),
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(samAccountName=%s)", pieces[1]),
		[]string{"dn"},
		nil,
	)

	err := ldapAuth(w, r, searchRequest)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/user-login", http.StatusFound)
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
