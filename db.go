package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// MySQL
type MySQL struct {
	db *sql.DB
}

func initializeDB() *sql.DB {
	dsn := fmt.Sprintf("%v:%v@tcp(10.0.0.33)/%v?parseTime=true", os.Getenv("MYSQL_LOGIN"),
		os.Getenv("MYSQL_PASS"), os.Getenv("DB_NAME"))
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	db.SetConnMaxLifetime(2 * time.Minute)
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(10)

	return db
}

// Payment struct
type Payment struct {
	Receipt string    `json:"receipt" csv:"Квитанция"`
	Method  string    `json:"method" csv:"Способ"`
	Admin   string    `json:"admin" csv:"Кто вносил"`
	Sum     int       `json:"sum" csv:"Сумма"`
	Date    time.Time `json:"date" csv:"Дата"`
}

// Operation struct
type Operation struct {
	Admin string    `json:"admin"`
	Type  string    `json:"type"`
	Date  time.Time `json:"date"`
}

// Tariff struct
type Tariff struct {
	ID    int    `json:"id"`
	Price int    `json:"price"`
	Name  string `json:"name"`
}

// User struct
type User struct {
	ID         uint        `json:"id"`
	Balance    int         `json:"balance"`
	Activity   bool        `json:"activity"`
	Paid       bool        `json:"paid"`
	Name       string      `json:"name"`
	Room       string      `json:"room"`
	Mac        string      `json:"mac"`
	Login      string      `json:"login"`
	Phone      string      `json:"phone"`
	Comment    string      `json:"comment"`
	ExtIP      string      `json:"ext_ip"`
	InnerIP    string      `json:"inner_ip"`
	Tariff     Tariff      `json:"tariff"`
	Payments   []Payment   `json:"payments,omitempty"`
	Operations []Operation `json:"operations,omitempty"`
	Agreement  string      `json:"agreement"`
	// separate for a more beautiful view
	IsDeactivated           bool      `json:"is_deactivated"`
	IsEmployee              bool      `json:"is_employee"`
	IsArchived              bool      `json:"is_archived"`
	IsLimited               bool      `json:"is_limited"`
	ExpiredDate             time.Time `json:"expired_date"`
	AgreementConclusionDate time.Time `json:"agreement_conclusion_date"`
	ConnectionPlace         string    `json:"connection_place"`
}

func (u User) hasEnoughMoneyForPayment() bool {
	return !(u.Balance < u.Tariff.Price)
}

// GetAllUsers returns all users from db
func (mysql MySQL) GetAllUsers() ([]User, error) {
	rows, err := mysql.db.Query(`SELECT users.id, balance, users.name, mac, login, agreement, expired_date,
		connection_place, activity, paid, comment, is_archived, phone, rooms.name, tariffs.id AS tariff_id, tariffs.name,
		price, ips.ip, ext_ip
	FROM ((( users
		INNER JOIN ips ON users.ip_id = ips.id)
		INNER JOIN rooms ON users.room_id = rooms.id)
		INNER JOIN tariffs ON users.tariff = tariffs.id)`)
	if err != nil {
		return nil, fmt.Errorf("cannot get all users: %v", err)
	}
	defer rows.Close()

	var user User
	users := make([]User, 0)
	for rows.Next() {
		err := rows.Scan(&user.ID, &user.Balance, &user.Name, &user.Mac, &user.Login, &user.Agreement, &user.ExpiredDate,
			&user.ConnectionPlace, &user.Activity, &user.Paid, &user.Comment, &user.IsArchived, &user.Phone, &user.Room,
			&user.Tariff.ID, &user.Tariff.Name, &user.Tariff.Price, &user.InnerIP, &user.ExtIP)
		if err != nil {
			return nil, fmt.Errorf("cannot get one row: %v", err)
		}
		users = append(users, user)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("something happened with rows: %v", err)
	}

	return users, nil
}

