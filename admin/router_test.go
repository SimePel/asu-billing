package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
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

func TestEditUserHandler(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		editUserHandler(w, r)
	}))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/edit-user")
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

func TestAddUserPostHandler(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		addUserPostHandler(w, r)
	}))
	defer ts.Close()

	expected := struct {
		Name            string
		Agreement       string
		Login           string
		Phone           string
		Room            string
		Tariff          int
		ConnectionPlace string
	}{
		"Tестовый Тест Тестович4",
		"П-004",
		"aloha.125",
		"88005553554",
		"555",
		1,
		"",
	}

	formValues := url.Values{}
	formValues.Add("name", expected.Name)
	formValues.Add("agreement", expected.Agreement)
	formValues.Add("login", expected.Login)
	formValues.Add("phone", expected.Phone)
	formValues.Add("room", expected.Room)
	formValues.Add("connectionPlace", expected.ConnectionPlace)
	formValues.Add("tariff", strconv.Itoa(expected.Tariff))

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.PostForm(ts.URL+"/add-user", formValues)
	require.NoError(t, err)
	assert.Equal(t, 303, resp.StatusCode)

	mysql := MySQL{db: openTestDBconnection()}
	id, err := mysql.GetUserIDbyLogin(expected.Login)
	require.NoError(t, err)

	user, err := mysql.GetUserByID(int(id))
	require.NoError(t, err)

	assert.Equal(t, expected.Name, user.Name)
	assert.Equal(t, expected.Agreement, user.Agreement)
	assert.Equal(t, expected.Login, user.Login)
	assert.Equal(t, expected.Phone, user.Phone)
	assert.Equal(t, expected.Room, user.Room)
	assert.Equal(t, expected.ConnectionPlace, user.ConnectionPlace)
	assert.Equal(t, expected.Tariff, user.Tariff.ID)
}

func TestEditUserPostHandler(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		editUserPostHandler(w, r)
	}))
	defer ts.Close()

	user := User{
		Name:      "Tестовый Тест Тестович127",
		Agreement: "П-127",
		Login:     "update.128",
		Tariff: Tariff{
			ID: 1,
		},
	}

	mysql := MySQL{db: openTestDBconnection()}
	id, err := mysql.AddUser(user)
	require.NoError(t, err)

	expected := struct {
		Name            string
		Agreement       string
		Login           string
		Phone           string
		Room            string
		Tariff          int
		ConnectionPlace string
	}{
		"Tестовый Тест Тестович128",
		"П-128",
		"wasUpdated.128",
		"88005553128",
		"128",
		1,
		"рандом",
	}

	formValues := url.Values{}
	formValues.Add("id", strconv.Itoa(id))
	formValues.Add("name", expected.Name)
	formValues.Add("agreement", expected.Agreement)
	formValues.Add("login", expected.Login)
	formValues.Add("phone", expected.Phone)
	formValues.Add("room", expected.Room)
	formValues.Add("tariff", strconv.Itoa(expected.Tariff))
	formValues.Add("connectionPlace", expected.ConnectionPlace)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.PostForm(ts.URL+"/edit-user", formValues)
	require.NoError(t, err)
	assert.Equal(t, 303, resp.StatusCode)

	updatedUser, err := mysql.GetUserByID(id)
	require.NoError(t, err)

	assert.Equal(t, expected.Name, updatedUser.Name)
	assert.Equal(t, expected.Agreement, updatedUser.Agreement)
	assert.Equal(t, expected.Login, updatedUser.Login)
	assert.Equal(t, expected.Phone, updatedUser.Phone)
	assert.Equal(t, expected.Room, updatedUser.Room)
	assert.Equal(t, expected.Tariff, updatedUser.Tariff.ID)
	assert.Equal(t, expected.ConnectionPlace, updatedUser.ConnectionPlace)
}

func TestPaymentPostHandler(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paymentPostHandler(w, r)
	}))
	defer ts.Close()

	user := User{
		ID:        100,
		Paid:      false,
		Name:      "Тестовый Тест Тестович100",
		Agreement: "П-100",
		Phone:     "88005553100",
		Login:     "blabla.1000",
		Balance:   0,
		Tariff: Tariff{
			ID:    1,
			Name:  "Базовый-30",
			Price: 200,
		},
	}

	mysql := MySQL{db: openTestDBconnection()}
	userID, err := mysql.AddUser(user)
	require.NoError(t, err)

	var payment struct {
		UserID int `json:"id"`
		Sum    int `json:"sum"`
	}
	payment.UserID = userID
	payment.Sum = 100
	b, err := json.Marshal(&payment)
	require.NoError(t, err)

	resp, err := http.Post(ts.URL+"/payment", "application/json; charset=utf-8", bytes.NewReader(b))
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	actualUser, err := mysql.GetUserByID(userID)
	require.NoError(t, err)

	assert.Equal(t, payment.Sum, actualUser.Balance)
	assert.Equal(t, user.Paid, actualUser.Paid)

	resp, err = http.Post(ts.URL+"/payment", "application/json; charset=utf-8", bytes.NewReader(b))
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	actualUser, err = mysql.GetUserByID(userID)
	require.NoError(t, err)

	assert.Equal(t, 0, actualUser.Balance)
	assert.Equal(t, true, actualUser.Paid)

	// Еще проверить записи в табличке payments
}

func TestGetStatsAboutUsers(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		getStatsAboutUsers(w, r)
	}))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/stats")
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var J struct {
		ActiveUsersCount   int `json:"active_users_count"`
		InactiveUsersCount int `json:"inactive_users_count"`
		AllMoney           int `json:"all_money"`
	}

	err = json.NewDecoder(resp.Body).Decode(&J)
	require.NoError(t, err)
	resp.Body.Close()

	assert.NotZero(t, J.ActiveUsersCount)
	assert.NotZero(t, J.InactiveUsersCount)
	assert.NotZero(t, J.AllMoney)
}
