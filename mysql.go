package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var db = newDB()

func newDB() *sql.DB {
	dsn := fmt.Sprintf("%v:%v@tcp(10.0.0.33)/billingdev?parseTime=true", os.Getenv("MYSQL_LOGIN"), os.Getenv("MYSQL_PASS"))
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	db.SetConnMaxLifetime(2 * time.Minute)
	db.SetMaxIdleConns(4)
	db.SetMaxOpenConns(8)

	return db
}

func userExistInDB(login string) bool {
	var num int
	err := db.QueryRow(`SELECT COUNT(*) FROM bl_users WHERE auth=?`, login).Scan(&num)
	if err != nil {
		return false
	}

	if num == 0 {
		return false
	}

	return true
}

func addMoney(id, money int) error {
	_, err := db.Exec(`UPDATE bl_users SET balance = balance + ? WHERE id=?`, money, id)
	if err != nil {
		return fmt.Errorf("could not update balance field: %v", err)
	}

	return nil
}

func addPaymentInfo(id, money int) error {
	_, err := db.Exec(`INSERT INTO bl_pays (userid, summa, pay_date, expired_date)
		VALUES (?,?,?,?)`, id, money, time.Now().Add(time.Hour*7), time.Now().AddDate(0, 1, 0).Add(time.Hour*7))
	if err != nil {
		return fmt.Errorf("could not insert payment info: %v", err)
	}

	return nil
}

func withdrawMoney(id int, callFromWeb bool) error {
	user, err := getUserByID(id)
	if err != nil {
		return fmt.Errorf("could not get user by id: %v", err)
	}

	if user.Money < user.Tariff.Price {
		return nil
	}

	if (callFromWeb && !user.Active) || (!callFromWeb) {
		t := time.Now().AddDate(0, 1, 0).Add(time.Hour * 7)
		_, err = db.Exec(`UPDATE bl_users SET expired_date=?, activity=1, balance=balance-? WHERE id=?`,
			t, user.Tariff.Price, id)
		if err != nil {
			return fmt.Errorf("could not update expired_date: %v", err)
		}

		go func() {
			time.Sleep(time.Until(t))
			withdrawMoney(id, false)
		}()
	}

	return nil
}

const (
	wired = iota + 1
	wireless
)

func updateUser(user User) error {
	typeConnect := wired
	if user.Tariff.ID == 3 {
		typeConnect = wireless
	}

	_, err := db.Exec(`UPDATE bl_users SET name=?, account=?, auth=?, tariff=?, phone=?, comment=?, type_connect_id=? WHERE id=?`,
		user.Name, user.Agreement, user.Login, user.Tariff.ID, user.Phone, user.Comment, typeConnect, user.ID)
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
	err = db.QueryRow(`SELECT id FROM bl_ipaddr WHERE ipaddr=?`, inIP).Scan(&inIPID)
	if err != nil {
		return 0, fmt.Errorf("could not get inIPID: %v", err)
	}

	var extIPID int
	err = db.QueryRow(`SELECT ext_ip_id FROM bl_external_ip WHERE ext_ip=?`, "82.200.46.10").Scan(&extIPID)
	if err != nil {
		return 0, fmt.Errorf("could not get extIPID: %v", err)
	}

	typeConnect := wired
	if user.Tariff.ID == 3 {
		typeConnect = wireless
	}

	res, err := db.Exec(`INSERT INTO bl_users (ip_id, ext_ip_id, tariff, balance, name, account, phone, auth, comment, type_connect_id)
		VALUES (?,?,?,?,?,?,?,?,?,?)`, inIPID, extIPID, user.Tariff.ID, user.Money, user.Name, user.Agreement, user.Phone, user.Login, user.Comment, typeConnect)
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
	err := db.QueryRow(`SELECT ipaddr FROM bl_ipaddr WHERE used = 0`).Scan(&inIP)
	if err != nil {
		return "", fmt.Errorf("could not get unusedInIP: %v", err)
	}

	return inIP, nil
}

