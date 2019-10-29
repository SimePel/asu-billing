package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTryToRenewPayment(t *testing.T) {
	user := User{
		Paid:      false,
		Name:      "Тестовый Тест Тестович126",
		Agreement: "П-010",
		Phone:     "88005553126",
		Login:     "renew.payment",
		Balance:   300,
		Tariff: Tariff{
			ID:    1,
			Price: 200,
		},
	}

	mysql := MySQL{db: openTestDBconnection()}
	id, err := mysql.AddUser(user)
	require.NoError(t, err)
	user.ID = uint(id)

	tryToRenewPayment(mysql, int(user.ID))

	updatedUser, err := mysql.GetUserByID(id)
	require.NoError(t, err)

	assert.Equal(t, user.Balance-200, updatedUser.Balance)
	assert.Equal(t, !user.Paid, updatedUser.Paid)

	tryToRenewPayment(mysql, int(user.ID))
}

func TestSendNotification(t *testing.T) {
	user := User{
		Name:      "Тестовый Тест Тестович18",
		Agreement: "П-180",
		Phone:     "89039496867",
		Login:     "checkNotification",
		Balance:   400,
		Tariff: Tariff{
			ID: 1,
		},
	}

	mysql := MySQL{db: openTestDBconnection()}
	actualID, err := mysql.AddUser(user)
	require.NoError(t, err)

	user.ID = uint(actualID)
	_, err = mysql.PayForNextMonth(user)
	require.NoError(t, err)

	// sms will not be sent, because user is able to pay for the next month
	err = sendNotification(mysql, actualID)
	require.NoError(t, err)

	// Нужно доделать тест
}

func TestSendSMS(t *testing.T) {
	phone := "89039496867"
	message := "Test message"
	err := sendSMS(phone, message)
	require.NoError(t, err)
}
