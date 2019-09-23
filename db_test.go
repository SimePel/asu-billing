package main

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func prepareDB(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
		name varchar(40) COLLATE utf8_unicode_ci NOT NULL,
		balance int(11) NOT NULL DEFAULT '0',
		agreement varchar(6) COLLATE utf8_unicode_ci NOT NULL,
		create_date timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		expired_date datetime NOT NULL DEFAULT '0000-00-00 00:00:00',
		login varchar(45) COLLATE utf8_unicode_ci NOT NULL,
		connection_place varchar(17) COLLATE utf8_unicode_ci NOT NULL,
		phone varchar(12) COLLATE utf8_unicode_ci NOT NULL,
		room varchar(14) COLLATE utf8_unicode_ci NOT NULL,
		paid tinyint(1) NOT NULL DEFAULT '0',
		activity tinyint(1) NOT NULL DEFAULT '0',
		tariff int(10) unsigned NOT NULL,
		ip_id int(10) unsigned NOT NULL,
		ext_ip varchar(15) COLLATE utf8_unicode_ci NOT NULL,
		PRIMARY KEY (id),
		UNIQUE (ip_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`ALTER TABLE users
		ADD FOREIGN KEY (ip_id) REFERENCES ips(id);`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS ips (
		id int(10) unsigned NOT NULL,
		ip varchar(16) COLLATE utf8_unicode_ci NOT NULL,
		used tinyint(1) NOT NULL DEFAULT '0',
		PRIMARY KEY (id),
		UNIQUE (ip)
	  ) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS tariffs (
		id int(10) unsigned NOT NULL,
		name varchar(20) COLLATE utf8_unicode_ci NOT NULL,
		price smallint(6) NOT NULL,
		PRIMARY KEY (id)
	  ) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS payments (
		id int(10) unsigned NOT NULL AUTO_INCREMENT,
		user_id bigint(20) unsigned NOT NULL,
		receipt_id mediumint(8) unsigned NOT NULL,
		sum smallint(6) NOT NULL,
		date timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		PRIMARY KEY (id),
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	  ) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`INSERT INTO ips (id, ip, used) VALUES
	(1, '10.1.108.1', 1),
	(2, '10.1.108.10', 1),
	(3, '10.1.108.100', 0),
	(4, '10.1.108.101', 0),
	(5, '10.1.108.102', 0),
	(6, '10.1.108.103', 0),
	(7, '10.1.108.104', 0),
	(8, '10.1.108.105', 0),
	(9, '10.1.108.106', 0),
	(10, '10.1.108.107', 0),
	(11, '10.1.108.108', 0),
	(12, '10.1.108.109', 0),
	(13, '10.1.108.11', 0),
	(14, '10.1.108.110', 0),
	(15, '10.1.108.111', 0),
	(16, '10.1.108.112', 0),
	(17, '10.1.108.113', 0),
	(18, '10.1.108.114', 0),
	(19, '10.1.108.115', 0),
	(20, '10.1.108.116', 0),
	(21, '10.1.108.117', 0),
	(22, '10.1.108.118', 0),
	(23, '10.1.108.119', 0),
	(24, '10.1.108.12', 0);`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`INSERT INTO users (id, name, balance, agreement, create_date, expired_date, login, 
		connection_place, phone, room, paid, activity, tariff, ip_id, ext_ip) VALUES (1, 'Тестовый Тест Тестович',
		100, 'П-001', '2019-06-11 05:49:05', '2019-06-27 04:25:26', 'blabla.123', '', '88005553550', '', 1, 1, 
		1, 1, '82.200.46.10'), (2, 'Тестовый Тест Тестович2', 300, 'П-002', '2019-08-12 07:46:35',
		'0000-00-00 00:00:00', 'bla.124', '', '', '501c', 0, 0, 1, 2, '82.200.46.10');`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`INSERT INTO payments (id, receipt_id, user_id, sum, date) VALUES (1, 001, 1, 200, '2019-06-07 07:32:50');`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`INSERT INTO tariffs (id, name, price) VALUES (1, 'Базовый-30', 200);`)
	if err != nil {
		return err
	}

	return nil
}

