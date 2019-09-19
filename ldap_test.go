package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	bindUsername := os.Getenv("LDAP_LOGIN")
	err = os.Setenv("LDAP_LOGIN", "a")
	require.NoError(t, err)

	err = ldapAuth("l", "p")
	require.Error(t, err)

	err = os.Setenv("LDAP_LOGIN", bindUsername)
	require.NoError(t, err)
}
