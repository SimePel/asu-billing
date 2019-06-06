package main

import (
	"fmt"
	"log"
	"os"

	"upper.io/db.v3/mysql"
)

var settings = newSettings()

func newSettings() mysql.ConnectionURL {
	dsn := fmt.Sprintf("%v:%v@tcp(10.0.0.33)/billingdev?parseTime=true", os.Getenv("MYSQL_LOGIN"), os.Getenv("MYSQL_PASS"))
	settings, err := mysql.ParseURL(dsn)
	if err != nil {
		log.Fatal(err)
	}

	return settings
}

// User struct
type User struct {
	ID        uint   `db:"id,omitempty" json:"id"`
	Balance   int    `db:"balance" json:"balance"`
	Activity  bool   `db:"activity" json:"activity"`
	Name      string `db:"name" json:"name"`
	Room      string `db:"room" json:"room"`
	Login     string `db:"login" json:"login"`
	Phone     string `db:"phone" json:"phone"`
	Agreement string `db:"agreement" json:"agreement"`
	// separate for a more beautiful view
	ConnectionPlace string `db:"connection_place" json:"connection_place"`
}

func dbGetUser(userID string) (*User, error) {
	sess, err := mysql.Open(settings)
	if err != nil {
		return nil, fmt.Errorf("cannot open mysql session: %v", err)
	}
	defer sess.Close()

	var user User
	usersColl := sess.Collection("users")
	err = usersColl.Find().Where("id", userID).One(&user)
	if err != nil {
		return nil, fmt.Errorf("cannot get user by id: %v", err)
	}

	return &user, nil
}

func dbGetIDbyLogin(login string) (uint, error) {
	sess, err := mysql.Open(settings)
	if err != nil {
		return 0, fmt.Errorf("cannot open mysql session: %v", err)
	}
	defer sess.Close()

	var user struct {
		ID uint `db:"id" json:"id"`
	}
	err = sess.Collection("users").Find().Where("login", login).One(&user)
	if err != nil {
		return 0, fmt.Errorf("cannot get user id by login: %v", err)
	}

	return user.ID, nil
}
