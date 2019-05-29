package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

var smsNotificationStatus = true

func smsStatus(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	change := r.FormValue("change")
	if change == "" {
		w.Write([]byte(strconv.FormatBool(smsNotificationStatus)))
		return
	}
	smsNotificationStatus = !smsNotificationStatus
	w.Write([]byte(strconv.FormatBool(smsNotificationStatus)))
}

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

	fmt.Println(string(body))
	return nil
}

func getStatusOfSMS(id string) (string, error) {
	user := os.Getenv("BEELINE_USER")
	password := os.Getenv("BEELINE_PASS")

	resp, err := http.PostForm("https://beeline.amega-inform.ru/sms_send/", url.Values{
		"user": {user}, "pass": {password}, "action": {"status"}, "sms_id": {id},
	})
	if err != nil {
		return "", fmt.Errorf("cannot do post request: %v", err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("cannot read from body: %v", err)
	}

	return string(body), nil
}
