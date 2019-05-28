package main

import (
	"fmt"
	"net/smtp"
)

func confirmEmail(recipient, url string) error {
	return sendEmail(recipient, "Подтверждение email адреса",
		"Перейдите по данной ссылке, чтобы подтвердить ваш email.\r\n"+url)
}

func sendEmail(recipient, subject, message string) error {
	c, err := smtp.Dial("mail.asu.ru:25")
	if err != nil {
		return fmt.Errorf("Cannot dial with mail server: %v", err)
	}

	err = c.Mail("billing@asu.ru")
	if err != nil {
		return fmt.Errorf("Cannot set from address: %v", err)
	}

	err = c.Rcpt(recipient)
	if err != nil {
		return fmt.Errorf("Cannot set recipient address: %v", err)
	}

	wc, err := c.Data()
	if err != nil {
		return fmt.Errorf("Cannot get data stream: %v", err)
	}

	msg := "To: " + recipient + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" + message + "\r\n"

	_, err = fmt.Fprint(wc, msg)
	if err != nil {
		return fmt.Errorf("Cannot fprint to data stream: %v", err)
	}

	err = wc.Close()
	if err != nil {
		return fmt.Errorf("Cannot close data stream: %v", err)
	}

	err = c.Quit()
	if err != nil {
		return fmt.Errorf("Cannot send quit message: %v", err)
	}

	return nil
}
