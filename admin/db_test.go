package main

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUserByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	expectedUser := User{
		ID:              1,
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

	rows := sqlmock.NewRows([]string{"id", "balance", "name", "login", "agreement", "expired_date",
		"connection_place", "activity", "room", "phone", "tariff_id", "tariff_name", "price", "ip", "ext_ip"}).
		AddRow(expectedUser.ID, expectedUser.Balance, expectedUser.Name, expectedUser.Login,
			expectedUser.Agreement, expectedUser.ExpiredDate, expectedUser.ConnectionPlace,
			expectedUser.Activity, expectedUser.Room, expectedUser.Phone, expectedUser.Tariff.ID,
			expectedUser.Tariff.Name, expectedUser.Tariff.Price, expectedUser.InnerIP, expectedUser.ExtIP)

	mock.ExpectQuery("SELECT (.+) FROM (.+) WHERE users.id = ").WithArgs(expectedUser.ID).WillReturnRows(rows)

	actualUser, err := GetUserByID(db, int(expectedUser.ID))
	require.NoError(t, err)
	assert.Equal(t, expectedUser, actualUser)
}

func TestDBgetIDbyLogin(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	user := User{
		ID:    1,
		Login: "blabla.123",
	}

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(user.ID)

	mock.ExpectQuery("SELECT id FROM users WHERE login = ").WithArgs(user.Login).WillReturnRows(rows)

	actualID, err := GetUserIDbyLogin(db, user.Login)
	require.NoError(t, err)
	assert.Equal(t, user.ID, actualID)
}
