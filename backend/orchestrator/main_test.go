package main

import (
	"calculatorapi/utility/models"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestPingServers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ping" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":            "running",
				"maxGoroutines":     10,
				"currentGoroutines": 5,
			})
		}
	}))
	defer server.Close()

	servers = []string{server.URL}

	statuses := pingServers()

	if len(statuses) != 1 || !statuses[0].Running {
		t.Errorf("Expected the server to be running but got %v", statuses[0])
	}
}

func TestSubmitCalculations(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "userId", "operation", "add_duration", "subtract_duration", "multiply_duration", "divide_duration"}).
		AddRow(1, 1, "2+2", 10, 10, 10, 10)
	mock.ExpectQuery("^SELECT (.+) FROM calculations").WillReturnRows(rows)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/calculate" {
			var req models.CalculationRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Operation != "2+2" {
				t.Errorf("Expected operation '2+2', got '%s'", req.Operation)
			}
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	servers = []string{server.URL}

	submitCalculations(db)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}
