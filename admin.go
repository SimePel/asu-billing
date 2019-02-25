package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	ldap "gopkg.in/ldap.v3"
)

const (
	ldapServer = "ads.mc.asu.ru:3268"
)

var (
	admT = template.Must(template.New("adm").ParseGlob("templates/adm/*.html"))
)

func adminLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := store.Get(r, "admin")
	if session.Values["admin_logged"] == "true" {
		http.Redirect(w, r, "/admin", http.StatusFound)
		return
	}
	admT.ExecuteTemplate(w, "login", nil)
}

func authAdmin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := store.Get(r, "admin")
	if session.Values["admin_logged"] == "true" {
		http.Redirect(w, r, "/admin", http.StatusFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Println("form parsing: ", err)
		http.Error(w, "Problems with fetching your data from the form. Please try again", http.StatusInternalServerError)
		return
	}

	login := r.FormValue("login")
	searchRequest := ldap.NewSearchRequest(
		"dc=mc,dc=asu,dc=ru",
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(memberOf=cn=billing,ou=groups,ou=vc,dc=mc,dc=asu,dc=ru)(samAccountName=%s))", login),
		[]string{"dn"},
		nil,
	)

	err := ldapAuth(w, r, searchRequest)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/admin-login", http.StatusFound)
		return
	}

	session.Values["admin_logged"] = "true"
	session.Save(r, w)
	http.Redirect(w, r, "/admin", http.StatusFound)
}

func adminIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := store.Get(r, "admin")
	if session.Values["admin_logged"] == "false" || session.Values["admin_logged"] == nil {
		http.Redirect(w, r, "/admin-login", http.StatusFound)
		return
	}

	t := r.URL.Query().Get("type")
	admT.ExecuteTemplate(w, "index", getUsersByType(t))
}

func adminLogout(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := store.Get(r, "admin")
	session.Values["admin_logged"] = "false"
	session.Save(r, w)
	http.Redirect(w, r, "/admin-login", http.StatusFound)
}

func userInfo(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := store.Get(r, "admin")
	if session.Values["admin_logged"] == "false" || session.Values["admin_logged"] == nil {
		http.Redirect(w, r, "/admin-login", http.StatusFound)
		return
	}

	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	admT.ExecuteTemplate(w, "user-info", getUserDataByID(id))
}

func newUserForm(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := store.Get(r, "admin")
	if session.Values["admin_logged"] == "true" {
		admT.ExecuteTemplate(w, "new-user-form", nil)
		return
	}
	http.Redirect(w, r, "/admin-login", http.StatusFound)
}

func addNewUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := store.Get(r, "admin")
	if session.Values["admin_logged"] == "false" || session.Values["admin_logged"] == nil {
		http.Redirect(w, r, "/admin-login", http.StatusFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Println("form parsing: ", err)
		http.Error(w, "Problems with fetching your data from the form. Please try again", http.StatusInternalServerError)
		return
	}

	name := r.FormValue("name")
	login := r.FormValue("login") + "@stud.asu.ru"
	moneyStr := r.FormValue("money")
	tariffStr := r.FormValue("tariff")
	tariff, _ := strconv.Atoi(tariffStr)
	money := 0
	if moneyStr != "" {
		money, _ = strconv.Atoi(moneyStr)
	}

	err := addUserIntoMongo(name, login, tariff, money)
	if err != nil {
		log.Fatal(err)
	}

	http.Redirect(w, r, "/admin", http.StatusFound)
}

func payForm(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := store.Get(r, "admin")
	if session.Values["admin_logged"] == "false" || session.Values["admin_logged"] == nil {
		http.Redirect(w, r, "/admin-login", http.StatusFound)
		return
	}

	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	admT.ExecuteTemplate(w, "payment", getUserDataByID(id))
}

func pay(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := store.Get(r, "admin")
	if session.Values["admin_logged"] == "false" || session.Values["admin_logged"] == nil {
		http.Redirect(w, r, "/admin-login", http.StatusFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Println("form parsing: ", err)
		http.Error(w, "Problems with fetching your data from the form. Please try again", http.StatusInternalServerError)
		return
	}

	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	moneyStr := r.FormValue("money")
	money, _ := strconv.Atoi(moneyStr)
	addMoneyToUser(id, money)

	http.Redirect(w, r, "/admin", http.StatusFound)
}
