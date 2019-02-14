package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/julienschmidt/httprouter"
	ldap "gopkg.in/ldap.v3"
)

var (
	usrT = template.Must(template.New("usr").ParseGlob("templates/usr/*.html"))
)

func userIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	usrT.ExecuteTemplate(w, "index", nil)
}

func userLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	usrT.ExecuteTemplate(w, "login", nil)
}

func authUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

	login := r.FormValue("login")
	pieces := strings.Split(login, "\\")
	searchRequest := ldap.NewSearchRequest(
		fmt.Sprintf("dc=%s,dc=asu,dc=ru", pieces[0]),
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(samAccountName=%s)", pieces[1]),
		[]string{"dn"},
		nil,
	)
	sr, err := l.Search(searchRequest)
	if err != nil {
		log.Fatal("could not to do ldap.Search: ", err)
	}

	if len(sr.Entries) != 1 {
		log.Println("Uncorrect login")
		http.Redirect(w, r, "/user-login", http.StatusFound)
		return
	}

	// Bind as the user to verify their password
	userdn := sr.Entries[0].DN
	password := r.FormValue("password")
	err = l.Bind(userdn, password)
	if err != nil {
		log.Println("Uncorrect password: ", err)
		http.Redirect(w, r, "/user-login", http.StatusFound)
		return
	}
	l.Close()
}