func clearDB(db *sql.DB) error {
	_, err := db.Exec(`SET FOREIGN_KEY_CHECKS = 0;`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`TRUNCATE TABLE ips;`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`TRUNCATE TABLE users;`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`TRUNCATE TABLE payments;`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`TRUNCATE TABLE tariffs;`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`SET FOREIGN_KEY_CHECKS = 1;`)
	if err != nil {
		return err
	}

	return nil
}

func openTestDBconnection() *sql.DB {
	dsn := fmt.Sprintf("%v:%v@tcp(10.0.0.33)/billingdev?parseTime=true", os.Getenv("MYSQL_LOGIN"), os.Getenv("MYSQL_PASS"))
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	return db
}

func TestMain(m *testing.M) {
	mysql := MySQL{db: openTestDBconnection()}

	err := clearDB(mysql.db)
	if err != nil {
		panic(err)
	}

	err = prepareDB(mysql.db)
	if err != nil {
		panic(err)
	}
	defer mysql.db.Close()

	os.Exit(m.Run())
}

func TestPing(t *testing.T) {
	mysql := MySQL{db: openTestDBconnection()}
	require.Nil(t, mysql.db.Ping())
}

func TestGetAllUsers(t *testing.T) {
	expectedUsers := []User{
		{
			ID:          1,
			Paid:        true,
			Activity:    true,
			Name:        "Тестовый Тест Тестович",
			Agreement:   "П-001",
			Phone:       "88005553550",
			Login:       "blabla.123",
			InnerIP:     "10.1.108.1",
			ExtIP:       "82.200.46.10",
			Balance:     100,
			ExpiredDate: time.Date(2019, time.June, 27, 4, 25, 26, 0, time.UTC),
			Tariff: Tariff{
				ID:    1,
				Name:  "Базовый-30",
				Price: 200,
			}},
		{
			ID:        2,
			Paid:      false,
			Name:      "Тестовый Тест Тестович2",
			Agreement: "П-002",
			Room:      "501c",
			Login:     "bla.124",
			InnerIP:   "10.1.108.10",
			ExtIP:     "82.200.46.10",
			Balance:   300,
			Tariff: Tariff{
				ID:    1,
				Name:  "Базовый-30",
				Price: 200,
			}},
	}

	mysql := MySQL{db: openTestDBconnection()}
	actualUsers, err := mysql.GetAllUsers()
	require.NoError(t, err)
	assert.Equal(t, expectedUsers, actualUsers)
}

func TestGetUserByID(t *testing.T) {
	expectedUser := User{
		ID:          1,
		Paid:        true,
		Activity:    true,
		Name:        "Тестовый Тест Тестович",
		Agreement:   "П-001",
		Phone:       "88005553550",
		Login:       "blabla.123",
		InnerIP:     "10.1.108.1",
		ExtIP:       "82.200.46.10",
		Balance:     100,
		ExpiredDate: time.Date(2019, time.June, 27, 4, 25, 26, 0, time.UTC),
		Payments: []Payment{
			Payment{
				ReceiptID: 1,
				Sum:       200,
				Date:      time.Date(2019, time.June, 7, 7, 32, 50, 0, time.UTC),
			},
		},
		Tariff: Tariff{
			ID:    1,
			Name:  "Базовый-30",
			Price: 200,
		},
	}

	mysql := MySQL{db: openTestDBconnection()}
	actualUser, err := mysql.GetUserByID(int(expectedUser.ID))
	require.NoError(t, err)
	assert.Equal(t, expectedUser, actualUser)
}

func TestGetUserIDbyLogin(t *testing.T) {
	user := User{
		ID:    1,
		Login: "blabla.123",
	}

	mysql := MySQL{db: openTestDBconnection()}
	actualID, err := mysql.GetUserIDbyLogin(user.Login)
	require.NoError(t, err)
	assert.Equal(t, user.ID, actualID)
}

func TestAddUser(t *testing.T) {
	expectedUser := User{
		Paid:  false,
		Name:  "Тестовый Тест Тестович3",
		Phone: "88005553553",
		Login: "baloga.154",
		ExtIP: "82.200.46.10",
		Tariff: Tariff{
			ID:    1,
			Name:  "Базовый-30",
			Price: 200,
		},
	}

	mysql := MySQL{db: openTestDBconnection()}
	actualID, err := mysql.AddUser(expectedUser)
	require.NoError(t, err)
	assert.Equal(t, 3, actualID)
}