// GetUserByID returns user from db
func (mysql MySQL) GetUserByID(id int) (User, error) {
	var user User
	err := mysql.db.QueryRow(`SELECT users.id, balance, users.name, mac, login, agreement, expired_date,
		connection_place, activity, paid, comment, is_deactivated, is_employee, is_archived, is_limited, phone,
		agreement_conclusion_date, rooms.name, tariffs.id AS tariff_id, tariffs.name AS tariff_name, price, ips.ip, ext_ip
	FROM ((( users
		INNER JOIN ips ON users.ip_id = ips.id)
		INNER JOIN rooms ON users.room_id = rooms.id)
		INNER JOIN tariffs ON users.tariff = tariffs.id)
	WHERE users.id = ?`, id).Scan(&user.ID, &user.Balance, &user.Name, &user.Mac, &user.Login, &user.Agreement,
		&user.ExpiredDate, &user.ConnectionPlace, &user.Activity, &user.Paid, &user.Comment, &user.IsDeactivated,
		&user.IsEmployee, &user.IsArchived, &user.IsLimited, &user.Phone, &user.AgreementConclusionDate, &user.Room,
		&user.Tariff.ID, &user.Tariff.Name, &user.Tariff.Price, &user.InnerIP, &user.ExtIP)
	if err != nil {
		return user, fmt.Errorf("cannot get user with id=%v: %v", id, err)
	}

	payments, err := mysql.GetPaymentsByID(id)
	if err != nil {
		return user, fmt.Errorf("cannot get payments with id=%v: %v", id, err)
	}
	user.Payments = payments

	operations, err := mysql.GetOperationsByID(id)
	if err != nil {
		return user, fmt.Errorf("cannot get operations with user id=%v: %v", id, err)
	}
	user.Operations = operations

	return user, nil
}

// GetUserIDbyLogin returns user id from db
func (mysql MySQL) GetUserIDbyLogin(login string) (uint, error) {
	var id uint
	err := mysql.db.QueryRow(`SELECT id FROM users WHERE login = ?`, login).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("cannot get user id by login: %v", err)
	}

	return id, nil
}

// GetPaymentsByID returns info about user payments. Fresh records first
func (mysql MySQL) GetPaymentsByID(userID int) ([]Payment, error) {
	rows, err := mysql.db.Query(`SELECT admin, receipt, method, sum, date FROM payments WHERE user_id=? ORDER BY date DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("cannot get payments by id: %v", err)
	}

	var payment Payment
	payments := make([]Payment, 0)
	for rows.Next() {
		err := rows.Scan(&payment.Admin, &payment.Receipt, &payment.Method, &payment.Sum, &payment.Date)
		if err != nil {
			return nil, fmt.Errorf("cannot scan from row: %v", err)
		}
		payments = append(payments, payment)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("something happened with rows: %v", err)
	}

	return payments, nil
}

func (mysql MySQL) GetOperationsByID(userID int) ([]Operation, error) {
	rows, err := mysql.db.Query(`SELECT admin, type, date FROM operations WHERE user_id = ? ORDER BY date DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("cannot get operations by user id: %v", err)
	}

	var operation Operation
	operations := make([]Operation, 0)
	for rows.Next() {
		err := rows.Scan(&operation.Admin, &operation.Type, &operation.Date)
		if err != nil {
			return nil, fmt.Errorf("cannot scan from row: %v", err)
		}
		operations = append(operations, operation)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("something happened with rows: %v", err)
	}

	return operations, nil
}

