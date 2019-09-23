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

	return db
}

// Payment struct
type Payment struct {
	Sum       int       `json:"sum"`
	Date      time.Time `json:"date"`
	ReceiptID int       `json:"receipt_id"`
}

// Tariff struct
type Tariff struct {
	ID    int    `json:"id"`
	Price int    `json:"price"`
	Name  string `json:"name"`
}

// User struct
type User struct {
	ID        uint      `json:"id"`
	Balance   int       `json:"balance"`
	Activity  bool      `json:"activity"`
	Paid      bool      `json:"paid"`
	Name      string    `json:"name"`
	Room      string    `json:"room"`
	Login     string    `json:"login"`
	Phone     string    `json:"phone"`
	Comment   string    `json:"comment"`
	ExtIP     string    `json:"ext_ip"`
	InnerIP   string    `json:"inner_ip"`
	Tariff    Tariff    `json:"tariff"`
	Payments  []Payment `json:"payments,omitempty"`
	Agreement string    `json:"agreement"`
	// separate for a more beautiful view
	ExpiredDate     time.Time `json:"expired_date"`
	ConnectionPlace string    `json:"connection_place"`
}

func (u User) hasEnoughMoneyForPayment() bool {
	if u.Balance < u.Tariff.Price {
		return false
	}

	return true
}

// GetAllUsers returns all users from db
func (mysql MySQL) GetAllUsers() ([]User, error) {
	rows, err := mysql.db.Query(`SELECT users.id, balance, users.name, login, agreement, expired_date,
		connection_place, activity, paid, room, comment, phone, tariffs.id AS tariff_id, tariffs.name,
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
			&user.ConnectionPlace, &user.Activity, &user.Paid, &user.Room, &user.Comment, &user.Phone,
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
		activity, paid, room, comment, phone, tariffs.id AS tariff_id, tariffs.name AS tariff_name, price, ips.ip, ext_ip  
	FROM (( users
		INNER JOIN ips ON users.ip_id = ips.id)
		INNER JOIN tariffs ON users.tariff = tariffs.id)
	WHERE users.id = ?`, id).Scan(&user.ID, &user.Balance, &user.Name, &user.Login, &user.Agreement, &user.ExpiredDate,
		&user.ConnectionPlace, &user.Activity, &user.Paid, &user.Room, &user.Comment, &user.Phone, &user.Tariff.ID,
		&user.Tariff.Name, &user.Tariff.Price, &user.InnerIP, &user.ExtIP)
	if err != nil {
		return user, fmt.Errorf("cannot get user with id=%v: %v", id, err)
	}

	payments, err := mysql.GetPaymentsByID(id)
	if err != nil {
		return user, fmt.Errorf("cannot get payments with id=%v: %v", id, err)
	}
	user.Payments = payments

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
	rows, err := mysql.db.Query(`SELECT receipt_id, sum, date FROM payments WHERE user_id= ?`, userID)
	if err != nil {
		return nil, fmt.Errorf("cannot get payments by id: %v", err)
	}

	var payment Payment
	payments := make([]Payment, 0)
	for rows.Next() {
		err := rows.Scan(&payment.ReceiptID, &payment.Sum, &payment.Date)
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

// AddUser adds user to db
func (mysql MySQL) AddUser(user User) (int, error) {
	innerIPid, err := mysql.getUnusedInnerIPid()
	if err != nil {
		return 0, fmt.Errorf("cannot get unused id of inner ip: %v", err)
	}

	agreement, err := mysql.getNextAgreement()
	if err != nil {
		return 0, fmt.Errorf("cannot get next agreement: %v", err)
	}

	res, err := mysql.db.Exec(`INSERT INTO users (balance, paid, name, room, comment, login, phone, ext_ip, ip_id,
		tariff, agreement, connection_place, expired_date) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		user.Balance, user.Paid, user.Name, user.Room, user.Comment, user.Login, user.Phone, "82.200.46.10",
		innerIPid, user.Tariff.ID, agreement, user.ConnectionPlace, user.ExpiredDate)
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

func (mysql MySQL) getNextAgreement() (string, error) {
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

func (mysql MySQL) UpdateUser(user User) error {
	_, err := mysql.db.Exec(`UPDATE users SET name=?, agreement=?, login=?, tariff=?, phone=?, room=?, comment=?,
	 	connection_place=? WHERE id=?`,
		user.Name, user.Agreement, user.Login, user.Tariff.ID, user.Phone, user.Room, user.Comment,
		user.ConnectionPlace, user.ID)
	if err != nil {
		return fmt.Errorf("cannot update user fields: %v", err)
	}

	return nil
}

// ProcessPayment updates balance and insert record into payments table
func (mysql MySQL) ProcessPayment(userID, receiptID, sum int) error {
	_, err := mysql.db.Exec(`UPDATE users SET balance=balance+? WHERE id=?`, sum, userID)
	if err != nil {
		return fmt.Errorf("cannot increase balance field: %v", err)
	}

	_, err = mysql.db.Exec(`INSERT INTO payments (user_id, receipt_id, sum, date) VALUES (?,?,?,?)`,
		userID, receiptID, sum, time.Now().Add(time.Hour*7))
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

func (mysql MySQL) ArchiveUserByID(id int) error {
	user, err := mysql.GetUserByID(id)
	if err != nil {
		return fmt.Errorf("cannot get user with id=%v: %v", id, err)
	}

	_, err = mysql.db.Exec(`INSERT INTO archive_users (id, name, agreement, login, phone, room, comment,
		connection_place, ip_id) VALUES (?,?,?,?,?,?,?,?,?)`, id, user.Name, user.Agreement, user.Login,
		user.Phone, user.Room, user.Comment, user.ConnectionPlace, user.InnerIP)
	if err != nil {
		return fmt.Errorf("cannot insert values in archive_users: %v", err)
	}

	_, err = mysql.db.Exec(`DELETE FROM users WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("cannot delete user with id=%v: %v", id, err)
	}

	return nil
}

func (mysql MySQL) GetCountOfActiveUsers() (int, error) {
	var count int
	err := mysql.db.QueryRow(`SELECT COUNT(*) FROM users WHERE activity=true`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot do queryRow: %v", err)
	}

	return count, nil
}

func (mysql MySQL) GetCountOfInactiveUsers() (int, error) {
	var count int
	err := mysql.db.QueryRow(`SELECT COUNT(*) FROM users WHERE activity=false`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot do queryRow: %v", err)
	}

	return count, nil
}

func (mysql MySQL) GetAllMoneyWeHave() (int, error) {
	var sum int
	err := mysql.db.QueryRow(`SELECT SUM(sum) FROM payments`).Scan(&sum)
	if err != nil {
		return 0, fmt.Errorf("cannot do queryRow: %v", err)
	}

	return sum, nil
}