func setInIPAsUsed(inIP string) error {
	_, err := db.Exec(`UPDATE bl_ipaddr SET used = 1 WHERE ipaddr=?`, inIP)
	if err != nil {
		return fmt.Errorf("could not set true used state to InIP: %v", err)
	}

	return nil
}

func getUsersByName(name string) ([]User, error) {
	sqlQuery := fmt.Sprintf(`SELECT bl_users.id, bl_users.name, bl_users.account, bl_users.auth,
		bl_users.balance, bl_users.activity, bl_users.phone, bl_users.comment,
		bl_users.expired_date, bl_ipaddr.ipaddr, bl_external_ip.ext_ip, bl_tariffs.tariff_id,
		bl_tariffs.tariff_name, bl_tariffs.tariff_summa
	FROM (((bl_users
		INNER JOIN bl_ipaddr ON bl_users.ip_id = bl_ipaddr.id)
		INNER JOIN bl_external_ip ON bl_users.ext_ip_id = bl_external_ip.ext_ip_id)
		INNER JOIN bl_tariffs ON bl_users.tariff = bl_tariffs.tariff_id)
	WHERE name LIKE '%%%v%%'`, name)
	rows, err := db.Query(sqlQuery)
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
		return db.Query(`SELECT bl_users.id, bl_users.name, bl_users.account, bl_users.auth,
			bl_users.balance, bl_users.activity, bl_users.phone, bl_users.comment,
			bl_users.expired_date, bl_ipaddr.ipaddr, bl_external_ip.ext_ip, bl_tariffs.tariff_id,
			bl_tariffs.tariff_name, bl_tariffs.tariff_summa
		FROM (((bl_users
			INNER JOIN bl_ipaddr ON bl_users.ip_id = bl_ipaddr.id)
			INNER JOIN bl_external_ip ON bl_users.ext_ip_id = bl_external_ip.ext_ip_id)
			INNER JOIN bl_tariffs ON bl_users.tariff = bl_tariffs.tariff_id)
		WHERE tariff = 1 OR tariff = 2`)
	}
	if t == "wireless" {
		return db.Query(`SELECT bl_users.id, bl_users.name, bl_users.account, bl_users.auth,
			bl_users.balance, bl_users.activity, bl_users.phone, bl_users.comment,
			bl_users.expired_date, bl_ipaddr.ipaddr, bl_external_ip.ext_ip, bl_tariffs.tariff_id,
			bl_tariffs.tariff_name, bl_tariffs.tariff_summa
		FROM (((bl_users
			INNER JOIN bl_ipaddr ON bl_users.ip_id = bl_ipaddr.id)
			INNER JOIN bl_external_ip ON bl_users.ext_ip_id = bl_external_ip.ext_ip_id)
			INNER JOIN bl_tariffs ON bl_users.tariff = bl_tariffs.tariff_id)
		WHERE tariff = 3`)
	}
	if t == "active" {
		return db.Query(`SELECT bl_users.id, bl_users.name, bl_users.account, bl_users.auth,
			bl_users.balance, bl_users.activity, bl_users.phone, bl_users.comment,
			bl_users.expired_date, bl_ipaddr.ipaddr, bl_external_ip.ext_ip, bl_tariffs.tariff_id,
			bl_tariffs.tariff_name, bl_tariffs.tariff_summa
		FROM (((bl_users
			INNER JOIN bl_ipaddr ON bl_users.ip_id = bl_ipaddr.id)
			INNER JOIN bl_external_ip ON bl_users.ext_ip_id = bl_external_ip.ext_ip_id)
			INNER JOIN bl_tariffs ON bl_users.tariff = bl_tariffs.tariff_id)
		WHERE activity = 1`)
	}
	if t == "inactive" {
		return db.Query(`SELECT bl_users.id, bl_users.name, bl_users.account, bl_users.auth,
			bl_users.balance, bl_users.activity, bl_users.phone, bl_users.comment,
			bl_users.expired_date, bl_ipaddr.ipaddr, bl_external_ip.ext_ip, bl_tariffs.tariff_id,
			bl_tariffs.tariff_name, bl_tariffs.tariff_summa
		FROM (((bl_users
			INNER JOIN bl_ipaddr ON bl_users.ip_id = bl_ipaddr.id)
			INNER JOIN bl_external_ip ON bl_users.ext_ip_id = bl_external_ip.ext_ip_id)
			INNER JOIN bl_tariffs ON bl_users.tariff = bl_tariffs.tariff_id)
		WHERE activity = 0`)
	}
	return db.Query(`SELECT bl_users.id, bl_users.name, bl_users.account, bl_users.auth,
		bl_users.balance, bl_users.activity, bl_users.phone, bl_users.comment,
		bl_users.expired_date, bl_ipaddr.ipaddr, bl_external_ip.ext_ip, bl_tariffs.tariff_id,
		bl_tariffs.tariff_name, bl_tariffs.tariff_summa
	FROM (((bl_users
		INNER JOIN bl_ipaddr ON bl_users.ip_id = bl_ipaddr.id)
		INNER JOIN bl_external_ip ON bl_users.ext_ip_id = bl_external_ip.ext_ip_id)
		INNER JOIN bl_tariffs ON bl_users.tariff = bl_tariffs.tariff_id)`)
}

