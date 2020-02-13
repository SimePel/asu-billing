package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

var smsNotificationStatus = true

func tryToRenewPayment(mysql MySQL, id int) {
	user, err := mysql.GetUserByID(id)
	if err != nil {
		log.Println(err)
		return
	}

	if user.ExpiredDate.After(time.Now()) {
		return
	}

	if user.IsDeactivated {
		return
	}

	if user.hasEnoughMoneyForPayment() {
		expirationDate, err := mysql.PayForNextMonth(user)
		if err != nil {
			log.Println(err)
			return
		}

		paymentFunc := createTryToRenewPaymentFunc(mysql, user)
		time.AfterFunc(time.Until(expirationDate), paymentFunc)

		notificationDate := expirationDate.AddDate(0, 0, -3)
		notificationFunc := createSendNotificationFunc(mysql, user)
		time.AfterFunc(time.Until(notificationDate), notificationFunc)
	}
}

func createTryToRenewPaymentFunc(mysql MySQL, u User) func() {
	return func() {
		tryToRenewPayment(mysql, int(u.ID))
	}
}

func createSendNotificationFunc(mysql MySQL, u User) func() {
	return func() {
		err := sendNotification(mysql, int(u.ID))
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func sendNotification(mysql MySQL, id int) error {
	user, err := mysql.GetUserByID(id)
	if err != nil {
		return fmt.Errorf("cannot get user by id: %v", err)
	}

	if user.Balance >= user.Tariff.Price {
		return nil
	}

	if user.ExpiredDate.After(time.Now().AddDate(0, 0, 3)) {
		return nil
	}

	message := fmt.Sprintf("На ЛС: %v %vр. Пополните счет за проводное подключение к сети АГУ", user.Agreement, user.Balance)
	err = sendSMS(user.Phone, message)
	if err != nil {
		return fmt.Errorf("cannot send sms: %v", err)
	}

	return nil
}

func sendSMS(phone, message string) error {
	if !smsNotificationStatus {
		return nil
	}

	user := os.Getenv("BEELINE_USER")
	password := os.Getenv("BEELINE_PASS")

	resp, err := http.PostForm("https://beeline.amega-inform.ru/sms_send/", url.Values{
		"user": {user}, "pass": {password}, "action": {"post_sms"},
		"message": {message}, "target": {phone}, "sender": {"asu"},
	})
	if err != nil {
		return fmt.Errorf("cannot do post request: %v", err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("cannot read from body: %v", err)
	}

	log.Println(string(body))
	return nil
}
