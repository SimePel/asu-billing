package main

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func main() {
	router := httprouter.New()

	router.ServeFiles("/assets/*filepath", http.Dir("assets/"))

	router.GET("/admin-login", adminLogin)
	router.GET("/admin-index", adminIndex)
	router.GET("/user-login", userLogin)

	router.POST("/admin-login", authAdmin)
	router.POST("/user-login", authUser)

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
