package model

import (
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestUserFindByIdNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Unexpected error %s connecting to database", err)
	}

	defer db.Close()
	mock.ExpectQuery("SELECT \\* FROM users WHERE id = ?").
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

	user, _ := FindUserById(db, 2)

	if user != nil {
		t.Fatalf("Error: user should be nil for non existing id")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestUserFindByIdSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Unexpected error %s connecting to database", err)
	}

	defer db.Close()
	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "User 1")

	mock.ExpectQuery("SELECT \\* FROM users WHERE id = ?").
		WithArgs(1).
		WillReturnRows(rows)

	user, err := FindUserById(db, 1)

	if user == nil {
		t.Fatalf("Error: user should not be nil for non existing id")
	}

	if err != nil && err != sql.ErrNoRows {
		t.Fatalf("Unexpected error: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestFindUserByGoogleIdNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Unexpected error %s connecting to database", err)
	}

	defer db.Close()

	rows := mock.NewRows([]string{"id", "google_uid"})

	mock.ExpectQuery("SELECT \\* FROM users WHERE google_uid = ?").
		WithArgs("12345").
		WillReturnRows(rows)

	user, err := FindUserByGoogleUid(db, "12345")
	if user != nil {
		t.Fatalf("User should be nil for non existing google_uid")
	}

	if err != nil && err != sql.ErrNoRows {
		t.Fatalf("Unexpected error: %s", err)
	}

}

func TestFindUserByGoogleIdFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Unexpected error %s connecting to database", err)
	}

	defer db.Close()

	rows := mock.NewRows([]string{"id", "google_uid"}).AddRow(1, "12345")

	mock.ExpectQuery("SELECT \\* FROM users WHERE google_uid = ?").
		WithArgs("12345").
		WillReturnRows(rows)

	user, err := FindUserByGoogleUid(db, "12345")
	if user == nil {
		t.Fatalf("User should not be nil for existing google_uid")
	}

	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

}

func TestFindReceiptBySupermarketDateAmountNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Unexpected error %s connecting to database", err)
	}

	defer db.Close()
	ts := time.Now()

	rows := mock.NewRows([]string{"id", "supermarket", "date", "currency", "total"})

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, supermarket, receipt_date, currency, total FROM receipts WHERE supermarket LIKE ? AND DATE(receipt_date) = DATE(?) AND total = ?")).
		WithArgs("%other%", ts.Format(time.RFC3339), 543.21).
		WillReturnRows(rows)

	receipt, err := FindReceiptBySupermarketDateAmount(db, "other", ts, 543.21)

	if receipt != nil {
		t.Fatalf("Receipt should not be nil for not existing params")
	}

	if err != nil && err != sql.ErrNoRows {
		t.Fatalf("Unexpected error: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestFindReceiptBySupermarketDateAmountFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Unexpected error %s connecting to database", err)
	}

	defer db.Close()
	ts := time.Now()

	rows := mock.NewRows([]string{"id", "user_id", "supermarket", "date", "currency", "total"}).
		AddRow(1, 1, "Any", ts, "EUR", 123.45)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, supermarket, receipt_date, currency, total FROM receipts WHERE supermarket LIKE ? AND DATE(receipt_date) = DATE(?) AND total = ?")).
		WithArgs("%Any%", ts.Format(time.RFC3339), 123.45).
		WillReturnRows(rows)

	receipt, err := FindReceiptBySupermarketDateAmount(db, "Any", ts, 123.45)

	if err != nil && err != sql.ErrNoRows {
		t.Fatalf("Unexpected error: %s", err)
	}

	if receipt == nil {
		t.Fatalf("Receipt should not be nil for existing params")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestCreateDuplicatedReceipt(t *testing.T) {
	db, mock, err := sqlmock.New()

	if err != nil {
		t.Fatalf("Unexpected error %s connecting to database", err)
	}

	defer db.Close()

	ts := time.Now()

	//receipt := Receipt{UserID: 1, Supermarket: "Any", Date: ts, Total: 100.0}

	receipt_rows := mock.NewRows([]string{"id", "user_id", "supermarket", "date", "currency", "total"}).
		AddRow(1, 1, "Any", ts, "EUR", 123.45)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, supermarket, receipt_date, currency, total FROM receipts WHERE supermarket LIKE ? AND DATE(receipt_date) = DATE(?) AND total = ?")).
		WithArgs("%Any%", ts.Format(time.RFC3339), 123.45).
		WillReturnRows(receipt_rows)

	mock.ExpectBegin()
	mock.ExpectCommit()

	receipt := Receipt{UserID: 1, Supermarket: "Any", Date: ts, Total: 123.45}
	created_receipt, err := CreateReceipt(db, &receipt)

	if created_receipt != nil {
		t.Fatalf("Created duplicated receipt")
	}

	if err.Error() != "Receipt already exists" {
		t.Fatalf("Unexpected error creating receipt: %s", err)
	}

}

func TestCreateDuplicatedReceiptForDifferentUser(t *testing.T) {
	db, mock, err := sqlmock.New()

	if err != nil {
		t.Fatalf("Unexpected error %s connecting to database", err)
	}

	defer db.Close()

	ts := time.Now()

	receipt_rows := mock.NewRows([]string{"id", "user_id", "supermarket", "date", "currency", "total"}).
		AddRow(1, 1, "Any", ts, "EUR", 123.45)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, supermarket, receipt_date, currency, total FROM receipts WHERE supermarket LIKE ? AND DATE(receipt_date) = DATE(?) AND total = ?")).
		WithArgs("%Any%", ts.Format(time.RFC3339), 123.45).
		WillReturnRows(receipt_rows)

	mock.ExpectBegin()

	items := []ReceiptItem{{Name: "Item 1", Quantity: 1, Price: 10, UnitPrice: 11},
		{Name: "Item 2", Quantity: 2, Price: 20, UnitPrice: 22}}

	receipt := Receipt{UserID: 2, Supermarket: "Any", Date: ts, Currency: "EUR", Total: 123.45, Items: items}

	// Insert receipt
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO receipts (user_id, supermarket, receipt_date, currency, total) VALUES (?, ?, ?, ?, ?)")).
		WithArgs(2, "Any", ts.Format(time.RFC3339), "EUR", 123.45).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Insert receipt items
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO receipt_items (receipt_id, quantity, name, unit_price, price) VALUES (?, ?, ?, ?, ?)")).
		WithArgs(1, 1.0, "Item 1", 11.0, 10.0).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO receipt_items (receipt_id, quantity, name, unit_price, price) VALUES (?, ?, ?, ?, ?)")).
		WithArgs(1, 2.0, "Item 2", 22.0, 20.0).
		WillReturnResult(sqlmock.NewResult(2, 1))

	created_receipt, err := CreateReceipt(db, &receipt)

	if created_receipt == nil {
		t.Fatalf("Receipt not created")
	}

	if err != nil {
		t.Fatalf("Unexpected error creating receipt: %s", err)
	}

	mock.ExpectCommit()
}

func TestCreateNonDuplicatedReceipt(t *testing.T) {
	db, mock, err := sqlmock.New()

	if err != nil {
		t.Fatalf("Unexpected error %s connecting to database", err)
	}

	defer db.Close()

	ts := time.Now()

	//receipt := Receipt{UserID: 1, Supermarket: "Any", Date: ts, Total: 100.0}

	receipt_rows := mock.NewRows([]string{"id", "user_id", "supermarket", "date", "currency", "total"})

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, supermarket, receipt_date, currency, total FROM receipts WHERE supermarket LIKE ? AND DATE(receipt_date) = DATE(?) AND total = ?")).
		WithArgs("%Any%", ts.Format(time.RFC3339), 123.45).
		WillReturnRows(receipt_rows)

	mock.ExpectBegin()

	items := []ReceiptItem{{Name: "Item 1", Quantity: 1, Price: 10, UnitPrice: 11},
		{Name: "Item 2", Quantity: 2, Price: 20, UnitPrice: 22}}

	receipt := Receipt{UserID: 2, Supermarket: "Any", Date: ts, Total: 123.45, Currency: "EUR", Items: items}

	// Insert receipt
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO receipts (user_id, supermarket, receipt_date, currency, total) VALUES (?, ?, ?, ?, ?)")).
		WithArgs(2, "Any", ts.Format(time.RFC3339), "EUR", 123.45).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Insert receipt items
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO receipt_items (receipt_id, quantity, name, unit_price, price) VALUES (?, ?, ?, ?, ?)")).
		WithArgs(1, 1.0, "Item 1", 11.0, 10.0).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO receipt_items (receipt_id, quantity, name, unit_price, price) VALUES (?, ?, ?, ?, ?)")).
		WithArgs(1, 2.0, "Item 2", 22.0, 20.0).
		WillReturnResult(sqlmock.NewResult(2, 1))

	created_receipt, err := CreateReceipt(db, &receipt)

	if created_receipt == nil {
		t.Fatalf("Receipt not created")
	}

	if err != nil {
		t.Fatalf("Unexpected error creating receipt: %s", err)
	}

	mock.ExpectCommit()
}
