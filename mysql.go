package main

import "fmt"

func getUserByID(id int) (User, error) {
	var user User
	err := db.QueryRow(`SELECT Users.Name, Users.Login, Users.Money, Users.Active, Users.Phone,
	 	Users.Comment, Users.Payments_ends, In_IPs.IP, Ext_IPs.IP, Tariffs.ID, Tariffs.Name, Tariffs.Price
	FROM (((Users
		INNER JOIN In_IPs ON Users.In_IP_ID = In_IPs.ID)
		INNER JOIN Ext_IPs ON Users.Ext_IP_ID = Ext_IPs.ID)
		INNER JOIN Tariffs ON Users.Tariff_ID = Tariffs.ID)
	WHERE Users.ID = ?`, id).Scan(&user.Name, &user.Login, &user.Money, &user.Active, &user.Phone,
		&user.Comment, &user.PaymentsEnds, &user.InIP, &user.ExtIP, &user.Tariff.ID, &user.Tariff.Name, &user.Tariff.Price)
	if err != nil {
		return user, fmt.Errorf("could not do queryRow: %v", err)
	}

	user.ID = id
	return user, nil
}
