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
	ID        int64   `db:"id" json:"id"`
	ReceiptID int64   `db:"receipt_id" json:"receipt_id"`
	Name      string  `db:"name" json:"name"`
	Quantity  float64 `db:"quantity" json:"quantity"`
	Price     float64 `db:"price" json:"price"`
	UnitPrice float64 `db:"unit_price" json:"unit_price"`
}

type Receipt struct {
	ID          int64     `db:"id" json:"id"`
	UserID      int64     `db:"user_id" json:"user_id"`
	Supermarket string    `db:"supermarket" json:"supermarket"`
	Date        time.Time `db:"receipt_date" json:"date"`
	Total       float64   `db:"total" json:"total"`
	Currency    string    `db:"currency" json:"currency"`
	Items       []ReceiptItem
}

type User struct {
	ID        int64  `db:"id"`
	GoogleUID string `db:"google_uid"`
}

type Storage struct {
	db *sql.DB
}

func NewStorage() (*Storage, error) {
	db_name := os.Getenv("DB_NAME")
	if len(db_name) == 0 {
		return nil, errors.New("empty value for DB_NAME")
	}

	db_adapter := os.Getenv("DB_ADAPTER")
	if len(db_adapter) == 0 {
		return nil, errors.New("empty value for DB_NAME")
	}

	db, err := sql.Open("sqlite3", db_name)
	if err != nil {
		return nil, err
	}

	return &Storage{db: db}, nil
}

func (s *Storage) InitDB() error {

	const create = `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER NOT NULL PRIMARY KEY,
		google_uid varchar(255)
	  );
	  
	CREATE TABLE IF NOT EXISTS receipts (
		id INTEGER NOT NULL PRIMARY KEY,
		user_id int,
		supermarket varchar(255),
		receipt_date date,
    	currency varchar(3),
		total decimal(6, 2)
	);

	CREATE TABLE IF NOT EXISTS receipt_items (
		id INTEGER NOT NULL PRIMARY KEY,
		receipt_id int,
		name varchar(255),
		quantity float,
		price decimal(6, 2),
		unit_price decimal(6, 2)
	);`

	if _, err := s.db.Exec(create); err != nil {
		return err
	}

	return nil
}

// Find user by given ID and return User instance or error
func (s *Storage) FindUserById(user_id int) (*User, error) {
	row := s.db.QueryRow("SELECT * FROM users WHERE id = ?", user_id)

	user := User{}
	if err := row.Scan(&user.ID, &user.GoogleUID); err != nil {
		return nil, err
	}

	return &user, nil
}

// Check if given google id user exists in database
func (s *Storage) FindUserByGoogleUid(google_uid string) (*User, error) {
	row := s.db.QueryRow("SELECT * FROM users WHERE google_uid = ?", google_uid)

	user := User{}

	if err := row.Scan(&user.ID, &user.GoogleUID); err != nil {
		fmt.Println(err)
		return nil, err
	}
	return &user, nil
}

// Check if exists a receipt for given supermarket, date and amount (these values should be unique)
func (s *Storage) FindReceiptBySupermarketDateAmount(supermarket string, date time.Time, total float64) (*Receipt, error) {
	row := s.db.QueryRow("SELECT id, user_id, supermarket, receipt_date, currency, total FROM receipts WHERE supermarket LIKE ? AND DATE(receipt_date) = DATE(?) AND total = ?", fmt.Sprintf("%%%s%%", supermarket), date.Format(time.RFC3339), total)

	receipt := Receipt{}
	var currency sql.NullString
	err := row.Scan(&receipt.ID, &receipt.UserID, &receipt.Supermarket, &receipt.Date, &currency, &receipt.Total)

	if err != nil {
		return nil, err
	}

	receipt.Currency = currency.String
	return &receipt, nil
}

// Create a new receipt in the database and return record ID or error if could not be created
func (s *Storage) CreateReceipt(receipt *Receipt) (*Receipt, error) {
	// Check if receipt already exists
	ereceipt, err := s.FindReceiptBySupermarketDateAmount(receipt.Supermarket, receipt.Date, receipt.Total)

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if ereceipt != nil && ereceipt.UserID == receipt.UserID {
		return nil, errors.New("Receipt already exists")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	// Create receipt
	res, err := s.db.Exec("INSERT INTO receipts (user_id, supermarket, receipt_date, currency, total) VALUES (?, ?, ?, ?, ?)", receipt.UserID, receipt.Supermarket, receipt.Date.Format(time.RFC3339), receipt.Currency, receipt.Total)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	receipt.ID = id

	// Create receipt items
	for index, item := range receipt.Items {
		// Create receipt item
		res, err := s.db.Exec("INSERT INTO receipt_items (receipt_id, quantity, name, unit_price, price) VALUES (?, ?, ?, ?, ?)",
			id, item.Quantity, item.Name, item.UnitPrice, item.Price)

		if err != nil {
			return nil, err
		}

		item_id, err := res.LastInsertId()
		if err != nil {
			return nil, err
		}

		// Update item ID in receipt object
		receipt.Items[index].ID = item_id
	}

	tx.Commit()
	return receipt, nil
}
