package main

import (
	"fmt"
	"os"

	ldap "gopkg.in/ldap.v3"
)

type loginLDAPerror struct {
	message string
}

func newLoginLDAPerror(message string) *loginLDAPerror {
	return &loginLDAPerror{
		message: message,
	}
}

func (e *loginLDAPerror) Error() string {
	return e.message
}

const ldapServer = "ads.mc.asu.ru:3268"

func ldapAuth(login, password string) error {
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

	searchRequest := ldap.NewSearchRequest(
		"dc=mc,dc=asu,dc=ru",
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(memberOf=cn=billing,ou=groups,ou=vc,dc=mc,dc=asu,dc=ru)(samAccountName=%s))", login),
		[]string{"dn"},
		nil,
	)
	sr, err := l.Search(searchRequest)
	if err != nil {
		return newLoginLDAPerror("Не удалось найти пользователя с таким логином")
	}

	if len(sr.Entries) != 1 {
		return newLoginLDAPerror("Неверный логин")
	}

	// Подключаемся пользователем для проверки пароля
	userdn := sr.Entries[0].DN
	err = l.Bind(userdn, password)
	if err != nil {
		return newLoginLDAPerror("Неверный пароль")
	}

	return nil
}
