package main

import (
	"fmt"
	"net/smtp"
	"os"
)

func confirmEmail(recipient, url string) error {
	auth := smtp.PlainAuth("", os.Getenv("MAIL_LOGIN"), os.Getenv("MAIL_PASS"), "mail.asu.ru")

	to := []string{recipient}
	msg := []byte("To: " + recipient + "\r\n" +
		"Subject: Подтвердите email адрес!\r\n" +
		"\r\n" +
		"Перейдите по данной ссылке, чтобы подтвердить ваш email.\r\n" +
		url + "\r\n")
	err := smtp.SendMail("mail.asu.ru:25", auth, "billing@asu.ru", to, msg)
	if err != nil {
		return fmt.Errorf("Unable to send email: %v", err)
	}

	return nil
}
