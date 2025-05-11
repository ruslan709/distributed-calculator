package database

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestInitializeDB(t *testing.T) {
	db, mock, err := sqlmock.New() // Create a new instance of sqlmock
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Setup expectations
	mock.ExpectPing() // Expect a database ping call

	// Call InitializeDB with the mocked DB
	if err := InitializeTestDB(db); err != nil {
		t.Errorf("Failed to initialize database: %s", err)
	}

	// Ensure that all expectations set on the mock database were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
