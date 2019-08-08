package main

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAllUsers(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	expectedUsers := []User{
		{
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
			ExpiredDate:     time.Now().Add(time.Minute).Local(),
			Tariff: Tariff{
				ID:    1,
				Name:  "Проводной",
				Price: 200,
			}},
		{
			ID:              2,
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

	rows := sqlmock.NewRows([]string{"id", "balance", "name", "login", "agreement", "expired_date",
		"connection_place", "activity", "room", "phone", "tariff_id", "tariff_name", "price", "ip", "ext_ip"}).
		AddRow(expectedUsers[0].ID, expectedUsers[0].Balance, expectedUsers[0].Name, expectedUsers[0].Login,
			expectedUsers[0].Agreement, expectedUsers[0].ExpiredDate, expectedUsers[0].ConnectionPlace,
			expectedUsers[0].Activity, expectedUsers[0].Room, expectedUsers[0].Phone, expectedUsers[0].Tariff.ID,
			expectedUsers[0].Tariff.Name, expectedUsers[0].Tariff.Price, expectedUsers[0].InnerIP, expectedUsers[0].ExtIP).
		AddRow(expectedUsers[1].ID, expectedUsers[1].Balance, expectedUsers[1].Name, expectedUsers[1].Login,
			expectedUsers[1].Agreement, expectedUsers[1].ExpiredDate, expectedUsers[1].ConnectionPlace,
			expectedUsers[1].Activity, expectedUsers[1].Room, expectedUsers[1].Phone, expectedUsers[1].Tariff.ID,
			expectedUsers[1].Tariff.Name, expectedUsers[1].Tariff.Price, expectedUsers[1].InnerIP, expectedUsers[1].ExtIP)

	mock.ExpectQuery("SELECT (.+) FROM (.+)").WillReturnRows(rows)
	actualUsers, err := GetAllUsers(db)
	require.NoError(t, err)
	assert.Equal(t, expectedUsers, actualUsers)
}

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

func TestAddUserToDB(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow("2")
	mock.ExpectQuery(`SELECT id FROM ips WHERE used = 0`).WillReturnRows(rows)

	res := sqlmock.NewResult(1, 1)
	mock.ExpectExec(`UPDATE ips SET used=1 WHERE id = `).WithArgs(2).WillReturnResult(res)

	mock.ExpectExec(`INSERT INTO users (.+) VALUES (.+)`).WillReturnResult(res)

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

	_, err = AddUserToDB(db, expectedUser)
	require.NoError(t, err)
}
