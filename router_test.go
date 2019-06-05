package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRouter(t *testing.T) {
	ts := httptest.NewServer(newRouter())
	defer ts.Close()
}

func TestIndexHandler(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		indexHandler(w, r)
	}))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	require.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestLoginHandler(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loginHandler(w, r)
	}))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/login")
	require.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestLoginPostHandler(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loginPostHandler(w, r)
	}))
	defer ts.Close()

	var L struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	L.Login = os.Getenv("LDAP_TEST_LOGIN")
	L.Password = os.Getenv("LDAP_TEST_PASSWORD")

	b, err := json.Marshal(&L)
	require.Nil(t, err)
	resp, err := http.Post(ts.URL+"/login", "application/json; charset=utf-8", bytes.NewReader(b))
	require.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var J struct {
		Answer string `json:"answer"`
		Error  string `json:"error,omitempty"`
	}
	err = json.NewDecoder(resp.Body).Decode(&J)
	require.Nil(t, err)
	resp.Body.Close()
	assert.Equal(t, "ok", J.Answer)
	assert.Empty(t, J.Error)

	c := resp.Cookies()[0]
	assert.NotEmpty(t, c)
	assert.Equal(t, "jwt", c.Name)
	token, _ := jwt.Parse(c.Value, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_KEY")), nil
	})
	claims := token.Claims.(jwt.MapClaims)
	assert.Equal(t, true, token.Valid)
	assert.Equal(t, L.Login, claims["login"])

	L.Password = "bad password"
	b, err = json.Marshal(&L)
	require.Nil(t, err)
	resp, err = http.Post(ts.URL+"/login", "application/json; charset=utf-8", bytes.NewReader(b))
	require.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	err = json.NewDecoder(resp.Body).Decode(&J)
	require.Nil(t, err)
	assert.Equal(t, "bad", J.Answer)
	assert.Equal(t, J.Error, "Неверный логин или пароль.")
	resp.Body.Close()
}
