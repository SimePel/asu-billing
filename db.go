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
	Activity  bool   `db:"activity" json:"activity"`
	Name      string `db:"name" json:"name"`
	Agreement string `db:"agreement" json:"agreement"`
	Phone     string `db:"phone" json:"phone"`
	Login     string `db:"login" json:"login"`
	Room      string `db:"room" json:"room"`
	Balance   string `db:"balance" json:"balance"`
	// separate for a more beautiful view
	ConnectionPlace string `db:"connection_place" json:"connection_place"`
}

func dbGetUser(userID string) *User {
	sess, err := mysql.Open(settings)
	if err != nil {
		log.Fatal("cannot open mysql session, ", err)
	}
	defer sess.Close()

	var user User
	usersColl := sess.Collection("users")
	err = usersColl.Find(userID).One(&user)
	if err != nil {
		log.Fatal("cannot get user, ", err)
	}

	return &user
}
