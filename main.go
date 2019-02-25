package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/julienschmidt/httprouter"
)

var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))

func init() {
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
	}
}

func main() {
	router := httprouter.New()

	router.ServeFiles("/assets/*filepath", http.Dir("assets/"))

	router.GET("/admin-login", adminLogin)
	router.GET("/admin", adminIndex)
	router.GET("/admin-logout", adminLogout)
	router.GET("/user-logout", userLogout)
	router.GET("/user-login", userLogin)
	router.GET("/add-user", newUserForm)
	router.GET("/user-info", userInfo)
	router.GET("/user", userIndex)
	router.GET("/pay", payForm)

	router.POST("/admin-login", authAdmin)
	router.POST("/user-login", authUser)
	router.POST("/add-user", addNewUser)
	router.POST("/pay", pay)

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
