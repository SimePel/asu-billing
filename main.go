package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	err := restorePaymentsTimers()
	if err != nil {
		log.Fatal(err)
	}

	archiveOldUsers()

	r := newRouter()
	log.Fatal(http.ListenAndServe(":8081", r))
}

func restorePaymentsTimers() error {
	mysql := MySQL{db: initializeDB()}
	users, err := mysql.GetAllUsers()
	if err != nil {
		return fmt.Errorf("cannot get all users: %v", err)
	}

	for _, user := range users {
		if user.IsDeactivated || user.IsArchived {
			continue
		}

		if user.ExpiredDate.After(time.Now()) {
			paymentFunc := createTryToRenewPaymentFunc(mysql, user)
			time.AfterFunc(time.Until(user.ExpiredDate), paymentFunc)

			// Если оказались внутри периода из трех дней, то ничего страшного не произойдет
			// Until вернет отрицательное число и функция выполниться в этот же момент
			notificationDate := user.ExpiredDate.AddDate(0, 0, -3)
			notificationFunc := createSendNotificationFunc(mysql, user)
			time.AfterFunc(time.Until(notificationDate), notificationFunc)
			continue
		}
		tryToRenewPayment(mysql, int(user.ID))
	}

	return nil
}

func archiveOldUsers() {
	mysql := MySQL{db: initializeDB()}
	users, err := mysql.GetAllUsers()
	if err != nil {
		log.Println("cannot get all users: ", err)
		return
	}

	threeMonthsAgo := time.Now().AddDate(0, -3, 0)
	for _, user := range users {
		if len(user.Payments) <= 0 || user.IsEmployee {
			continue
		}

		if user.Payments[len(user.Payments)-1].Date.Before(threeMonthsAgo) {
			err = mysql.ArchiveUserByID(int(user.ID))
			if err != nil {
				log.Println(err)
			}
		}
	}

	nextCheck := time.Now().AddDate(0, 1, 0)
	time.AfterFunc(time.Until(nextCheck), archiveOldUsers)
}
