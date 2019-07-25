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
	var expectedID uint = 1
	r := chi.NewRouter()
	r.Route("/users", func(r chi.Router) {
		r.Route("/{userID}", func(r chi.Router) {
			r.Use(jsonContentType)
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					db, mock, err := sqlmock.New()
					require.NoError(t, err)

					user := User{
						ID:              expectedID,
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
						ExpiredDate:     time.Now().Add(time.Minute),
						Tariff: Tariff{
							ID:    1,
							Name:  "Проводной",
							Price: 200,
						},
					}
					rows := sqlmock.NewRows([]string{"id", "balance", "name", "login", "agreement",
						"expired_date", "connection_place", "activity", "room", "phone", "tariff_id",
						"tariff_name", "price", "ip", "ext_ip"}).AddRow(user.ID, user.Balance,
						user.Name, user.Login, user.Agreement, user.ExpiredDate,
						user.ConnectionPlace, user.Activity, user.Room, user.Phone,
						user.Tariff.ID, user.Tariff.Name, user.Tariff.Price, user.InnerIP, user.ExtIP)
					mock.ExpectQuery(`SELECT (.+) FROM (.+) WHERE users.id = `).WithArgs(expectedID).WillReturnRows(rows)
					ctx := context.WithValue(r.Context(), dbCtxKey("db"), db)
					next.ServeHTTP(w, r.WithContext(ctx))
				})
			})
			r.Use(userCtx)
			r.Get("/", getUser)
		})
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	// Да - бред, но пока не понял, почему если добавлять юзера в самом тесте, то функция его не находит
	res, err := http.Get(ts.URL + fmt.Sprintf("/users/%v", expectedID))
	require.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)

	var actualUser *User
	err = json.NewDecoder(res.Body).Decode(&actualUser)
	require.NoError(t, err)
	assert.Equal(t, expectedID, actualUser.ID)
}

func TestGetUser(t *testing.T) {
	req, err := http.NewRequest("GET", "/users/userID", nil)
	require.Nil(t, err)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		getUser(w, r)
	})

	user := &User{
		ID:              100,
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
