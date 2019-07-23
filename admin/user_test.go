package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
