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

	rows := mock.NewRows([]string{"id", "supermarket", "date", "total"})

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM receipts WHERE supermarket LIKE ? AND date = DATE(?) AND total = ?")).
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

	rows := mock.NewRows([]string{"id", "supermarket", "date", "total"}).
		AddRow(1, "Any", ts, 123.45)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM receipts WHERE supermarket LIKE ? AND date = DATE(?) AND total = ?")).
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
