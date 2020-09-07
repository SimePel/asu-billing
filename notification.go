package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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

	L := struct {
		GroupIDs []int  `json:"group_ids"`
		Message  string `json:"message"`
		Phones   string `json:"phones"`
	}{
		GroupIDs: []int{},
		Message:  message,
		Phones:   phone,
	}

	b, err := json.Marshal(&L)
	if err != nil {
		return fmt.Errorf("cannot marshal json: %v", err)
	}

	req, err := http.NewRequest("POST", "https://sms-gate.asu.ru/v1/send-sms", bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("cannot make post request: %v", err)
	}

	c := &http.Cookie{
		Name:     "jwt",
		Value:    os.Getenv("JWT_FOR_SMS_GATE"),
		HttpOnly: true,
		Expires:  time.Now().AddDate(0, 0, 1),
	}
	req.AddCookie(c)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("cannot send request: %v", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("cannot read response body: %v", err)
	}

	log.Println(string(body)) // В будующем тут будет лежать статус сообщения
	return nil
}
