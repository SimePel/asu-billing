package main

import (
	"net/http"
	"os"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateJWTtoken(t *testing.T) {
	tokenString, err := createJWTtoken("login")
	require.NoError(t, err)

	token, err := parseJWTtoken(tokenString)
	require.NoError(t, err)
	assert.True(t, token.Valid)
	assert.Equal(t, tokenString, token.Raw)

	claims, ok := token.Claims.(jwt.MapClaims)
	assert.True(t, ok)
	assert.Equal(t, claims["login"], "login")

	tokenString, err = createJWTtoken("fail")
	require.NoError(t, err)

	err = os.Setenv("JWT_KEY", "")
	require.NoError(t, err)

	_, err = parseJWTtoken(tokenString)
	require.Error(t, err)

	err = os.Setenv("JWT_KEY", "returnJWT_KEY")
	require.NoError(t, err)
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
