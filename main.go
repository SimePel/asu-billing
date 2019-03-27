package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
	"github.com/julienschmidt/httprouter"
)

var (
	dsn   = fmt.Sprintf("%v:%v@tcp(10.0.0.33)/billingdev?parseTime=true", os.Getenv("MYSQL_LOGIN"), os.Getenv("MYSQL_PASS"))
	store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
	db    = newDB()
	prod  bool
)

func newDB() *sql.DB {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	db.SetConnMaxLifetime(2 * time.Minute)
	db.SetMaxIdleConns(4)
	db.SetMaxOpenConns(8)

	return db
}

func init() {
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
	}
	flag.BoolVar(&prod, "prod", false, "Enable production mode")
}

func main() {
	router := httprouter.New()

	router.ServeFiles("/assets/*filepath", http.Dir("assets/"))

	router.GET("/admin-login", accessLog(adminLogin))
	router.GET("/admin", accessLog(adminAuthCheck(adminIndex)))
	router.GET("/admin-logout", accessLog(adminLogout))
	router.GET("/user-logout", accessLog(userLogout))
	router.GET("/user-login", accessLog(userLogin))
	router.GET("/add-user", accessLog(adminAuthCheck(newUserForm)))
	router.GET("/user-info", accessLog(adminAuthCheck(userInfo)))
	router.GET("/edit-user", accessLog(adminAuthCheck(userEditForm)))
	router.GET("/delete-user", accessLog(adminAuthCheck(deleteUser)))
	router.GET("/", accessLog(userIndex))
	router.GET("/pay", accessLog(adminAuthCheck(payForm)))

	router.POST("/admin-login", accessLog(authAdmin))
	router.POST("/user-login", accessLog(authUser))
	router.POST("/add-user", accessLog(adminAuthCheck(addNewUser)))
	router.POST("/edit-user", accessLog(adminAuthCheck(editUser)))
	router.POST("/pay", accessLog(adminAuthCheck(pay)))

	flag.Parse()
	var err error
	if prod {
		err = http.ListenAndServeTLS("billing-dev.asu.ru:443", "cert.pem", "privkey.pem", router)
	} else {
		err = http.ListenAndServe(":8080", router)
	}
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
