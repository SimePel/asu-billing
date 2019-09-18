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
			expirationDate := time.Until(user.ExpiredDate)
			f := createTryToRenewPaymentFunc(mysql, user)
			time.AfterFunc(expirationDate, f)
			continue
		}
		tryToRenewPayment(mysql, user)
	}

	return nil
}
