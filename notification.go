package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

var smsNotificationStatus = true

func sendSMS(phone, message string) error {
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
