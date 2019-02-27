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
	var name string
	t := r.FormValue("type")
	if t == "name" {
		name = r.FormValue("name")
	}
	admT.ExecuteTemplate(w, "index", getUsersByType(t, name))
}

func adminLogout(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := store.Get(r, "admin")
	session.Values["admin_logged"] = "false"
	session.Save(r, w)
	http.Redirect(w, r, "/admin-login", http.StatusFound)
}

func userInfo(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	id, _ := strconv.Atoi(r.FormValue("id"))
	admT.ExecuteTemplate(w, "user-info", getUserDataByID(id))
}

func newUserForm(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	admT.ExecuteTemplate(w, "new-user-form", nil)
}

func addNewUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	name := r.FormValue("name")
	login := r.FormValue("login") + "@stud.asu.ru"
	tariff := r.FormValue("tariff")
	phone := r.FormValue("phone")
	comment := r.FormValue("comment")

	moneyStr := r.FormValue("money")
	money := 0
	if moneyStr != "" {
		money, _ = strconv.Atoi(moneyStr)
	}

	id, err := addUserIntoMongo(name, login, tariff, phone, comment, money)
	if err != nil {
		log.Fatal(err)
	}

	withdrawMoney(id)

	http.Redirect(w, r, "/admin", http.StatusFound)
}

func payForm(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	id, _ := strconv.Atoi(r.FormValue("id"))
	admT.ExecuteTemplate(w, "payment", getUserDataByID(id))
}

func pay(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	id, _ := strconv.Atoi(r.FormValue("id"))
	moneyStr := r.FormValue("money")
	money, _ := strconv.Atoi(moneyStr)

	addMoneyToUser(id, money)

	err := withdrawMoney(id)
	if err != nil {
		log.Fatal(err)
	}
	http.Redirect(w, r, "/admin", http.StatusFound)
}
