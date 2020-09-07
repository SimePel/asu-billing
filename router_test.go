package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
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
	require.HTTPSuccess(t, indexHandler, "GET", "/", nil)
}

func TestLoginHandler(t *testing.T) {
	require.HTTPSuccess(t, loginHandler, "GET", "/login", nil)

	token, err := createJWTtoken("login")
	require.NoError(t, err)
	c := &http.Cookie{
		Name:     "jwt",
		Value:    token,
		HttpOnly: true,
		Expires:  time.Now().AddDate(0, 1, 0),
		SameSite: 3,
	}

	require.HTTPRedirect(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.AddCookie(c)
		loginHandler(w, r)
	}), "GET", "/login", nil)
}

func TestUserHandler(t *testing.T) {
	require.HTTPSuccess(t, userHandler, "GET", "/user", nil)
}

func TestAddUserHandler(t *testing.T) {
	require.HTTPSuccess(t, addUserHandler, "GET", "/add-user", nil)
}

func TestEditUserHandler(t *testing.T) {
	require.HTTPSuccess(t, editUserHandler, "GET", "/edit-user", nil)
}

func TestNotificationStatusHandler(t *testing.T) {
	require.HTTPSuccess(t, notificationStatusHandler, "GET", "/notification-status", nil)
	require.HTTPBodyContains(t, notificationStatusHandler, "GET", "/notification-status", nil, smsNotificationStatus)
}

func TestSendMassSMSHandler(t *testing.T) {
	require.HTTPSuccess(t, sendMassSMSHandler, "GET", "/send-mass-sms", nil)
}

func TestChangeNotificationStatusHandler(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		changeNotificationStatusHandler(w, r)
	}))
	defer ts.Close()

	r := strings.NewReader("blabla")
	resp, err := http.Post(ts.URL+"/change-notification-status", "text/plain; charset=utf-8", r)
	require.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)

	currentStatus := smsNotificationStatus
	r = strings.NewReader(strconv.FormatBool(currentStatus))
	resp, err = http.Post(ts.URL+"/change-notification-status", "text/plain; charset=utf-8", r)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, smsNotificationStatus, currentStatus)

	r = strings.NewReader(strconv.FormatBool(!currentStatus))
	resp, err = http.Post(ts.URL+"/change-notification-status", "text/plain; charset=utf-8", r)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, smsNotificationStatus, !currentStatus)
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

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(ts.URL + "/logout")
	require.NoError(t, err)
	assert.Equal(t, 303, resp.StatusCode)

	actualCookie := resp.Cookies()[0]
	assert.Equal(t, "", actualCookie.Value)
	assert.Greater(t, time.Since(actualCookie.Expires).Seconds(), float64(0))
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

	invalidJSON := []byte("{Answer: false, Login alesha,}")
	resp, err = http.Post(ts.URL+"/login", "application/json; charset=utf-8", bytes.NewBuffer(invalidJSON))
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	err = json.NewDecoder(resp.Body).Decode(&J)
	require.Nil(t, err)
	assert.Equal(t, "bad", J.Answer)
	assert.Equal(t, J.Error, "Ошибка парсинга json.")
	resp.Body.Close()
}

func TestAddUserPostHandler(t *testing.T) {
	expected := struct {
		Name            string
		IsEmployee      string
		Agreement       string
		Login           string
		Phone           string
		Room            string
		Comment         string
		Tariff          int
		ConnectionPlace string
	}{
		"Tестовый Тест Тестович4",
		"false",
		"П-004",
		"aloha.125",
		"88005553554",
		"555",
		"Важный пользователь",
		1,
		"",
	}

	formValues := url.Values{}
	formValues.Add("name", expected.Name)
	formValues.Add("isEmployee", expected.IsEmployee)
	formValues.Add("agreement", expected.Agreement)
	formValues.Add("login", expected.Login)
	formValues.Add("phone", expected.Phone)
	formValues.Add("room", expected.Room)
	formValues.Add("comment", expected.Comment)
	formValues.Add("connectionPlace", expected.ConnectionPlace)
	formValues.Add("tariff", strconv.Itoa(expected.Tariff))

	require.HTTPRedirect(t, addUserPostHandler, "POST", "/add-user", formValues)

	mysql := MySQL{db: openTestDBconnection()}
	id, err := mysql.GetUserIDbyLogin(expected.Login + "@stud.asu.ru")
	require.NoError(t, err)

	user, err := mysql.GetUserByID(int(id))
	require.NoError(t, err)

	assert.Equal(t, expected.Name, user.Name)
	assert.Equal(t, expected.IsEmployee, strconv.FormatBool(user.IsEmployee))
	assert.Equal(t, expected.Agreement, user.Agreement)
	assert.Equal(t, expected.Login+"@stud.asu.ru", user.Login)
	assert.Equal(t, expected.Phone, user.Phone)
	assert.Equal(t, expected.Room, user.Room)
	assert.Equal(t, expected.Comment, user.Comment)
	assert.Equal(t, expected.ConnectionPlace, user.ConnectionPlace)
	assert.Equal(t, expected.Tariff, user.Tariff.ID)
}

