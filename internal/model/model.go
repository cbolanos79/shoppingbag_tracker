package model

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type ReceiptItem struct {
	ID        int64
	Name      string
	Quantity  int64
	Price     float64
	UnitPrice float64
}

type Receipt struct {
	ID          int64
	Supermarket string
	Date        time.Time
	Total       float64
	Items       []ReceiptItem
}

type User struct {
	ID        int64  `db:"id"`
	GoogleUID string `db:"google_uid"`
}

func NewDB() (*sql.DB, error) {
	db_name := os.Getenv("DB_NAME")
	if len(db_name) == 0 {
		return nil, errors.New("Empty value for DB_NAME")
	}

	db_adapter := os.Getenv("DB_ADAPTER")
	if len(db_adapter) == 0 {
		return nil, errors.New("Empty value for DB_NAME")
	}

	db, err := sql.Open("sqlite3", db_name)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func InitDB(db *sql.DB) error {

	const create = `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER NOT NULL PRIMARY KEY,
		google_uid varchar(255)
	  );
	  
	CREATE TABLE IF NOT EXISTS receipts (
		id INTEGER NOT NULL PRIMARY KEY,
		supermarket varchar(255),
		date date,
		total decimal(6, 2)
	) `

	if _, err := db.Exec(create); err != nil {
		return err
	}

	return nil
}

// Check if given google id user exists in database
func SearchUserByGoogleUid(db *sql.DB, google_uid string) (*User, error) {
	row := db.QueryRow("SELECT * FROM users WHERE google_uid = ?", google_uid)

	user := User{}

	if err := row.Scan(&user.ID, &user.GoogleUID); err != nil {
		fmt.Println(err)
		return nil, err
	}
	return &user, nil
}

// Check if exists a receipt for given supermarket, date and amount (these values should be unique)
func SearchReceiptBySupermarketDateAmount(db *sql.DB, supermarket string, date time.Time, total float64) (*Receipt, error) {
	row := db.QueryRow("SELECT * FROM receipts WHERE supermarket LIKE ? AND date = DATE(?) AND total = ?", fmt.Sprintf("%%%s%%", supermarket), date.Format(time.RFC3339), total)

	receipt := Receipt{}
	if err := row.Scan(&receipt.ID, &receipt.Supermarket, &receipt.Date, &receipt.Total); err != nil {
		return nil, err
	}

	return &receipt, nil
}
