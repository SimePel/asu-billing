package main

import (
	"fmt"
	"log"
	"net/smtp"
)

func confirmEmail(recipient, url string) error {
	return sendEmail(recipient, "Подтверждение email адреса",
		"Перейдите по данной ссылке, чтобы подтвердить ваш email.\r\n"+url)
}

func sendPaymentNotification(users []User) error {
	for _, user := range users {
		err := sendEmail(user.Email, "Уведомление об оплате",
			fmt.Sprintf("%v, %v у вас заканчивается срок действия вашего интернет соединения.\r\nДля продолжения использования, оплатите установленную сумму в кассе М корпуса и принесите квитанцию об оплате в 103 кабинет Л корпуса.", user.Name, user.PaymentsEnds.Format("2.01 в 15:04")))
		if err != nil {
			log.Println("Cannot send email.", err)
		}
	}

	return nil
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
