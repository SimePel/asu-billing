package main

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
	"upper.io/db.v3/mysql"
)

func TestDBGetUser(t *testing.T) {
	sess, err := mysql.Open(settings)
	if err != nil {
		log.Fatal("cannot open mysql session, ", err)
	}
	defer sess.Close()

	expectedUser := &User{
		ID:              100,
		Activity:        false,
		Name:            "Тест",
		Agreement:       "П-777",
		Room:            "502",
		Phone:           "88005553550",
		Login:           "blabla.123",
		Balance:         0,
		ConnectionPlace: "Не важно",
	}

	_, err = sess.InsertInto("users").Values(expectedUser).Exec()
	require.NoError(t, err)

	actualUser, err := dbGetUser("100")
	require.NoError(t, err)
	assert.Equal(t, expectedUser, actualUser)

	_, err = sess.DeleteFrom("users").Where("id", 100).Exec()
	require.NoError(t, err)
}

func TestDBgetIDbyLogin(t *testing.T) {
	sess, err := mysql.Open(settings)
	if err != nil {
		log.Fatal("cannot open mysql session, ", err)
	}
	defer sess.Close()

	uploadUser := &User{
		ID:              101,
		Activity:        false,
		Name:            "Тест",
		Agreement:       "П-777",
		Room:            "502",
		Phone:           "88005553550",
		Login:           "blabla.321",
		Balance:         0,
		ConnectionPlace: "Не важно",
	}

	_, err = sess.InsertInto("users").Values(uploadUser).Exec()
	require.NoError(t, err)

	id, err := dbGetIDbyLogin(uploadUser.Login)
	require.NoError(t, err)
	assert.Equal(t, uploadUser.ID, id)

	_, err = sess.DeleteFrom("users").Where("id", 100).Exec()
	require.NoError(t, err)
}
