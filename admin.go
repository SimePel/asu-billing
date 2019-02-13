package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

var (
	admT = template.Must(template.New("adm").ParseGlob("templates/adm/*.html"))
)

func adminLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	admT.ExecuteTemplate(w, "login", nil)
}

func authAdmin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if err := r.ParseForm(); err != nil {
		log.Println("form parsing: ", err)
		http.Error(w, "Problems with fetching your data from the form. Please try again", http.StatusInternalServerError)
		return
	}

	// ldap auth here then sql query to get the role of user
	// if it is alright then redirect to admin page
}

func adminIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	admT.ExecuteTemplate(w, "index", nil)
}

func newUserForm(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	admT.ExecuteTemplate(w, "new-user-form", nil)
}

func addNewUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if err := r.ParseForm(); err != nil {
		log.Println("form parsing: ", err)
		http.Error(w, "Problems with fetching your data from the form. Please try again", http.StatusInternalServerError)
		return
	}
}
