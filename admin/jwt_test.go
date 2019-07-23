package main

import (
	"net/http"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateJWTtoken(t *testing.T) {
	tokenString, err := createJWTtoken("login")
	require.Nil(t, err)
	token, err := parseJWTtoken(tokenString)
	require.Nil(t, err)
	assert.True(t, token.Valid)
	assert.Equal(t, tokenString, token.Raw)
	claims, ok := token.Claims.(jwt.MapClaims)
	assert.True(t, ok)
	assert.Equal(t, claims["login"], "login")
}

func TestGetJWTtokenFromCookies(t *testing.T) {
	expectedToken, err := createJWTtoken("login")
	require.Nil(t, err)
	cookies := []*http.Cookie{
		{
			Name:  "someCookie",
			Value: "someValue",
		},
		{
			Name:     "jwt",
			Value:    expectedToken,
			HttpOnly: true,
			SameSite: 3,
		},
	}
	actualToken, err := getJWTtokenFromCookies(cookies)
	require.Nil(t, err)
	assert.Equal(t, expectedToken, actualToken.Raw)

	cookies = []*http.Cookie{}
	_, err = getJWTtokenFromCookies(cookies)
	require.EqualError(t, err, "jwt token was not found in cookies.")
}
