package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserCtx(t *testing.T) {
	r := chi.NewRouter()
	r.Route("/users", func(r chi.Router) {
		r.Route("/{userID}", func(r chi.Router) {
			r.Use(jsonContentType)
			r.Use(userCtx)
			r.Get("/", getUser)
		})
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	// Да - бред, но пока не понял, почему если добавлять юзера в самом тесте, то функция его не находит
	res, err := http.Get(ts.URL + "/users/1")
	require.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)

	var actualUser *User
	err = json.NewDecoder(res.Body).Decode(&actualUser)
	require.NoError(t, err)
	assert.Equal(t, 1, int(actualUser.ID))
}

func TestGetUser(t *testing.T) {
	req, err := http.NewRequest("GET", "/users/userID", nil)
	require.Nil(t, err)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		getUser(w, r)
	})

	expectedUser := &User{
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
	ctx = context.WithValue(ctx, userCtxKey("user"), expectedUser)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	var actualUser *User
	err = json.NewDecoder(rr.Body).Decode(&actualUser)
	require.Nil(t, err)

	assert.Equal(t, 200, rr.Code)
	assert.Equal(t, expectedUser, actualUser)
}