func TestAddEmployeeUserPostHandler(t *testing.T) {
	expected := struct {
		Name            string
		IsEmployee      string
		Agreement       string
		Login           string
		Phone           string
		Room            string
		Comment         string
		Tariff          int
		ConnectionPlace string
	}{
		"Employee",
		"true",
		"П-558",
		"employee.558",
		"",
		"",
		"Наш сотрудник",
		1,
		"",
	}

	formValues := url.Values{}
	formValues.Add("name", expected.Name)
	formValues.Add("isEmployee", expected.IsEmployee)
	formValues.Add("agreement", expected.Agreement)
	formValues.Add("login", expected.Login)
	formValues.Add("phone", expected.Phone)
	formValues.Add("room", expected.Room)
	formValues.Add("comment", expected.Comment)
	formValues.Add("connectionPlace", expected.ConnectionPlace)
	formValues.Add("tariff", strconv.Itoa(expected.Tariff))

	require.HTTPRedirect(t, addUserPostHandler, "POST", "/add-user", formValues)

	mysql := MySQL{db: openTestDBconnection()}
	id, err := mysql.GetUserIDbyLogin(expected.Login + "@stud.asu.ru")
	require.NoError(t, err)

	user, err := mysql.GetUserByID(int(id))
	require.NoError(t, err)

	assert.Equal(t, expected.Name, user.Name)
	assert.Equal(t, expected.IsEmployee, strconv.FormatBool(user.IsEmployee))
	assert.Equal(t, expected.Agreement, user.Agreement)
	assert.Equal(t, expected.Login+"@stud.asu.ru", user.Login)
	assert.Equal(t, expected.Phone, user.Phone)
	assert.Equal(t, expected.Room, user.Room)
	assert.Equal(t, expected.Comment, user.Comment)
	assert.Equal(t, expected.ConnectionPlace, user.ConnectionPlace)
	assert.Equal(t, expected.Tariff, user.Tariff.ID)
	assert.Equal(t, true, user.Paid)
	assert.NotEmpty(t, user.ExpiredDate)
}

func TestEditUserPostHandler(t *testing.T) {
	user := User{
		Name:  "Tестовый Тест Тестович127",
		Login: "update.128",
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
		Comment         string
		ConnectionPlace string
		ExpiredDate     time.Time
		Tariff          int
	}{
		"Tестовый Тест Тестович128",
		"П-128",
		"wasUpdated.128",
		"88005553128",
		"128",
		"улетел",
		"рандом",
		time.Now().AddDate(0, 1, 0),
		1,
	}

	formValues := url.Values{}
	formValues.Add("id", strconv.Itoa(id))
	formValues.Add("name", expected.Name)
	formValues.Add("agreement", expected.Agreement)
	formValues.Add("login", expected.Login)
	formValues.Add("phone", expected.Phone)
	formValues.Add("room", expected.Room)
	formValues.Add("comment", expected.Comment)
	formValues.Add("tariff", strconv.Itoa(expected.Tariff))
	formValues.Add("connectionPlace", expected.ConnectionPlace)
	formValues.Add("expiredDate", expected.ExpiredDate.Format("2006-01-02"))

	require.HTTPRedirect(t, editUserPostHandler, "POST", "/edit-user", formValues)

	updatedUser, err := mysql.GetUserByID(id)
	require.NoError(t, err)

	assert.Equal(t, expected.Name, updatedUser.Name)
	assert.Equal(t, expected.Agreement, updatedUser.Agreement)
	assert.Equal(t, expected.Login, updatedUser.Login)
	assert.Equal(t, expected.Phone, updatedUser.Phone)
	assert.Equal(t, expected.Room, updatedUser.Room)
	assert.Equal(t, expected.Comment, updatedUser.Comment)
	assert.Equal(t, expected.Tariff, updatedUser.Tariff.ID)
	assert.Equal(t, expected.ConnectionPlace, updatedUser.ConnectionPlace)
	assert.Equal(t, expected.ExpiredDate.Format("2006-01-02"), updatedUser.ExpiredDate.Format("2006-01-02"))
}

