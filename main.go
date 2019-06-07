package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/sessions"
	"github.com/julienschmidt/httprouter"
)

const (
	timeoutForNotification = 6 * time.Hour
	timeoutForWithdraw     = 1 * time.Minute
)

var (
	store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
)

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

	router.GET("/admin-login", accessLog(adminLogin))
	router.GET("/adm", accessLog(adminAuthCheck(adminIndex)))
	router.GET("/admin-logout", accessLog(adminLogout))
	router.GET("/user-logout", accessLog(userLogout))
	router.GET("/user-login", accessLog(userLogin))
	router.GET("/add-user", accessLog(adminAuthCheck(newUserForm)))
	router.GET("/user-info", accessLog(adminAuthCheck(userInfo)))
	router.GET("/edit-user", accessLog(adminAuthCheck(userEditForm)))
	router.GET("/delete-user", accessLog(adminAuthCheck(deleteUser)))
	router.GET("/", accessLog(userAuthCheck(userIndex)))
	router.GET("/pay", accessLog(adminAuthCheck(payForm)))
	router.GET("/sms-status", accessLog(adminAuthCheck(smsStatus)))
	router.GET("/stats", accessLog(adminAuthCheck(usersStatistics)))
	// router.GET("/settings", accessLog(userAuthCheck(userSettings)))
	// router.GET("/confirm-settings", accessLog(userAuthCheck(confirmSettings)))

	router.POST("/admin-login", accessLog(authAdmin))
	router.POST("/user-login", accessLog(authUser))
	router.POST("/add-user", accessLog(adminAuthCheck(addNewUser)))
	router.POST("/edit-user", accessLog(adminAuthCheck(editUser)))
	router.POST("/pay", accessLog(adminAuthCheck(pay)))
	// router.POST("/settings", accessLog(userAuthCheck(editUserSettings)))

	go func() {
		for {
			time.Sleep(timeoutForNotification)
			if !smsNotificationStatus {
				continue
			}
			stmt, err := db.Prepare(`SELECT id, account, phone, balance FROM bl_users
				WHERE activity=1 AND expired_date BETWEEN DATE_SUB(DATE_ADD(NOW(), INTERVAL 2 DAY), INTERVAL 3 HOUR)
					AND DATE_ADD(DATE_ADD(NOW(), INTERVAL 2 DAY), INTERVAL 3 HOUR);`)
			if err != nil {
				log.Println("Cannot prepare sql statement.", err)
				continue
			}

			users, err := checkPaymentNeed(stmt)
			if err != nil {
				log.Println("Cannot check payment need.", err)
				continue
			}

			err = sendPaymentNotification(users)
			if err != nil {
				log.Println("Cannot send payment notification.", err)
			}
		}
	}()

	go func() {
		for {
			time.Sleep(timeoutForWithdraw)
			stmt, err := db.Prepare(`SELECT id, account, phone, balance FROM bl_users
				WHERE expired_date BETWEEN NOW() AND DATE_ADD(NOW(), INTERVAL 60 SECOND);`)
			if err != nil {
				log.Println("Cannot prepare sql statement.", err)
				continue
			}

			users, err := checkPaymentNeed(stmt)
			if err != nil {
				log.Println("Cannot check payment need.", err)
			}

			for _, u := range users {
				err = withdrawMoney(u.ID)
				if err != nil {
					log.Println("Cannot withdraw money from user with id= ", u.ID, err)
				}
			}
		}
	}()

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
