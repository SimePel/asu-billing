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

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func adminLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	t.ExecuteTemplate(w, "admin-login", nil)
}