func TestNotEmployeeUserBecomeEmployeeEditPostHandler(t *testing.T) {
	user := User{
		Name:  "Employee2",
		Login: "Employee.2",
		Tariff: Tariff{
			ID: 1,
		},
	}

	mysql := MySQL{db: openTestDBconnection()}
	id, err := mysql.AddUser(user)
	require.NoError(t, err)

	expected := struct {
		Name       string
		IsEmployee string
		Agreement  string
		Phone      string
		Room       string
		Comment    string
		Tariff     int
	}{
		"Employee2",
		"true",
		"Е-103",
		"88005553128",
		"Е103",
		"админ Л корпуса",
		1,
	}

	formValues := url.Values{}
	formValues.Add("id", strconv.Itoa(id))
	formValues.Add("name", expected.Name)
	formValues.Add("isEmployee", expected.IsEmployee)
	formValues.Add("agreement", expected.Agreement)
	formValues.Add("phone", expected.Phone)
	formValues.Add("room", expected.Room)
	formValues.Add("comment", expected.Comment)
	formValues.Add("tariff", strconv.Itoa(expected.Tariff))

	require.HTTPRedirect(t, editUserPostHandler, "POST", "/edit-user", formValues)

	updatedUser, err := mysql.GetUserByID(id)
	require.NoError(t, err)

	assert.Equal(t, expected.Name, updatedUser.Name)
	assert.Equal(t, expected.IsEmployee, strconv.FormatBool(updatedUser.IsEmployee))
	assert.Equal(t, true, updatedUser.Paid)
	assert.Equal(t, expected.Agreement, updatedUser.Agreement)
	assert.Equal(t, expected.Phone, updatedUser.Phone)
	assert.Equal(t, expected.Room, updatedUser.Room)
	assert.Equal(t, expected.Comment, updatedUser.Comment)
	assert.Equal(t, expected.Tariff, updatedUser.Tariff.ID)
}

func TestEmployeeUserBecomeNotEmployeeEditPostHandler(t *testing.T) {
	user := User{
		Name:       "Employee3",
		IsEmployee: true,
		Login:      "Employee.3",
		Tariff: Tariff{
			ID: 1,
		},
	}

	mysql := MySQL{db: openTestDBconnection()}
	id, err := mysql.AddUser(user)
	require.NoError(t, err)

	expected := struct {
		Name       string
		IsEmployee string
		Agreement  string
		Phone      string
		Room       string
		Comment    string
		Tariff     int
	}{
		"Employee3",
		"false",
		"Е-104",
		"88005123128",
		"Е104",
		"админ M корпуса",
		1,
	}

	formValues := url.Values{}
	formValues.Add("id", strconv.Itoa(id))
	formValues.Add("name", expected.Name)
	formValues.Add("isEmployee", expected.IsEmployee)
	formValues.Add("agreement", expected.Agreement)
	formValues.Add("phone", expected.Phone)
	formValues.Add("room", expected.Room)
	formValues.Add("comment", expected.Comment)
	formValues.Add("tariff", strconv.Itoa(expected.Tariff))

	require.HTTPRedirect(t, editUserPostHandler, "POST", "/edit-user", formValues)

	updatedUser, err := mysql.GetUserByID(id)
	require.NoError(t, err)

	assert.Equal(t, expected.Name, updatedUser.Name)
	assert.Equal(t, expected.IsEmployee, strconv.FormatBool(updatedUser.IsEmployee))
	assert.Equal(t, false, updatedUser.Paid)
	assert.Equal(t, expected.Agreement, updatedUser.Agreement)
	assert.Equal(t, expected.Phone, updatedUser.Phone)
	assert.Equal(t, expected.Room, updatedUser.Room)
	assert.Equal(t, expected.Comment, updatedUser.Comment)
	assert.Equal(t, expected.Tariff, updatedUser.Tariff.ID)
}

func TestSendMassSMSPostHandler(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sendMassSMSPostHandler(w, r)
	}))
	defer ts.Close()

	invalidJson := bytes.NewReader([]byte("{ \"message\": \"alo\", \"sdf\"::\"asdf\" }"))
	resp, err := http.Post(ts.URL+"/send-mass-sms", "application/json; charset=utf-8", invalidJson)
	require.Nil(t, err)
	assert.Equal(t, 500, resp.StatusCode)

	{
		var j struct {
			Message string `json:"message"`
			Phones  string `json:"phones"`
		}
		j.Message = "test"
		j.Phones = "89039496867,+79029995361"

		b, err := json.Marshal(&j)
		require.Nil(t, err)

		resp, err := http.Post(ts.URL+"/send-mass-sms", "application/json; charset=utf-8", bytes.NewReader(b))
		require.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	}
}

