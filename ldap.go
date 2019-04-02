package main

import (
	"fmt"
	"net/http"
	"os"

	ldap "gopkg.in/ldap.v3"
)

const (
	ldapServer = "ads.mc.asu.ru:3268"
)

func ldapAuth(w http.ResponseWriter, r *http.Request, searchRequest *ldap.SearchRequest) error {
	l, err := ldap.Dial("tcp", ldapServer)
	if err != nil {
		return fmt.Errorf("Не удалось подключиться к ldap серверу")
	}
	defer l.Close()

	bindUsername := os.Getenv("LDAP_LOGIN")
	bindPassword := os.Getenv("LDAP_PASSWORD")
	err = l.Bind(bindUsername, bindPassword)
	if err != nil {
		return fmt.Errorf("Не удалось подключиться read only пользователем")
	}

	sr, err := l.Search(searchRequest)
	if err != nil {
		return fmt.Errorf("Не удалось найти пользователя с таким логином")
	}

	if len(sr.Entries) != 1 {
		return fmt.Errorf("Неверный логин")
	}

	// Подключаемся пользователем для проверки пароля
	userdn := sr.Entries[0].DN
	password := r.FormValue("password")
	err = l.Bind(userdn, password)
	if err != nil {
		return fmt.Errorf("Неверный пароль")
	}

	return nil
}
