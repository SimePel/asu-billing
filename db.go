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
	Sum     int       `json:"sum"`
	Date    time.Time `json:"date"`
	Receipt string    `json:"receipt"`
	Admin   string    `json:"admin"`
}

// Operation struct
type Operation struct {
	Type string    `json:"type"`
	Date time.Time `json:"date"`
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
	IsDeactivated   bool      `json:"is_deactivated"`
	IsEmployee      bool      `json:"is_employee"`
	IsArchived      bool      `json:"is_archived"`
	ExpiredDate     time.Time `json:"expired_date"`
	ConnectionPlace string    `json:"connection_place"`
}

func (u User) hasEnoughMoneyForPayment() bool {
	return !(u.Balance < u.Tariff.Price)
}

// GetAllUsers returns all users from db
func (mysql MySQL) GetAllUsers() ([]User, error) {
	rows, err := mysql.db.Query(`SELECT users.id, balance, users.name, login, agreement, expired_date,
		connection_place, activity, paid, room, comment, is_archived, phone, tariffs.id AS tariff_id, tariffs.name,
		price, ips.ip, ext_ip
	FROM (( users
		INNER JOIN ips ON users.ip_id = ips.id)
		INNER JOIN tariffs ON users.tariff = tariffs.id)`)
	if err != nil {
		return nil, fmt.Errorf("cannot get all users: %v", err)
	}
	defer rows.Close()

	var user User
	users := make([]User, 0)
	for rows.Next() {
		err := rows.Scan(&user.ID, &user.Balance, &user.Name, &user.Login, &user.Agreement, &user.ExpiredDate,
			&user.ConnectionPlace, &user.Activity, &user.Paid, &user.Room, &user.Comment, &user.IsArchived, &user.Phone,
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
	err := mysql.db.QueryRow(`SELECT users.id, balance, users.name, login, agreement, expired_date, connection_place,
		activity, paid, room, comment, is_deactivated, is_employee, is_archived, phone, tariffs.id AS tariff_id,
		tariffs.name AS tariff_name, price, ips.ip, ext_ip  
	FROM (( users
		INNER JOIN ips ON users.ip_id = ips.id)
		INNER JOIN tariffs ON users.tariff = tariffs.id)
	WHERE users.id = ?`, id).Scan(&user.ID, &user.Balance, &user.Name, &user.Login, &user.Agreement, &user.ExpiredDate,
		&user.ConnectionPlace, &user.Activity, &user.Paid, &user.Room, &user.Comment, &user.IsDeactivated,
		&user.IsEmployee, &user.IsArchived, &user.Phone, &user.Tariff.ID, &user.Tariff.Name, &user.Tariff.Price,
		&user.InnerIP, &user.ExtIP)
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

// GetPaymentsByID returns info about user payments
func (mysql MySQL) GetPaymentsByID(userID int) ([]Payment, error) {
	rows, err := mysql.db.Query(`SELECT admin, receipt, sum, date FROM payments WHERE user_id= ?`, userID)
	if err != nil {
		return nil, fmt.Errorf("cannot get payments by id: %v", err)
	}

	var payment Payment
	payments := make([]Payment, 0)
	for rows.Next() {
		err := rows.Scan(&payment.Admin, &payment.Receipt, &payment.Sum, &payment.Date)
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
	rows, err := mysql.db.Query(`SELECT type, date FROM operations WHERE user_id = ?`, userID)
	if err != nil {
		return nil, fmt.Errorf("cannot get operations by user id: %v", err)
	}

	var operation Operation
	operations := make([]Operation, 0)
	for rows.Next() {
		err := rows.Scan(&operation.Type, &operation.Date)
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

	res, err := mysql.db.Exec(`INSERT INTO users (balance, paid, name, is_employee, room, comment, login, phone,
		ext_ip, ip_id, tariff, agreement, connection_place, expired_date) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		user.Balance, user.Paid, user.Name, user.IsEmployee, user.Room, user.Comment, user.Login, user.Phone, "82.200.46.10",
		innerIPid, user.Tariff.ID, user.Agreement, user.ConnectionPlace, user.ExpiredDate)
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
	_, err := mysql.db.Exec(`UPDATE users SET name=?, agreement=?, is_employee=?, login=?, tariff=?, phone=?, room=?,
	 		comment=?, connection_place=?, expired_date=? WHERE id=?`,
		user.Name, user.Agreement, user.IsEmployee, user.Login, user.Tariff.ID, user.Phone, user.Room, user.Comment,
		user.ConnectionPlace, user.ExpiredDate, user.ID)
	if err != nil {
		return fmt.Errorf("cannot update user fields: %v", err)
	}

	return nil
}

// ProcessPayment updates balance and insert record into payments table
func (mysql MySQL) ProcessPayment(userID, sum int, receipt, admin string) error {
	_, err := mysql.db.Exec(`UPDATE users SET balance=balance+? WHERE id=?`, sum, userID)
	if err != nil {
		return fmt.Errorf("cannot increase balance field: %v", err)
	}

	_, err = mysql.db.Exec(`INSERT INTO payments (user_id, admin, receipt, sum, date) VALUES (?,?,?,?,?)`,
		userID, admin, receipt, sum, time.Now().Add(time.Hour*7))
	if err != nil {
		return fmt.Errorf("cannot insert record about payment: %v", err)
	}

	return nil
}

// PayForNextMonth activates user for next month
func (mysql MySQL) PayForNextMonth(user User) (time.Time, error) {
	t := time.Now().AddDate(0, 1, 0).Add(time.Hour * 7)
	_, err := mysql.db.Exec(`UPDATE users SET expired_date=?, paid=1, balance=balance-? WHERE id=?`,
		t, user.Tariff.Price, user.ID)
	if err != nil {
		return t, fmt.Errorf("cannot update user's info after payment: %v", err)
	}

	return t, nil
}

func (mysql MySQL) ActivateUserByID(id int) error {
	_, err := mysql.db.Exec(`UPDATE users SET is_deactivated=0 WHERE id=?`, id)
	if err != nil {
		return fmt.Errorf("cannot activate user: %v", err)
	}

	_, err = mysql.db.Exec(`INSERT INTO operations (user_id, type, date) VALUES (?,?,?)`, id, "activate", time.Now().Add(time.Hour*7))
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

func (mysql MySQL) DeactivateUserByID(id int) error {
	_, err := mysql.db.Exec(`UPDATE users SET is_deactivated=1 WHERE id=?`, id)
	if err != nil {
		return fmt.Errorf("cannot deactivate user: %v", err)
	}

	_, err = mysql.db.Exec(`INSERT INTO operations (user_id, type, date) VALUES (?,?,?)`, id, "deactivate", time.Now().Add(time.Hour*7))
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

	return nil
}

func (mysql MySQL) RestoreUserByID(id int) error {
	_, err := mysql.db.Exec(`UPDATE users SET is_archived=0 WHERE id=?`, id)
	if err != nil {
		return fmt.Errorf("cannot restore user: %v", err)
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

func (mysql MySQL) GetAllMoneyWeHave() (int, error) {
	var sum int
	err := mysql.db.QueryRow(`SELECT SUM(money) FROM ( SELECT SUM(payments.sum) AS money FROM payments
	 	WHERE payments.user_id IN (SELECT id FROM users WHERE users.is_employee=0) UNION
		SELECT SUM(users.balance) FROM users WHERE users.is_employee=0) AS a`).Scan(&sum)
	if err != nil {
		return 0, fmt.Errorf("cannot do queryRow: %v", err)
	}

	return sum, nil
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