func TestPaymentPostHandler(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := createJWTtoken("login")
		require.NoError(t, err)
		cookie := http.Cookie{
			Name:     "jwt",
			Value:    token,
			HttpOnly: true,
			SameSite: 3,
		}
		r.AddCookie(&cookie)
		paymentPostHandler(w, r)
	}))
	defer ts.Close()

	user := User{
		Paid:      false,
		Name:      "Тестовый Тест Тестович100",
		Agreement: "П-009",
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
		UserID  int    `json:"id"`
		Sum     int    `json:"sum"`
		Receipt string `json:"receipt"`
		Admin   string `json:"admin"`
	}
	payment.UserID = userID
	payment.Sum = 100
	payment.Receipt = "№1001 от 27.09.2019"
	payment.Admin = "rozhkov"
	b, err := json.Marshal(&payment)
	require.NoError(t, err)

	resp, err := http.Post(ts.URL+"/payment", "application/json; charset=utf-8", bytes.NewReader(b))
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	actualUser, err := mysql.GetUserByID(userID)
	require.NoError(t, err)

	assert.Equal(t, payment.Sum, actualUser.Balance)
	assert.Equal(t, payment.Receipt, actualUser.Payments[len(actualUser.Payments)-1].Receipt)
	assert.Equal(t, payment.Admin, actualUser.Payments[len(actualUser.Payments)-1].Admin)
	assert.Equal(t, user.Paid, actualUser.Paid)

	payment.Receipt = "№1002 от 27.09.2019"
	b, err = json.Marshal(&payment)
	require.NoError(t, err)

	resp, err = http.Post(ts.URL+"/payment", "application/json; charset=utf-8", bytes.NewReader(b))
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	actualUser, err = mysql.GetUserByID(userID)
	require.NoError(t, err)

	assert.Equal(t, 0, actualUser.Balance)
	assert.Equal(t, payment.Receipt, actualUser.Payments[len(actualUser.Payments)-1].Receipt)
	assert.Equal(t, payment.Admin, actualUser.Payments[len(actualUser.Payments)-1].Admin)
	assert.Equal(t, true, actualUser.Paid)

	invalidJSON := []byte("{UserID: 10000, ReceiptID: 10000, Field true,}")
	resp, err = http.Post(ts.URL+"/payment", "application/json; charset=utf-8", bytes.NewBuffer(invalidJSON))
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	payment.UserID = 100000
	payment.Sum = 100000
	payment.Receipt = "№100000 от 1.1.2"
	payment.Admin = "error"
	b, err = json.Marshal(&payment)
	require.NoError(t, err)

	resp, err = http.Post(ts.URL+"/payment", "application/json; charset=utf-8", bytes.NewReader(b))
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Еще проверить записи в табличке payments
}

func TestGetStatsAboutUsers(t *testing.T) {
	require.HTTPSuccess(t, getStatsAboutUsers, "GET", "/stats", nil)

	var J struct {
		ActiveUsersCount   int `json:"active_users_count"`
		InactiveUsersCount int `json:"inactive_users_count"`
		ArchivedUsersCount int `json:"archived_users_count"`
		Cash               int `json:"cash"`
	}

	body := assert.HTTPBody(getStatsAboutUsers, "GET", "/stats", nil)
	err := json.NewDecoder(strings.NewReader(body)).Decode(&J)
	require.NoError(t, err)

	assert.NotZero(t, J.ActiveUsersCount)
	assert.NotZero(t, J.InactiveUsersCount)
	assert.NotZero(t, J.ArchivedUsersCount)
	assert.NotZero(t, J.Cash)
}

func TestGetIncomeForPeriodHandler(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		getIncomeForPeriodHandler(w, r)
	}))
	defer ts.Close()

	invalidJson := bytes.NewReader([]byte("{ \"from\": \"nottime\", \"to\"::\"2019.01.01\" }"))
	resp, err := http.Post(ts.URL+"/income-for-period", "application/json; charset=utf-8", invalidJson)
	require.Nil(t, err)
	assert.Equal(t, 500, resp.StatusCode)

	// Не обработаны успешные случаи
}

func TestGetNextAgreementHandler(t *testing.T) {
	require.HTTPSuccess(t, getNextAgreementHandler, "GET", "/next-agreement", nil)

	var J struct {
		Agreement string `json:"agreement"`
	}

	body := assert.HTTPBody(getNextAgreementHandler, "GET", "/next-agreement", nil)
	err := json.NewDecoder(strings.NewReader(body)).Decode(&J)
	require.NoError(t, err)

	assert.NotEmpty(t, J.Agreement)
}
