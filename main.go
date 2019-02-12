package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

var (
	t = template.Must(template.New("adm").ParseGlob("templates/adm/*.html"))
)

func main() {
	router := httprouter.New()

	router.ServeFiles("/assets/*filepath", http.Dir("assets/"))
	router.GET("/admin-login", adminLogin)
	router.POST("/admin-login", authAdmin)

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func adminLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	t.ExecuteTemplate(w, "admin-login", nil)
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
