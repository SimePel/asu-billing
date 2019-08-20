package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRouter(t *testing.T) {
	ts := httptest.NewServer(newRouter())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/login")
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	resp, err = http.Post(ts.URL+"/login", "application/json; charset=utf-8", bytes.NewBuffer([]byte("{}")))
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	resp, err = http.Get(ts.URL + "/")
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
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

func TestUserHandler(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userHandler(w, r)
	}))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/user")
	require.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestAddUserHandler(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		addUserHandler(w, r)
	}))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/add-user")
	require.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestLogoutHandler(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := http.Cookie{
			Name:    "jwt",
			Value:   "token",
			Expires: time.Now().AddDate(0, 0, 1),
		}
		r.AddCookie(&c)
		logoutHandler(w, r)
	}))
	defer ts.Close()

	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if req.URL.Path == "/login" {
				return errors.New("ОК")
			}
			return nil
		},
	}

	resp, err := client.Get(ts.URL + "/logout")
	require.NotNil(t, err)
	assert.Equal(t, 303, resp.StatusCode)

	actualCookie := resp.Cookies()[0]
	assert.Equal(t, "", actualCookie.Value)
	assert.Greater(t, time.Now().Sub(actualCookie.Expires).Seconds(), float64(0))
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

	token, err := getJWTtokenFromCookies(resp.Cookies())
	require.Nil(t, err)
	claims := token.Claims.(jwt.MapClaims)
	assert.True(t, token.Valid)
	assert.NotEmpty(t, claims["login"])

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
