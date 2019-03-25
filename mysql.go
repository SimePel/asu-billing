package main

import (
	"fmt"
)

func getUsersByType(t string) ([]User, error) {
	rows, err := db.Query(`SELECT Users.ID, Users.Name, Users.Login, Users.Money, Users.Active, Users.Phone,
		Users.Comment, Users.Payments_ends, In_IPs.IP, Ext_IPs.IP, Tariffs.ID, Tariffs.Name, Tariffs.Price
	FROM (((Users
		INNER JOIN In_IPs ON Users.In_IP_ID = In_IPs.ID)
		INNER JOIN Ext_IPs ON Users.Ext_IP_ID = Ext_IPs.ID)
		INNER JOIN Tariffs ON Users.Tariff_ID = Tariffs.ID)`)
	if err != nil {
		return nil, fmt.Errorf("could not do query: %v", err)
	}
	defer rows.Close()

	var user User
	users := make([]User, 0)
	for rows.Next() {
		err := rows.Scan(&user.ID, &user.Name, &user.Login, &user.Money, &user.Active, &user.Phone, &user.Comment,
			&user.PaymentsEnds, &user.InIP, &user.ExtIP, &user.Tariff.ID, &user.Tariff.Name, &user.Tariff.Price)
		if err != nil {
			return nil, fmt.Errorf("could not scan from row: %v", err)
		}
		users = append(users, user)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("something happened with rows: %v", err)
	}

	return users, nil
}

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

	payments, err := getPaymentsByID(id)
	if err != nil {
		return user, fmt.Errorf("could not get payments with id=%v: %v", id, err)
	}
	user.ID = id
	user.Payments = payments

	return user, nil
}

func getPaymentsByID(id int) ([]Payment, error) {
	rows, err := db.Query(`SELECT Amount, Date FROM Payments WHERE User_ID= ?`, id)
	if err != nil {
		return nil, fmt.Errorf("could not get payments by id: %v", err)
	}

	var payment Payment
	payments := make([]Payment, 0)
	for rows.Next() {
		err := rows.Scan(&payment.Amount, &payment.Last)
		if err != nil {
			return nil, fmt.Errorf("could not scan from row: %v", err)
		}
		payments = append(payments, payment)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("something happened with rows: %v", err)
	}

	return payments, nil
}

func deleteUserByID(id int) error {
	var inIPID string
	err := db.QueryRow(`SELECT In_IP_ID FROM Users WHERE ID = ?`, id).Scan(&inIPID)
	if err != nil {
		return fmt.Errorf("could not get inIPID by user id: %v", err)
	}

	_, err = db.Exec(`UPDATE In_IPs SET used = 0 WHERE ID=?`, inIPID)
	if err != nil {
		return fmt.Errorf("could not set false used state to In_IP_ID: %v", err)
	}

	_, err = db.Exec(`DELETE FROM Users WHERE ID = ?`, id)
	if err != nil {
		return fmt.Errorf("could not delete user: %v", err)
	}

	return nil
}
