package model

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type ReceiptItem struct {
	ID        int64   `db:"id"`
	ReceiptID int64   `db:"receipt_id"`
	Name      string  `db:"name"`
	Quantity  float64 `db:"quantity"`
	Price     float64 `db:"price"`
	UnitPrice float64 `db:"unit_price"`
}

type Receipt struct {
	ID          int64     `db:"id"`
	UserID      int64     `db:"user_id"`
	Supermarket string    `db:"supermarket"`
	Date        time.Time `db:"receipt_date"`
	Total       float64   `db:"total"`
	Currency    string    `db:"currency"`
	Items       []ReceiptItem
}

type ReceiptFilter struct {
	Supermarket string
	Page        int64
	PerPage     int64
	MinDate     *time.Time
	MaxDate     *time.Time
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

	if _, err := db.Exec(create); err != nil {
		return err
	}

	return nil
}

// Find user by given ID and return User instance or error
func FindUserById(db *sql.DB, user_id int) (*User, error) {
	row := db.QueryRow("SELECT * FROM users WHERE id = ?", user_id)

	user := User{}
	if err := row.Scan(&user.ID, &user.GoogleUID); err != nil {
		return nil, err
	}

	return &user, nil
}

// Check if given google id user exists in database
func FindUserByGoogleUid(db *sql.DB, google_uid string) (*User, error) {
	row := db.QueryRow("SELECT * FROM users WHERE google_uid = ?", google_uid)

	user := User{}

	if err := row.Scan(&user.ID, &user.GoogleUID); err != nil {
		log.Printf("FindUserByGoogleUid - Error scanning row for google_uid: %s\n%v", google_uid, err)
		return nil, fmt.Errorf("FindUserByGoogleUid - Error scanning row for google_uid: %s\n%v", google_uid, err)
	}
	return &user, nil
}

// Check if exists a receipt for given supermarket, date and amount (these values should be unique)
func FindReceiptBySupermarketDateAmount(db *sql.DB, supermarket string, date time.Time, total float64) (*Receipt, error) {
	row := db.QueryRow("SELECT id, user_id, supermarket, receipt_date, currency, total FROM receipts WHERE supermarket LIKE ? AND DATE(receipt_date) = DATE(?) AND total = ?", fmt.Sprintf("%%%s%%", supermarket), date.Format(time.RFC3339), total)

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
func CreateReceipt(db *sql.DB, receipt *Receipt) (*Receipt, error) {
	// Check if receipt already exists
	ereceipt, err := FindReceiptBySupermarketDateAmount(db, receipt.Supermarket, receipt.Date, receipt.Total)

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if ereceipt != nil && ereceipt.UserID == receipt.UserID {
		return nil, errors.New("Receipt already exists")
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	// Create receipt
	res, err := db.Exec("INSERT INTO receipts (user_id, supermarket, receipt_date, currency, total) VALUES (?, ?, ?, ?, ?)", receipt.UserID, receipt.Supermarket, receipt.Date.Format(time.RFC3339), receipt.Currency, receipt.Total)
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
		res, err := db.Exec("INSERT INTO receipt_items (receipt_id, quantity, name, unit_price, price) VALUES (?, ?, ?, ?, ?)",
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

func FindAllReceiptsForUser(db *sql.DB, user *User, filters *ReceiptFilter) (*[]Receipt, error) {
	var parameters []interface{}
	parameters = append(parameters, user.ID)

	sql := "SELECT id, supermarket, receipt_date, total FROM receipts WHERE user_id = ?"
	var limit, offset string

	// If there are filters apply the available ones
	if filters != nil {
		// Supermarket
		if len(filters.Supermarket) > 0 {
			parameters = append(parameters, fmt.Sprintf("%%%s%%", filters.Supermarket))
			sql = fmt.Sprintf("%s AND supermarket like ?", sql)
		}

		// Page and per page
		if filters.Page > 0 && filters.PerPage > 0 {
			limit = fmt.Sprintf("LIMIT %d", filters.PerPage)
			if filters.Page > 1 {
				offset = fmt.Sprintf("OFFSET %d", (filters.Page-1)*(filters.PerPage))
			}
		}

		// Date
		if filters.MinDate != nil {
			parameters = append(parameters, filters.MinDate)
			sql = fmt.Sprintf("%s AND DATE(receipt_date) >= DATE(?)", sql)

			// Set max date if present
			if filters.MaxDate != nil {

				// Avoid setting max date before min date
				if filters.MaxDate.Before(*filters.MinDate) {
					return nil, errors.New("MaxDate can not no lower than MinDate")
				}
				parameters = append(parameters, filters.MaxDate)
				sql = fmt.Sprintf("%s AND DATE(receipt_date) <= DATE(?)", sql)
			}
		}
	}

	sql = fmt.Sprintf("%s ORDER BY receipt_date DESC %s %s", sql, limit, offset)
	rows, err := db.Query(sql, parameters...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var receipts []Receipt

	for rows.Next() {
		receipt := Receipt{}
		rows.Scan(&receipt.ID, &receipt.Supermarket, &receipt.Date, &receipt.Total)
		receipts = append(receipts, receipt)
	}

	return &receipts, nil
}

func FindReceiptForUser(db *sql.DB, receipt_id int, user_id int) (*Receipt, error) {
	// Get receipt information filtering by given user
	row := db.QueryRow("SELECT id, supermarket, receipt_date, currency, total FROM receipts WHERE id = ? AND user_id = ?", receipt_id, user_id)

	receipt := Receipt{}

	var currency sql.NullString

	err := row.Scan(&receipt.ID, &receipt.Supermarket, &receipt.Date, &currency, &receipt.Total)
	receipt.Currency = currency.String

	if err != nil {
		return nil, err
	}

	// Get receipt items
	rows, err := db.Query("SELECT id, quantity, name, unit_price, price FROM receipt_items WHERE receipt_id = ? ORDER BY quantity DESC", receipt_id)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		item := ReceiptItem{}
		rows.Scan(&item.ID, &item.Quantity, &item.Name, &item.UnitPrice, &item.Price)
		receipt.Items = append(receipt.Items, item)
	}
	return &receipt, nil

}

/*
	// Page and per page
	if filters.Page > 0 && filters.PerPage > 0 {
		limit = fmt.Sprintf("LIMIT %d", filters.PerPage)
		if filters.Page > 1 {
			offset = fmt.Sprintf("OFFSET %d", (filters.Page-1)*(filters.PerPage))
		}
	}
*/
