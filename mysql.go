package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"os/exec"
	"time"
)

// Не возвращаю ip в состояние used=false специально
func turnOffInactiveUsers() error {
	ips, err := getOverdueIPs()
	if err != nil {
		return fmt.Errorf("could not get overdue ips: %v", err)
	}
	if ips == nil {
		return nil
	}

	for _, ip := range ips {
		_, err = db.Exec(`UPDATE Users INNER JOIN In_IPs ON Users.In_IP_ID = In_IPs.ID
		 	SET Active=0, Payments_ends=0 WHERE IP=?`, ip)
		if err != nil {
			return fmt.Errorf("could not update user active status: %v", err)
		}
	}

	err = removeUsersFromRouter(ips)
	if err != nil {
		return fmt.Errorf("could not block users ips on router: %v", err)
	}

	return nil
}

func getOverdueIPs() (ips []string, err error) {
	rows, err := db.Query(`SELECT IP FROM (Users
		INNER JOIN In_IPs ON Users.In_IP_ID = In_IPs.ID)
	WHERE NOW() > Payments_ends AND Active=1`)
	if err != nil {
		return nil, fmt.Errorf("could not get overdue ips: %v", err)
	}
	defer rows.Close()

	var ip string
	for rows.Next() {
		err := rows.Scan(&ip)
		if err != nil {
			return nil, fmt.Errorf("could not scan from row: %v", err)
		}
		ips = append(ips, ip)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("something happened with rows: %v", err)
	}

	return ips, nil
}

func removeUsersFromRouter(ips []string) error {
	for _, ip := range ips {
		expectCMD := exec.Command("echo", ip)
		var out bytes.Buffer
		expectCMD.Stdout = &out
		err := expectCMD.Run()
		if err != nil {
			return fmt.Errorf("could not run remove-expect cmd: %v", err)
		}

		fmt.Printf("%q", out)
	}

	return nil
}

func addMoney(id, money int) error {
	_, err := db.Exec(`UPDATE Users SET Money = Money + ? WHERE ID=?`, money, id)
	if err != nil {
		return fmt.Errorf("could not update money field: %v", err)
	}

	return nil
}

func addPaymentInfo(id, money int) error {
	_, err := db.Exec(`INSERT INTO Payments (User_ID, Amount, Date)
		VALUES (?,?,?)`, id, money, time.Now())
	if err != nil {
		return fmt.Errorf("could not insert payment info: %v", err)
	}

	return nil
}

func withdrawMoney(id int) error {
	user, err := getUserByID(id)
	if err != nil {
		return fmt.Errorf("could not get user by id: %v", err)
	}

	if user.Money < user.Tariff.Price {
		return nil
	}

	months := user.Money / user.Tariff.Price

	paymentsEnds := user.PaymentsEnds.AddDate(0, months, 0)
	if !user.Active {
		paymentsEnds = time.Now().AddDate(0, months, 0)

		err = addUserIPToRouter(user.InIP)
		if err != nil {
			return fmt.Errorf("could not permit user's ip on router: %v", err)
		}
	}

	_, err = db.Exec(`UPDATE Users SET Payments_ends=?, Active=1, Money=? WHERE ID=?`,
		paymentsEnds, user.Money-(user.Tariff.Price*months), id)
	if err != nil {
		return fmt.Errorf("could not update payments_ends: %v", err)
	}

	return nil
}

func addUserIPToRouter(ip string) error {
	expectCMD := exec.Command("echo", "add "+ip)
	var out bytes.Buffer
	expectCMD.Stdout = &out
	err := expectCMD.Run()
	if err != nil {
		return fmt.Errorf("could not run add-expect cmd: %v", err)
	}

	fmt.Printf("%q", out)
	return nil
}

func updateUser(user User) error {
	_, err := db.Exec(`UPDATE Users SET Name=?, Agreement=?, Login=?, Tariff_ID=?, Phone=?, Comment=? WHERE ID=?`,
		user.Name, user.Agreement, user.Login, user.Tariff.ID, user.Phone, user.Comment, user.ID)
	if err != nil {
		return fmt.Errorf("could not update user fields: %v", err)
	}

	return nil
}