// AddUser adds user to db
func (mysql MySQL) AddUser(user User) (int, error) {
	innerIPid, err := mysql.getUnusedInnerIPid()
	if err != nil {
		return 0, fmt.Errorf("cannot get unused id of inner ip: %v", err)
	}

	roomID, err := mysql.getRoomIDByName(user.Room)
	if err != nil {
		return 0, fmt.Errorf("cannot get room id: %v", err)
	}

	_, err = mysql.db.Exec(`UPDATE rooms SET vacant_esockets=vacant_esockets-1 WHERE id = ? AND vacant_esockets > 0`, roomID)
	if err != nil {
		return 0, fmt.Errorf("cannot decrease vacant ethernet sockets in the room - %v: %v", user.Room, err)
	}

	res, err := mysql.db.Exec(`INSERT INTO users (balance, paid, name, is_employee, comment, mac, login, phone, room_id,
		ext_ip, ip_id, agreement_conclusion_date, tariff, agreement, connection_place, expired_date)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, user.Balance, user.Paid, user.Name, user.IsEmployee, user.Comment,
		user.Mac, user.Login, user.Phone, roomID, "82.200.46.10", innerIPid, user.AgreementConclusionDate, user.Tariff.ID,
		user.Agreement, user.ConnectionPlace, user.ExpiredDate)
	if err != nil {
		return 0, fmt.Errorf("cannot insert values in db: %v", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("cannot get last insert id: %v", err)
	}

	return int(id), nil
}

func (mysql MySQL) getUnusedInnerIPid() (int, error) {
	var innerIPid int
	err := mysql.db.QueryRow(`SELECT id FROM ips WHERE used = 0`).Scan(&innerIPid)
	if err != nil {
		return 0, fmt.Errorf("cannot scan id from db to variable: %v", err)
	}

	_, err = mysql.db.Exec(`UPDATE ips SET used=1 WHERE id = ?`, innerIPid)
	if err != nil {
		return 0, fmt.Errorf("cannot set inner ip as used: %v", err)
	}

	return innerIPid, nil
}

func (mysql MySQL) getRoomIDByName(name string) (int, error) {
	var roomID int64
	err := mysql.db.QueryRow(`SELECT id FROM rooms WHERE name = ?`, name).Scan(&roomID)
	if err != nil {
		res, err := mysql.db.Exec(`INSERT INTO rooms (name) VALUES (?)`, name)
		if err != nil {
			return 0, fmt.Errorf("cannot add new room in the table: %v", err)
		}
		roomID, _ = res.LastInsertId()
	}

	return int(roomID), nil
}

func (mysql MySQL) FreePaymentForOneYear(id int) error {
	_, err := mysql.db.Exec(`UPDATE users SET paid=1, expired_date=? WHERE id=?`,
		time.Now().AddDate(1, 0, 0), id)
	if err != nil {
		return fmt.Errorf("cannot update values in db: %v", err)
	}

	return nil
}

func (mysql MySQL) ResetFreePaymentForOneYear(id int) error {
	_, err := mysql.db.Exec(`UPDATE users SET paid=0, expired_date=? WHERE id=?`,
		time.Now().AddDate(0, 0, -1), id)
	if err != nil {
		return fmt.Errorf("cannot update values in db: %v", err)
	}

	return nil
}

func (mysql MySQL) UpdateUser(user User) error {
	roomID, err := mysql.getRoomIDByName(user.Room)
	if err != nil {
		return fmt.Errorf("cannot get room id: %v", err)
	}

	_, err = mysql.db.Exec(`UPDATE users SET name=?, agreement=?, is_employee=?, mac=?, login=?, tariff=?, phone=?,
			room_id=?, comment=?, connection_place=?, agreement_conclusion_date=?, expired_date=? WHERE id=?`,
		user.Name, user.Agreement, user.IsEmployee, user.Mac, user.Login, user.Tariff.ID, user.Phone, roomID,
		user.Comment, user.ConnectionPlace, user.AgreementConclusionDate, user.ExpiredDate, user.ID)
	if err != nil {
		return fmt.Errorf("cannot update user fields: %v", err)
	}

	return nil
}

// ProcessPayment updates balance and insert record into payments table
func (mysql MySQL) ProcessPayment(userID, sum int, method, receipt, admin string) error {
	_, err := mysql.db.Exec(`UPDATE users SET balance=balance+? WHERE id=?`, sum, userID)
	if err != nil {
		return fmt.Errorf("cannot increase balance field: %v", err)
	}

	_, err = mysql.db.Exec(`INSERT INTO payments (user_id, admin, receipt, method, sum, date) VALUES (?,?,?,?,?,?)`,
		userID, admin, receipt, method, sum, time.Now())
	if err != nil {
		return fmt.Errorf("cannot insert record about payment: %v", err)
	}

	return nil
}

// PayForNextMonth activates user for next month
func (mysql MySQL) PayForNextMonth(user User) (time.Time, error) {
	t := time.Now().AddDate(0, 1, 0)
	_, err := mysql.db.Exec(`UPDATE users SET expired_date=?, paid=1, balance=balance-? WHERE id=?`,
		t, user.Tariff.Price, user.ID)
	if err != nil {
		return t, fmt.Errorf("cannot update user's info after payment: %v", err)
	}

	return t, nil
}

func (mysql MySQL) UnlimitUserByID(id int, admin string) error {
	_, err := mysql.db.Exec(`UPDATE users SET is_limited=0 WHERE id=?`, id)
	if err != nil {
		return fmt.Errorf("cannot unlimit user: %v", err)
	}

	_, err = mysql.db.Exec(`INSERT INTO operations (user_id, admin, type, date) VALUES (?,?,?,?)`, id, admin, "unlimit", time.Now())
	if err != nil {
		return fmt.Errorf("cannot add 'unlimit' operation: %v", err)
	}

	return nil
}

func (mysql MySQL) LimitUserByID(id int, admin string) error {
	_, err := mysql.db.Exec(`UPDATE users SET is_limited=1 WHERE id=?`, id)
	if err != nil {
		return fmt.Errorf("cannot limit user: %v", err)
	}

	_, err = mysql.db.Exec(`INSERT INTO operations (user_id, admin, type, date) VALUES (?,?,?,?)`, id, admin, "limit", time.Now())
	if err != nil {
		return fmt.Errorf("cannot add 'limit' operation: %v", err)
	}

	return nil
}

func (mysql MySQL) ActivateUserByID(id int, admin string) error {
	_, err := mysql.db.Exec(`UPDATE users SET is_deactivated=0 WHERE id=?`, id)
	if err != nil {
		return fmt.Errorf("cannot activate user: %v", err)
	}

	_, err = mysql.db.Exec(`INSERT INTO operations (user_id, admin, type, date) VALUES (?,?,?,?)`, id, admin, "activate", time.Now())
	if err != nil {
		return fmt.Errorf("cannot add 'activate' operation: %v", err)
	}

	user, err := mysql.GetUserByID(id)
	if err != nil {
		return fmt.Errorf("cannot get user by id: %v", err)
	}

	operations, err := mysql.GetOperationsByID(id)
	if err != nil {
		return fmt.Errorf("cannot get operations by user id: %v", err)
	}

	//	Берем предпоследнюю запись, так как последняя запись включения
	remainDurationForUser := user.ExpiredDate.Sub(operations[len(operations)-2].Date)
	_, err = mysql.db.Exec(`UPDATE users SET expired_date=? WHERE id=?`, operations[len(operations)-1].Date.Add(remainDurationForUser), id)
	if err != nil {
		return fmt.Errorf("cannot update expired_date: %v", err)
	}

	return nil
}

func (mysql MySQL) DeactivateUserByID(id int, admin string) error {
	_, err := mysql.db.Exec(`UPDATE users SET is_deactivated=1 WHERE id=?`, id)
	if err != nil {
		return fmt.Errorf("cannot deactivate user: %v", err)
	}

	_, err = mysql.db.Exec(`INSERT INTO operations (user_id, admin, type, date) VALUES (?,?,?,?)`, id, admin, "deactivate", time.Now())
	if err != nil {
		return fmt.Errorf("cannot add 'deactivate' operation: %v", err)
	}

	return nil
}

func (mysql MySQL) ArchiveUserByID(id int) error {
	_, err := mysql.db.Exec(`UPDATE users SET is_archived=1 WHERE id=?`, id)
	if err != nil {
		return fmt.Errorf("cannot archive user: %v", err)
	}

	_, err = mysql.db.Exec(`UPDATE rooms INNER JOIN users ON rooms.id = users.room_id
		SET vacant_esockets=vacant_esockets+1 WHERE users.id = ? AND connection_place != ''`, id)
	if err != nil {
		return fmt.Errorf("cannot restore vacant ethernet socket: %v", err)
	}

	return nil
}

func (mysql MySQL) RestoreUserByID(id int) error {
	_, err := mysql.db.Exec(`UPDATE users SET is_archived=0 WHERE id=?`, id)
	if err != nil {
		return fmt.Errorf("cannot restore user: %v", err)
	}

	_, err = mysql.db.Exec(`UPDATE rooms INNER JOIN users ON rooms.id = users.room_id
		SET vacant_esockets=vacant_esockets-1 WHERE users.id = ? AND connection_place != '' AND vacant_esockets > 0`, id)
	if err != nil {
		return fmt.Errorf("cannot decrease vacant ethernet socket: %v", err)
	}

	return nil
}

func (mysql MySQL) GetCountOfActiveUsers() (int, error) {
	var count int
	err := mysql.db.QueryRow(`SELECT COUNT(*) FROM users WHERE activity=true AND is_archived=false`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot do queryRow: %v", err)
	}

	return count, nil
}

func (mysql MySQL) GetCountOfInactiveUsers() (int, error) {
	var count int
	err := mysql.db.QueryRow(`SELECT COUNT(*) FROM users WHERE activity=false AND is_archived=false`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot do queryRow: %v", err)
	}

	return count, nil
}

func (mysql MySQL) GetCountOfArchivedUsers() (int, error) {
	var count int
	err := mysql.db.QueryRow(`SELECT COUNT(*) FROM users WHERE activity=false AND is_archived=true`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot do queryRow: %v", err)
	}

	return count, nil
}

func (mysql MySQL) GetAllMoneyWeHave() (int, error) {
	var sum int
	err := mysql.db.QueryRow(`SELECT SUM(money) FROM ( SELECT SUM(payments.sum) AS money FROM payments
		WHERE payments.user_id IN (SELECT id FROM users WHERE users.is_employee=0) ) AS a`).Scan(&sum)
	if err != nil {
		return 0, fmt.Errorf("cannot do queryRow: %v", err)
	}

	return sum, nil
}

func (mysql MySQL) GetIncomeForPeriod(from, to string) (int, error) {
	var sum int
	err := mysql.db.QueryRow(`SELECT SUM(money) FROM ( SELECT SUM(payments.sum) AS money FROM payments
		WHERE payments.user_id IN (SELECT id FROM users WHERE users.is_employee=0)
		AND payments.date >= ? AND payments.date < ? ) AS a`, from, to).Scan(&sum)
	if err != nil {
		return 0, fmt.Errorf("cannot do queryRow. From: %v, To: %v. Error: %v", from, to, err)
	}

	return sum, nil
}

type PaymentRecord struct {
	Name string `csv:"ФИО"`
	Payment
}

func (mysql MySQL) GetPaymentsRecords(from, to string) ([]PaymentRecord, error) {
	rows, err := mysql.db.Query(`SELECT users.name, receipt, method, sum, admin, date
		FROM ( users INNER JOIN payments ON users.id = payments.user_id )
		WHERE date >= ? AND date < ?  ORDER BY date DESC`, from, to)
	if err != nil {
		return nil, fmt.Errorf("cannot get payments records: %v", err)
	}

	var payment PaymentRecord
	payments := make([]PaymentRecord, 0)
	for rows.Next() {
		err := rows.Scan(&payment.Name, &payment.Receipt, &payment.Method, &payment.Sum, &payment.Admin, &payment.Date)
		if err != nil {
			return nil, fmt.Errorf("cannot scan from row: %v", err)
		}
		payments = append(payments, payment)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("something happened with rows: %v", err)
	}

	return payments, nil
}

// GetNextAgreement returns next agreement
func (mysql MySQL) GetNextAgreement() (string, error) {
	var lastAgreement string
	err := mysql.db.QueryRow(`SELECT MAX(agreement) FROM users`).Scan(&lastAgreement)
	if err != nil {
		return "", fmt.Errorf("cannot select max agreement: %v", err)
	}

	agreementParts := strings.Split(lastAgreement, "-")
	agreementID, err := strconv.Atoi(agreementParts[1])
	if err != nil {
		return "", fmt.Errorf("cannot convert agreement from string to int: %v", err)
	}

	agreementID += 1
	return fmt.Sprintf("%v-%03d", agreementParts[0], agreementID), nil
}
