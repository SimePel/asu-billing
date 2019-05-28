package main

import (
	"fmt"
	"log"
)

// Возможно в будующем появится выбор между способами оповещения
func sendPaymentNotification(users []User) error {
	for _, user := range users {
		message := fmt.Sprintf("На ЛС: %v %vр. Пополните счет за проводное подключение к сети АГУ", user.Agreement, user.Money)
		err := sendSMS(user.Phone, message)
		if err != nil {
			log.Println("Cannot send sms. ", err)
		}
	}

	return nil
}
