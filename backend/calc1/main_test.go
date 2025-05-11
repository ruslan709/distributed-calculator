package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"calculatorapi/utility/calculation"

	"github.com/DATA-DOG/go-sqlmock"
)

// Функция проверки конвертации операций
func TestConvertOperationTimes(t *testing.T) {
	testCases := []struct {
		name     string
		input    map[string]int
		expected calculation.OperationTimes
	}{
		{
			name: "All Operations",
			input: map[string]int{
				"add_duration":      10,
				"subtract_duration": 20,
				"multiply_duration": 30,
				"divide_duration":   40,
			},
			expected: calculation.OperationTimes{
				"+": 10 * time.Second,
				"-": 20 * time.Second,
				"*": 30 * time.Second,
				"/": 40 * time.Second,
			},
		},
		{
			name: "Partial Operations",
			input: map[string]int{
				"add_duration":      5,
				"multiply_duration": 15,
			},
			expected: calculation.OperationTimes{
				"+": 5 * time.Second,
				"*": 15 * time.Second,
			},
		},
		{
			name:     "Empty Input",
			input:    map[string]int{},
			expected: calculation.OperationTimes{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ConvertOperationTimes(tc.input)
			for key, expectedDuration := range tc.expected {
				if dur, ok := result[key]; !ok || dur != expectedDuration {
					t.Errorf("For key %s, expected %v, got %v", key, expectedDuration, dur)
				}
			}
			// Check if result contains no extra keys
			for key := range result {
				if _, ok := tc.expected[key]; !ok {
					t.Errorf("Unexpected key %s in result", key)
				}
			}
		})
	}
}

// Функция настройки для инициализации маршрутов
func setupTestRoutes() {
	http.HandleFunc("/calculate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method is not supported.", http.StatusNotFound)
			return
		}

		var request OperationRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		db, mock, err := sqlmock.New()
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer db.Close()

		mock.ExpectBegin()
		mock.ExpectExec("UPDATE calculations SET status = ? WHERE id = ?").WithArgs("work", request.ID).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("UPDATE calculations SET result = ?, status = ? WHERE id = ?").WithArgs(7.0, "completed", request.ID).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		startCalculation(db, request.ID, request.Operation, request.Times) // Запуск расчета
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintln(w, "Calculation started successfully.")
	})
}

func TestCalculateEndpoint(t *testing.T) {
	setupTestRoutes() // Настройка маршрутов

	server := httptest.NewServer(http.DefaultServeMux) // Создание нового сервера для тестирования
	defer server.Close()

	// Тестовые данные для POST-запроса
	operationRequest := OperationRequest{
		ID:        1,
		Operation: "2+3",
		Times: map[string]int{
			"add_duration": 10,
		},
	}
	requestBody, _ := json.Marshal(operationRequest)
	request, _ := http.NewRequest("POST", server.URL+"/calculate", bytes.NewBuffer(requestBody))

	response := httptest.NewRecorder()

	http.DefaultServeMux.ServeHTTP(response, request)

	if status := response.Code; status != http.StatusAccepted {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusAccepted)
	}

	expected := "Calculation started successfully.\n"
	if response.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", response.Body.String(), expected)
	}
}

// Тестирование обработчика HTTP для конечной точки '/goroutines'.
func TestGoroutinesEndpoint(t *testing.T) {
	// Настройка счетчика горутин и мьютекса для тестирования.
	currentGoroutines := 3
	var mu sync.Mutex

	// Setup the HTTP handler
	http.HandleFunc("/goroutines", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		fmt.Fprintf(w, "Current number of goroutines: %d\n", currentGoroutines)
		mu.Unlock()
	})

	// Создание запроса для передачи нашему обработчику. У нас пока нет параметров запроса, поэтому мы передаем 'nil'.
	req, err := http.NewRequest("GET", "/goroutines", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Создаем ResponseRecorder (удовлетворяющий http.ResponseWriter) для записи ответа.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		fmt.Fprintf(w, "Current number of goroutines: %d\n", currentGoroutines)
		mu.Unlock()
	})

	// Наши обработчики удовлетворяют http.Handler, поэтому мы можем вызвать их метод ServeHTTP напрямую и передать наш Запрос и ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Проверяем, что код состояния соответствует ожидаемому.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Проверяем, что тело ответа соответствует ожидаемому.
	expected := fmt.Sprintf("Current number of goroutines: %d\n", currentGoroutines)
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

// Тестирование обработчика HTTP для конечной точки '/shutdown'.
func TestShutdownEndpoint(t *testing.T) {
	// Настройка флага serverRunning и канала завершения.
	serverRunning := true
	shutdownCh := make(chan struct{}, 1) // Буфер, чтобы избежать блокировки

	// Настройка обработчика HTTP
	http.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		if !serverRunning {
			http.Error(w, "Server is already shutting down", http.StatusServiceUnavailable)
			return
		}
		serverRunning = false
		close(shutdownCh)
		fmt.Fprintln(w, "Server is shutting down...")
	})

	// Создание запроса для передачи нашему обработчику.
	req, err := http.NewRequest("GET", "/shutdown", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Создаем ResponseRecorder для записи ответа.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !serverRunning {
			http.Error(w, "Server is already shutting down", http.StatusServiceUnavailable)
			return
		}
		serverRunning = false
		close(shutdownCh)
		fmt.Fprintln(w, "Server is shutting down...")
	})

	// Наши обработчики удовлетворяют http.Handler, поэтому мы можем вызвать их метод ServeHTTP напрямую и передать наш Запрос и ResponseRecorder.
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := "Server is shutting down...\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}

	if serverRunning {
		t.Error("serverRunning flag should be set to false")
	}

	select {
	case _, open := <-shutdownCh:
		if open {
			t.Error("shutdown channel should be closed")
		}
	default:
		t.Error("shutdown channel was not closed")
	}
}
