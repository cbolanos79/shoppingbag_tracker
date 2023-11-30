package model

import (
	"testing"

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

	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}