func addUserToDB(user User) (int, error) {
	inIP, err := getUnusedInIP()
	if err != nil {
		return 0, fmt.Errorf("could not get unused inIP: %v", err)
	}

	err = setInIPAsUsed(inIP)
	if err != nil {
		return 0, fmt.Errorf("could not set InIP as used: %v", err)
	}

	var inIPID int
	err = db.QueryRow(`SELECT ID FROM In_IPs WHERE IP=?`, inIP).Scan(&inIPID)
	if err != nil {
		return 0, fmt.Errorf("could not get inIPID: %v", err)
	}

	var extIPID int
	err = db.QueryRow(`SELECT ID FROM Ext_IPs WHERE IP=?`, "82.200.46.10").Scan(&extIPID)
	if err != nil {
		return 0, fmt.Errorf("could not get extIPID: %v", err)
	}

	res, err := db.Exec(`INSERT INTO Users (In_IP_ID, Ext_IP_ID, Tariff_ID, Money, Name, Agreement, Phone, Login, Comment)
		VALUES (?,?,?,?,?,?,?,?,?)`, inIPID, extIPID, user.Tariff.ID, user.Money, user.Name, user.Agreement, user.Phone, user.Login, user.Comment)
	if err != nil {
		return 0, fmt.Errorf("could not insert user: %v", err)
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("could not get lastInsertID: %v", err)
	}

	return int(lastID), nil
}

func getUnusedInIP() (string, error) {
	var inIP string
	err := db.QueryRow(`SELECT IP FROM In_IPs WHERE Used = 0`).Scan(&inIP)
	if err != nil {
		return "", fmt.Errorf("could not get unusedInIP: %v", err)
	}

	return inIP, nil
}

func setInIPAsUsed(inIP string) error {
	_, err := db.Exec(`UPDATE In_IPs SET Used = 1 WHERE IP=?`, inIP)
	if err != nil {
		return fmt.Errorf("could not set true used state to InIP: %v", err)
	}

	return nil
}