func TestUpdateUser(t *testing.T) {
	mysql := MySQL{db: openTestDBconnection()}
	user, err := mysql.GetUserByID(2)
	require.NoError(t, err)

	user.Phone = "89993334455"
	user.ConnectionPlace = "рандом"

	err = mysql.UpdateUser(user)
	require.NoError(t, err)

	updatedUser, err := mysql.GetUserByID(2)
	require.NoError(t, err)

	assert.Equal(t, user, updatedUser)
}

func TestProcessPayment(t *testing.T) {
	mysql := MySQL{db: openTestDBconnection()}
	err := mysql.ProcessPayment(1, 2, 100)
	require.NoError(t, err)

	user, err := mysql.GetUserByID(1)
	require.NoError(t, err)

	assert.Equal(t, 200, user.Balance)
	assert.Equal(t, 2, user.Payments[len(user.Payments)-1].ReceiptID)
	// Еще протестить, что создалась запись в табличке payments
}

func TestGetPaymentsByID(t *testing.T) {
	user := User{
		Paid:    false,
		Name:    "Тестовый Тест Тестович6",
		Phone:   "88005553566",
		Login:   "payment.166",
		ExtIP:   "82.200.46.10",
		Balance: 0,
		Tariff: Tariff{
			ID:    1,
			Name:  "Базовый-30",
			Price: 200,
		},
	}

	mysql := MySQL{db: openTestDBconnection()}
	id, err := mysql.AddUser(user)
	require.NoError(t, err)
	user.ID = uint(id)

	err = mysql.ProcessPayment(id, 61, 200)
	require.NoError(t, err)

	err = mysql.ProcessPayment(id, 62, 100)
	require.NoError(t, err)

	actualPayments, err := mysql.GetPaymentsByID(id)
	require.NoError(t, err)

	assert.Equal(t, 2, len(actualPayments))
	assert.Equal(t, 200, actualPayments[0].Sum)
	assert.Equal(t, 61, actualPayments[0].ReceiptID)
	assert.Equal(t, 100, actualPayments[1].Sum)
	assert.Equal(t, 62, actualPayments[1].ReceiptID)
}

func TestPayForNextMonth(t *testing.T) {
	mysql := MySQL{db: openTestDBconnection()}
	user, err := mysql.GetUserByID(2)
	require.NoError(t, err)

	_, err = mysql.PayForNextMonth(user)
	require.NoError(t, err)

	actualUser, err := mysql.GetUserByID(2)
	require.NoError(t, err)

	expected := struct {
		ExpiredDate time.Time
		Paid        bool
		Balance     int
	}{
		time.Now().AddDate(0, 1, 0),
		true,
		100,
	}

	assert.Equal(t, expected.ExpiredDate.Format("2016.01.2 15:04:05"), actualUser.ExpiredDate.Format("2016.01.2 15:04:05"))
	assert.Equal(t, expected.Paid, actualUser.Paid)
	assert.Equal(t, expected.Balance, actualUser.Balance)
}

func TestDeleteUserByID(t *testing.T) {
	mysql := MySQL{db: openTestDBconnection()}
	user := User{
		Paid:  false,
		Name:  "Тестовый Тест Тестович4",
		Phone: "88005553441",
		Login: "unknown.444",
		ExtIP: "82.200.46.10",
		Tariff: Tariff{
			ID:    1,
			Name:  "Базовый-30",
			Price: 200,
		},
	}

	id, err := mysql.AddUser(user)
	require.NoError(t, err)

	err = mysql.ProcessPayment(id, 44, 200)
	require.NoError(t, err)

	err = mysql.DeleteUserByID(id)
	require.NoError(t, err)

	// Проверить, что и в табличке payments каскадно удалились записи
}

func TestGetCountOfActiveUsers(t *testing.T) {
	mysql := MySQL{db: openTestDBconnection()}
	count, err := mysql.GetCountOfActiveUsers()
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestGetCountOfInactiveUsers(t *testing.T) {
	mysql := MySQL{db: openTestDBconnection()}
	count, err := mysql.GetCountOfInactiveUsers()
	require.NoError(t, err)
	assert.NotZero(t, count)
}

func TestGetAllMoneyWeHave(t *testing.T) {
	mysql := MySQL{db: openTestDBconnection()}
	money, err := mysql.GetAllMoneyWeHave()
	require.NoError(t, err)
	assert.NotZero(t, money)
}
