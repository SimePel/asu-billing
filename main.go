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
		if user.ExpiredDate.After(time.Now()) {
			paymentFunc := createTryToRenewPaymentFunc(mysql, user)
			time.AfterFunc(time.Until(user.ExpiredDate), paymentFunc)

			// Если оказались внутри периода из трех дней, то ничего страшного не произойдет
			// Until вернет отрицательное число и функция выполниться в этот же момент
			notificationDate := user.ExpiredDate.AddDate(0, 0, -3)
			notificationFunc := createSendNotificationFunc(user)
			time.AfterFunc(time.Until(notificationDate), notificationFunc)
			continue
		}
		tryToRenewPayment(mysql, user)
	}

	return nil
}
