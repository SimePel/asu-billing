package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUserByID(t *testing.T) {
	expectedUser := User{
		Activity:        false,
		Name:            "Тест",
		Agreement:       "П-777",
		Room:            "502",
		Phone:           "88005553550",
		Login:           "blabla.123",
		Balance:         0,
		ConnectionPlace: "Не важно",
		ExpiredDate:     time.Now().Add(time.Minute),
		Tariff: Tariff{
			ID: 1,
		},
	}
	id, err := AddUserToDB(expectedUser)
	require.NoError(t, err)

	actualUser, err := GetUserByID(id)
	require.NoError(t, err)
	assert.Equal(t, expectedUser.Activity, actualUser.Activity)
	assert.Equal(t, expectedUser.Name, actualUser.Name)
	assert.Equal(t, expectedUser.Agreement, actualUser.Agreement)
	assert.Equal(t, expectedUser.Room, actualUser.Room)
	assert.Equal(t, expectedUser.Phone, actualUser.Phone)
	assert.Equal(t, expectedUser.Login, actualUser.Login)
	assert.Equal(t, expectedUser.Balance, actualUser.Balance)
	assert.Equal(t, expectedUser.ConnectionPlace, actualUser.ConnectionPlace)
	assert.Equal(t, expectedUser.Tariff.ID, actualUser.Tariff.ID)
	assert.NotEmpty(t, actualUser.InnerIP)
	assert.NotEmpty(t, actualUser.ExtIP)
	assert.NotEmpty(t, actualUser.Tariff.Name)
	assert.NotEmpty(t, actualUser.Tariff.Price)

	err = DeleteUserByID(id)
	require.NoError(t, err)
}

func TestDBgetIDbyLogin(t *testing.T) {
	user := User{
		Activity:        false,
		Name:            "Тест",
		Agreement:       "П-777",
		Room:            "502",
		Phone:           "88005553550",
		Login:           "blabla.123",
		Balance:         0,
		ConnectionPlace: "Не важно",
		ExpiredDate:     time.Now().Add(time.Minute),
		Tariff: Tariff{
			ID: 1,
		},
	}
	expectedID, err := AddUserToDB(user)
	require.NoError(t, err)

	actualID, err := GetUserIDbyLogin(user.Login)
	require.NoError(t, err)
	assert.Equal(t, expectedID, int(actualID))

	err = DeleteUserByID(expectedID)
	require.NoError(t, err)
}
