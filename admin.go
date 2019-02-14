package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	"gopkg.in/ldap.v3"
)

const (
	ldapServer = "ads.mc.asu.ru:3268"
)

var (
	admT = template.Must(template.New("adm").ParseGlob("templates/adm/*.html"))
)

func adminLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := store.Get(r, "session")
	if session.Values["admin_logged"] == "true" {
		http.Redirect(w, r, "/admin", http.StatusFound)
		return
	}

	admT.ExecuteTemplate(w, "login", nil)
}

func authAdmin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := store.Get(r, "session")
	if session.Values["admin_logged"] == "true" {
		http.Redirect(w, r, "/admin", http.StatusFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Println("form parsing: ", err)
		http.Error(w, "Problems with fetching your data from the form. Please try again", http.StatusInternalServerError)
		return
	}

	l, err := ldap.Dial("tcp", ldapServer)
	if err != nil {
		log.Fatal("could not connect to ldap server: ", err)
	}

	bindUsername := os.Getenv("LDAP_LOGIN")
	bindPassword := os.Getenv("LDAP_PASSWORD")
	err = l.Bind(bindUsername, bindPassword)
	if err != nil {
		log.Fatal("could not to bind: ", err)
	}

	username := r.FormValue("username")
	searchRequest := ldap.NewSearchRequest(
		"dc=mc,dc=asu,dc=ru",
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(memberOf=cn=billing,ou=groups,ou=vc,dc=mc,dc=asu,dc=ru)(samAccountName=%s))", username),
		[]string{"dn"},
		nil,
	)
	sr, err := l.Search(searchRequest)
	if err != nil {
		log.Fatal("could not to do ldap.Search: ", err)
	}

	if len(sr.Entries) != 1 {
		log.Println("Uncorrect login")
		http.Redirect(w, r, "/admin-login", http.StatusFound)
		return
	}

	// Bind as the user to verify their password
	userdn := sr.Entries[0].DN
	password := r.FormValue("password")
	err = l.Bind(userdn, password)
	if err != nil {
		log.Println("Uncorrect password: ", err)
		http.Redirect(w, r, "/admin-login", http.StatusFound)
		return
	}
	l.Close()

	session.Values["admin_logged"] = "true"
	session.Save(r, w)
	http.Redirect(w, r, "/admin", http.StatusFound)
}

func adminIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := store.Get(r, "session")
	if session.Values["admin_logged"] == "true" {
		admT.ExecuteTemplate(w, "index", nil)
		return
	}
	http.Redirect(w, r, "/admin-login", http.StatusFound)
}

func adminLogout(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := store.Get(r, "session")
	session.Values["admin_logged"] = "false"
	session.Save(r, w)
	http.Redirect(w, r, "/admin-login", http.StatusFound)
}

func newUserForm(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := store.Get(r, "session")
	if session.Values["admin_logged"] == "true" {
		admT.ExecuteTemplate(w, "new-user-form", nil)
		return
	}
	http.Redirect(w, r, "/admin-login", http.StatusFound)
}

func addNewUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if err := r.ParseForm(); err != nil {
		log.Println("form parsing: ", err)
		http.Error(w, "Problems with fetching your data from the form. Please try again", http.StatusInternalServerError)
		return
	}
}
