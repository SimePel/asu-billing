package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

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
	req, err := http.NewRequest("GET", "/users/", nil)
	require.Nil(t, err)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		getAllUsers(w, r)
	})

	mysql := MySQL{db: openTestDBconnection()}
	ctx := req.Context()
	ctx = context.WithValue(ctx, dbCtxKey("db"), mysql.db)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	var actualUsers []User
	err = json.NewDecoder(rr.Body).Decode(&actualUsers)
	require.Nil(t, err)

	assert.Equal(t, 200, rr.Code)
	assert.NotNil(t, actualUsers)
}

func TestDeleteUserHandler(t *testing.T) {
	req, err := http.NewRequest("DELETE", "/users/userID", nil)
	require.Nil(t, err)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deleteUser(w, r)
	})

	user := User{
		Name:  "Временно",
		Login: "Временно",
	}
	mysql := MySQL{db: openTestDBconnection()}
	id, err := mysql.AddUser(user)
	require.NoError(t, err)
	user.ID = uint(id)

	ctx := req.Context()
	ctx = context.WithValue(ctx, userCtxKey("user"), &user)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, 200, rr.Code)
}
