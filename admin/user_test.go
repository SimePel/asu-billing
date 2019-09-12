package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserCtx(t *testing.T) {
	mysql := MySQL{db: openTestDBconnection()}
	user := User{
		Paid:      false,
		Activity:  false,
		Name:      "Тестовый Тест Тестович701",
		Agreement: "П-701",
		Room:      "701а",
		Login:     "king.701",
		Balance:   0,
		Tariff: Tariff{
			ID:    1,
			Name:  "Проводной",
			Price: 200,
		},
	}
	expectedID, err := mysql.AddUser(user)
	require.NoError(t, err)

	r := chi.NewRouter()
	r.Route("/users", func(r chi.Router) {
		r.Route("/{userID}", func(r chi.Router) {
			r.Use(jsonContentType)
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					ctx := context.WithValue(r.Context(), dbCtxKey("db"), mysql.db)
					next.ServeHTTP(w, r.WithContext(ctx))
				})
			})
			r.Use(userCtx)
			r.Get("/", getUser)
		})
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	res, err := http.Get(ts.URL + fmt.Sprintf("/users/%v", expectedID))
	require.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)

	var actualUser *User
	err = json.NewDecoder(res.Body).Decode(&actualUser)
	require.NoError(t, err)
	assert.Equal(t, expectedID, int(actualUser.ID))
}

func TestGetUser(t *testing.T) {
	req, err := http.NewRequest("GET", "/users/userID", nil)
	require.Nil(t, err)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		getUser(w, r)
	})

	user := &User{
		ID:              100,
		Paid:            false,
		Activity:        false,
		Name:            "Тест",
		Agreement:       "П-777",
		Room:            "502",
		Phone:           "88005553550",
		Login:           "blabla.123@stud.asu.ru",
		Balance:         0,
		ConnectionPlace: "Не важно",
	}
	ctx := req.Context()
	ctx = context.WithValue(ctx, userCtxKey("user"), user)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	var actualUser *User
	err = json.NewDecoder(rr.Body).Decode(&actualUser)
	require.Nil(t, err)

	assert.Equal(t, 200, rr.Code)
	assert.Equal(t, user, actualUser)
}

func TestGetAllUsersHandler(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	expectedUsers := []User{
		{
			ID:              1,
			Paid:            false,
			Activity:        false,
			Name:            "Тест",
			Agreement:       "П-777",
			Room:            "502",
			Phone:           "88005553550",
			Login:           "blabla.123",
			InnerIP:         "10.80.80.1",
			ExtIP:           "82.200.46.10",
			Balance:         0,
			ConnectionPlace: "Не важно",
			ExpiredDate:     time.Now().Add(time.Minute).Local(),
			Tariff: Tariff{
				ID:    1,
				Name:  "Проводной",
				Price: 200,
			}},
		{
			ID:              2,
			Paid:            false,
			Activity:        false,
			Name:            "Тест2",
			Agreement:       "П-123",
			Room:            "501",
			Phone:           "88005553551",
			Login:           "bla.124",
			InnerIP:         "10.80.80.2",
			ExtIP:           "82.200.46.10",
			Balance:         0,
			ConnectionPlace: "Не важно",
			ExpiredDate:     time.Now().Add(time.Minute).Local(),
			Tariff: Tariff{
				ID:    1,
				Name:  "Проводной",
				Price: 200,
			}},
	}

	rows := sqlmock.NewRows([]string{"id", "balance", "name", "login", "agreement", "expired_date", "connection_place",
		"paid", "activity", "room", "phone", "tariff_id", "tariff_name", "price", "ip", "ext_ip"}).
		AddRow(expectedUsers[0].ID, expectedUsers[0].Balance, expectedUsers[0].Name, expectedUsers[0].Login,
			expectedUsers[0].Agreement, expectedUsers[0].ExpiredDate, expectedUsers[0].ConnectionPlace,
			expectedUsers[0].Paid, expectedUsers[0].Activity, expectedUsers[0].Room, expectedUsers[0].Phone,
			expectedUsers[0].Tariff.ID, expectedUsers[0].Tariff.Name, expectedUsers[0].Tariff.Price,
			expectedUsers[0].InnerIP, expectedUsers[0].ExtIP).
		AddRow(expectedUsers[1].ID, expectedUsers[1].Balance, expectedUsers[1].Name, expectedUsers[1].Login,
			expectedUsers[1].Agreement, expectedUsers[1].ExpiredDate, expectedUsers[1].ConnectionPlace,
			expectedUsers[1].Paid, expectedUsers[1].Activity, expectedUsers[1].Room, expectedUsers[1].Phone,
			expectedUsers[1].Tariff.ID, expectedUsers[1].Tariff.Name, expectedUsers[1].Tariff.Price,
			expectedUsers[1].InnerIP, expectedUsers[1].ExtIP)

	mock.ExpectQuery("SELECT (.+) FROM (.+)").WillReturnRows(rows)

	req, err := http.NewRequest("GET", "/users/", nil)
	require.Nil(t, err)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		getAllUsers(w, r)
	})

	ctx := req.Context()
	ctx = context.WithValue(ctx, dbCtxKey("db"), db)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	var actualUsers []User
	err = json.NewDecoder(rr.Body).Decode(&actualUsers)
	require.Nil(t, err)

	assert.Equal(t, 200, rr.Code)
	assert.Equal(t, expectedUsers, actualUsers)
}
