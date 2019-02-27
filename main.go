package main

import (
	"log"
	"net/http"
	"os"
	"time"

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
	router.GET("/admin", adminAuthCheck(adminIndex))
	router.GET("/admin-logout", adminLogout)
	router.GET("/user-logout", userLogout)
	router.GET("/user-login", userLogin)
	router.GET("/add-user", adminAuthCheck(newUserForm))
	router.GET("/user-info", adminAuthCheck(userInfo))
	router.GET("/edit-user", adminAuthCheck(userEditForm))
	router.GET("/user", userIndex)
	router.GET("/pay", adminAuthCheck(payForm))

	router.POST("/admin-login", authAdmin)
	router.POST("/user-login", authUser)
	router.POST("/add-user", adminAuthCheck(addNewUser))
	router.POST("/edit-user", adminAuthCheck(editUser))
	router.POST("/pay", adminAuthCheck(pay))

	go func() {
		for {
			time.Sleep(time.Hour * 12)
			err := turnOffInactiveUsers()
			if err != nil {
				log.Fatal(err)
			}
		}
	}()

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
