package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLDAPauth(t *testing.T) {
	wrongLogin := "abcde"
	wrongPassword := "abcde"
	err := ldapAuth(wrongLogin, wrongPassword)
	assert.EqualError(t, err, "Неверный логин")

	login := os.Getenv("LDAP_TEST_LOGIN")
	err = ldapAuth(login, wrongPassword)
	assert.EqualError(t, err, "Неверный пароль")

	password := os.Getenv("LDAP_TEST_PASSWORD")
	err = ldapAuth(login, password)
	assert.Nil(t, err)
}