func getUsersByType(t string) ([]User, error) {
	rows, err := getRowsByType(t)
	if err != nil {
		return nil, fmt.Errorf("could not do query: %v", err)
	}
	defer rows.Close()

	var user User
	users := make([]User, 0)
	for rows.Next() {
		err := rows.Scan(&user.ID, &user.Name, &user.Agreement, &user.Login, &user.Money, &user.Active, &user.Phone, &user.Comment,
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

func getRowsByType(t string) (*sql.Rows, error) {
	if t == "wired" {
		return db.Query(`SELECT Users.ID, Users.Name, Users.Agreement, Users.Login, Users.Money, Users.Active, Users.Phone,
		Users.Comment, Users.Payments_ends, In_IPs.IP, Ext_IPs.IP, Tariffs.ID, Tariffs.Name, Tariffs.Price
		FROM (((Users
			INNER JOIN In_IPs ON Users.In_IP_ID = In_IPs.ID)
			INNER JOIN Ext_IPs ON Users.Ext_IP_ID = Ext_IPs.ID)
			INNER JOIN Tariffs ON Users.Tariff_ID = Tariffs.ID)
		WHERE Tariff_ID = 1 OR Tariff_ID = 2`)
	}
	if t == "wireless" {
		return db.Query(`SELECT Users.ID, Users.Name, Users.Agreement, Users.Login, Users.Money, Users.Active, Users.Phone,
		Users.Comment, Users.Payments_ends, In_IPs.IP, Ext_IPs.IP, Tariffs.ID, Tariffs.Name, Tariffs.Price
		FROM (((Users
			INNER JOIN In_IPs ON Users.In_IP_ID = In_IPs.ID)
			INNER JOIN Ext_IPs ON Users.Ext_IP_ID = Ext_IPs.ID)
			INNER JOIN Tariffs ON Users.Tariff_ID = Tariffs.ID)
		WHERE Tariff_ID = 3`)
	}
	if t == "active" {
		return db.Query(`SELECT Users.ID, Users.Name, Users.Agreement, Users.Login, Users.Money, Users.Active, Users.Phone,
		Users.Comment, Users.Payments_ends, In_IPs.IP, Ext_IPs.IP, Tariffs.ID, Tariffs.Name, Tariffs.Price
		FROM (((Users
			INNER JOIN In_IPs ON Users.In_IP_ID = In_IPs.ID)
			INNER JOIN Ext_IPs ON Users.Ext_IP_ID = Ext_IPs.ID)
			INNER JOIN Tariffs ON Users.Tariff_ID = Tariffs.ID)
		WHERE Active = 1`)
	}
	if t == "inactive" {
		return db.Query(`SELECT Users.ID, Users.Name, Users.Agreement, Users.Login, Users.Money, Users.Active, Users.Phone,
		Users.Comment, Users.Payments_ends, In_IPs.IP, Ext_IPs.IP, Tariffs.ID, Tariffs.Name, Tariffs.Price
		FROM (((Users
			INNER JOIN In_IPs ON Users.In_IP_ID = In_IPs.ID)
			INNER JOIN Ext_IPs ON Users.Ext_IP_ID = Ext_IPs.ID)
			INNER JOIN Tariffs ON Users.Tariff_ID = Tariffs.ID)
		WHERE Active = 0`)
	}
	return db.Query(`SELECT Users.ID, Users.Name, Users.Agreement, Users.Login, Users.Money, Users.Active, Users.Phone,
	Users.Comment, Users.Payments_ends, In_IPs.IP, Ext_IPs.IP, Tariffs.ID, Tariffs.Name, Tariffs.Price
	FROM (((Users
		INNER JOIN In_IPs ON Users.In_IP_ID = In_IPs.ID)
		INNER JOIN Ext_IPs ON Users.Ext_IP_ID = Ext_IPs.ID)
		INNER JOIN Tariffs ON Users.Tariff_ID = Tariffs.ID)`)
}

func getUserByID(id int) (User, error) {
	var user User
	err := db.QueryRow(`SELECT Users.Name, Users.Agreement, Users.Login, Users.Money, Users.Active, Users.Phone,
	 	Users.Comment, Users.Payments_ends, In_IPs.IP, Ext_IPs.IP, Tariffs.ID, Tariffs.Name, Tariffs.Price
	FROM (((Users
		INNER JOIN In_IPs ON Users.In_IP_ID = In_IPs.ID)
		INNER JOIN Ext_IPs ON Users.Ext_IP_ID = Ext_IPs.ID)
		INNER JOIN Tariffs ON Users.Tariff_ID = Tariffs.ID)
	WHERE Users.ID = ?`, id).Scan(&user.Name, &user.Agreement, &user.Login, &user.Money, &user.Active, &user.Phone,
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

func getUserByLogin(login string) (User, error) {
	var user User
	err := db.QueryRow(`SELECT Users.ID, Users.Name, Users.Agreement, Users.Login, Users.Money, Users.Active, Users.Phone,
		Users.Comment, Users.Payments_ends, In_IPs.IP, Ext_IPs.IP, Tariffs.ID, Tariffs.Name, Tariffs.Price
	FROM (((Users
		INNER JOIN In_IPs ON Users.In_IP_ID = In_IPs.ID)
		INNER JOIN Ext_IPs ON Users.Ext_IP_ID = Ext_IPs.ID)
		INNER JOIN Tariffs ON Users.Tariff_ID = Tariffs.ID)
		WHERE Users.Login = ?`, login).Scan(&user.ID, &user.Name, &user.Agreement, &user.Login, &user.Money, &user.Active, &user.Phone,
		&user.Comment, &user.PaymentsEnds, &user.InIP, &user.ExtIP, &user.Tariff.ID, &user.Tariff.Name, &user.Tariff.Price)
	if err != nil {
		return user, fmt.Errorf("could not do queryRow: %v", err)
	}

	payments, err := getPaymentsByID(user.ID)
	if err != nil {
		return user, fmt.Errorf("could not get payments with id=%v: %v", user.ID, err)
	}
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

	_, err = db.Exec(`UPDATE In_IPs SET Used = 0 WHERE ID=?`, inIPID)
	if err != nil {
		return fmt.Errorf("could not set false used state to In_IP_ID: %v", err)
	}

	_, err = db.Exec(`DELETE FROM Users WHERE ID = ?`, id)
	if err != nil {
		return fmt.Errorf("could not delete user: %v", err)
	}

	_, err = db.Exec(`DELETE FROM Payments WHERE User_ID = ?`, id)
	if err != nil {
		return fmt.Errorf("could not delete payments info: %v", err)
	}

	return nil
}

func formatTime(t time.Time) string {
	if t.Unix() < 0 {
		return ""
	}
	return t.Format("2.01.2006")
}