func getUserByID(id int) (User, error) {
	var user User
	err := db.QueryRow(`SELECT bl_users.name, bl_users.account, bl_users.auth, bl_users.balance,
		bl_users.activity, bl_users.phone, bl_users.comment, bl_users.expired_date,
		bl_ipaddr.ipaddr, bl_external_ip.ext_ip, bl_tariffs.tariff_id, bl_tariffs.tariff_name,
		bl_tariffs.tariff_summa
	FROM (((bl_users
		INNER JOIN bl_ipaddr ON bl_users.ip_id = bl_ipaddr.id)
		INNER JOIN bl_external_ip ON bl_users.ext_ip_id = bl_external_ip.ext_ip_id)
		INNER JOIN bl_tariffs ON bl_users.tariff = bl_tariffs.tariff_id)
	WHERE bl_users.id = ?`, id).Scan(&user.Name, &user.Agreement, &user.Login, &user.Money, &user.Active, &user.Phone,
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
	err := db.QueryRow(`SELECT bl_users.id, bl_users.name, bl_users.account, bl_users.auth, bl_users.balance,
		bl_users.activity, bl_users.phone, bl_users.comment, bl_users.expired_date,
		bl_ipaddr.ipaddr, bl_external_ip.ext_ip, bl_tariffs.tariff_id, bl_tariffs.tariff_name,
		bl_tariffs.tariff_summa
	FROM (((bl_users
		INNER JOIN bl_ipaddr ON bl_users.ip_id = bl_ipaddr.id)
		INNER JOIN bl_external_ip ON bl_users.ext_ip_id = bl_external_ip.ext_ip_id)
		INNER JOIN bl_tariffs ON bl_users.tariff = bl_tariffs.tariff_id)
	WHERE bl_users.auth = ?`, login).Scan(&user.ID, &user.Name, &user.Agreement, &user.Login, &user.Money, &user.Active, &user.Phone,
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
	rows, err := db.Query(`SELECT summa, pay_date FROM bl_pays WHERE userid= ?`, id)
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
	err := db.QueryRow(`SELECT ip_id FROM bl_users WHERE id = ?`, id).Scan(&inIPID)
	if err != nil {
		return fmt.Errorf("could not get inIPID by user id: %v", err)
	}

	_, err = db.Exec(`UPDATE bl_ipaddr SET Used = 0 WHERE id=?`, inIPID)
	if err != nil {
		return fmt.Errorf("could not set false used state to ip_id: %v", err)
	}

	_, err = db.Exec(`DELETE FROM bl_users WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("could not delete user: %v", err)
	}

	_, err = db.Exec(`DELETE FROM bl_pays WHERE userid = ?`, id)
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
