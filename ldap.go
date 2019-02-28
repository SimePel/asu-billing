package main

import (
	"fmt"
	"net/http"
	"os"

	ldap "gopkg.in/ldap.v3"
)

func ldapAuth(w http.ResponseWriter, r *http.Request, searchRequest *ldap.SearchRequest) error {
	l, err := ldap.Dial("tcp", ldapServer)
	if err != nil {
		return fmt.Errorf("could not connect to ldap server: %v", err)
	}
	defer l.Close()

	bindUsername := os.Getenv("LDAP_LOGIN")
	bindPassword := os.Getenv("LDAP_PASSWORD")
	err = l.Bind(bindUsername, bindPassword)
	if err != nil {
		return fmt.Errorf("could not to bind: %v", err)
	}

	sr, err := l.Search(searchRequest)
	if err != nil {
		return fmt.Errorf("could not to do ldap.Search: %v", err)
	}

	if len(sr.Entries) != 1 {
		return fmt.Errorf("Неверный логин")
	}

	// Bind as the user to verify their password
	userdn := sr.Entries[0].DN
	password := r.FormValue("password")
	err = l.Bind(userdn, password)
	if err != nil {
		return fmt.Errorf("Неверный пароль")
	}

	return nil
}
