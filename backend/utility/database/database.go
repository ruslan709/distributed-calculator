// Пакет database предоставляет функции для работы с базой данных PostgreSQL.
package database

import (
	"calculatorapi/utility/models" // Структуры данных для калькулятора
	"database/sql"                 // Импорт пакета для работы с SQL базами данных
	"fmt"                          // Форматированный вывод
	"log"                          // Логирование
	"sync"                         // Синхронизация горутин
	"time"                         // Работа со временем

	_ "github.com/lib/pq"        // Драйвер PostgreSQL
	"golang.org/x/crypto/bcrypt" // Драйвер для хэширования паролей
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "yandexcalculate123"
	dbname   = "postgres"
)

var (
	db   *sql.DB
	dbMu sync.Mutex
)

func InitializeDB() {
	dbMu.Lock()
	defer dbMu.Unlock()

	if db != nil {
		return
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	var err error
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	fmt.Println("Database connection established")
}

func InitializeTestDB(db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("cannot connect to database: %v", err)
	}

	fmt.Println("Database connection established")
	return nil
}
func GetDB() *sql.DB {
	dbMu.Lock()         // Блокировка мьютекса
	defer dbMu.Unlock() // Освобождение мьютекса

	if db == nil {
		InitializeDB()
	}

	// Проверка, живо ли соединение
	if err := db.Ping(); err != nil {
		fmt.Println("Reconnecting to the database...")
		InitializeDB()
	}

	return db
}

func ConnectToDatabase() (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err // Возвращение ошибки при неудаче
	}
	return db, nil // Возвращение объекта соединения
}

func SetupDatabase() (*sql.DB, error) {

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	exists, err := DatabaseExists(db, dbname)
	if err != nil {
		return nil, err
	}

	if !exists {
		_, err := db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbname))
		if err != nil {
			return nil, err
		}
		fmt.Printf("Database '%s' created successfully.\n", dbname)
	}
	fmt.Println("Checking and creating tables if necessary...")
	err = CreateTableIfNotExists(db)
	if err != nil {
		log.Fatalf("Failed to create Calculations tables: %v", err)
		return nil, err
	}

	err = CreateUserTableIfNotExists(db)
	if err != nil {
		log.Fatalf("Failed to create User tables: %v", err)
		return nil, err
	}

	return db, nil
}

func DatabaseExists(db *sql.DB, dbName string) (bool, error) {
	// Выполнение SQL запроса для проверки существования базы данных
	var exists bool
	err := db.QueryRow("SELECT EXISTS (SELECT FROM pg_database WHERE datname = $1)", dbName).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func CreateTableIfNotExists(db *sql.DB) error {
	// Выполнение SQL запроса для создания таблицы
	var tableExists bool
	err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'calculations')").Scan(&tableExists)
	if err != nil {
		return err
	}

	if tableExists {
		fmt.Println("Table 'calculations' already exists.")
		return nil
	}

	query := `
		CREATE TABLE calculations (
			id SERIAL PRIMARY KEY,
            userId INTEGER NOT NULL,
			operation TEXT,
			result DOUBLE PRECISION,
			status TEXT,
			created_time TIMESTAMP,
			start_time TIMESTAMP,
			end_time TIMESTAMP,
			operation_server TEXT,
			server_status TEXT,
			add_duration INTEGER,
			subtract_duration INTEGER,
			multiply_duration INTEGER,
			divide_duration INTEGER,
			inactive_server_time INTEGER
		)
	`

	_, err = db.Exec(query)
	if err != nil {
		return err
	}
	fmt.Println("Table 'calculations' created successfully.")
	return nil
}

func InsertCalculation(db *sql.DB, userId int, operation string, addDuration, subtractDuration, multiplyDuration, divideDuration, inactiveServerTime int) (int, error) {

	if err := db.Ping(); err != nil {

		fmt.Println("Reconnecting to the database...")
		if err := db.Close(); err != nil {
			return 0, err
		}
		db, err = SetupDatabase()
		if err != nil {
			return 0, err
		}
	}

	query := `
        INSERT INTO calculations (userId, operation, status, created_time, add_duration, subtract_duration, multiply_duration, divide_duration, inactive_server_time)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        RETURNING id
    `
	status := `created`
	createdTime := time.Now().UTC()

	var id int
	err := db.QueryRow(query, userId, operation, status, createdTime, addDuration, subtractDuration, multiplyDuration, divideDuration, inactiveServerTime).Scan(&id)
	if err != nil {
		return 0, err
	}

	fmt.Println("Calculation record inserted successfully.")
	return id, nil
}

func RunCheckCreatedRecords(db *sql.DB) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := checkAndPrintCreatedRecords(db)
			if err != nil {
				fmt.Printf("Error during checkAndPrintCreatedRecords: %v\n", err)
			}
		}
	}
}

func checkAndPrintCreatedRecords(db *sql.DB) error {

	query := `
		SELECT id, userId, operation, created_time, add_duration, subtract_duration, multiply_duration, divide_duration, inactive_server_time
		FROM calculations
		WHERE status = 'created'
	`

	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("error querying database: %w", err)
	}
	defer rows.Close()

	fmt.Println("Records with status 'created':")
	for rows.Next() { // Перебор всех полученных записей.
		var (
			id                 int
			userId             int
			operation          string
			createdTime        time.Time
			addDuration        int
			subtractDuration   int
			multiplyDuration   int
			divideDuration     int
			inactiveServerTime int
		)

		if err := rows.Scan(&id, &userId, &operation, &createdTime, &addDuration, &subtractDuration, &multiplyDuration, &divideDuration, &inactiveServerTime); err != nil {
			return fmt.Errorf("error scanning row: %w", err)
		}

		fmt.Printf("ID: %d, User ID: %d, Operation: %s, Created Time: %s, Add Duration: %d, Subtract Duration: %d, Multiply Duration: %d, Divide Duration: %d, Inactive Server Time: %d\n",
			id, userId, operation, createdTime.Format("2006-01-02 15:04:05"), addDuration, subtractDuration, multiplyDuration, divideDuration, inactiveServerTime)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating over rows: %w", err)
	}
	return nil
}

func UpdateCalculation(db *sql.DB, id int, result float64, status string) error {

	query := `
        UPDATE calculations
        SET result = $1, status = $2, end_time = $3
        WHERE id = $4
    `
	endTime := time.Now().UTC()

	_, err := db.Exec(query, result, status, endTime, id)
	if err != nil {
		return err
	}

	fmt.Printf("Calculation record with ID %d updated successfully.\n", id)
	return nil
}

func UpdateCalculationStatusToWork(db *sql.DB, id int) error {
	query := `
        UPDATE calculations
        SET status = 'work', start_time = timezone('UTC', NOW())
        WHERE id = $1
    `

	_, err := db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error updating calculation status to work and setting start time: %w", err)
	}

	fmt.Printf("Calculation record with ID %d status updated to work and start time set.\n", id)
	return nil
}

func FetchCalculationsToProcess(db *sql.DB) ([]models.CalculationRequest, error) {
	var calculations []models.CalculationRequest

	query := `SELECT id, userId, operation, add_duration, subtract_duration, multiply_duration, divide_duration FROM calculations WHERE status = 'created' LIMIT 5`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // Закрытие результата запроса при выходе из функции.

	for rows.Next() { // Перебор всех полученных записей.
		var calc models.CalculationRequest
		if err := rows.Scan(&calc.ID, &calc.UserId, &calc.Operation, &calc.AddDuration, &calc.SubtractDuration, &calc.MultiplyDuration, &calc.DivideDuration); err != nil {
			return nil, err // Возврат ошибки при возникновении.
		}
		calculations = append(calculations, calc) // Добавление записи в слайс.
	}

	if err = rows.Err(); err != nil {
		return nil, err // Возврат ошибки при возникновении.
	}

	return calculations, nil // Возвращение слайса с результатами и nil в случае успешного выполнения функции.
}

// GetCalculationResultByID извлекает результат вычисления по его ID.
func GetCalculationResultByID(db *sql.DB, id int) (*models.CalculationResponse, error) {
	var (
		operation string
		result    sql.NullFloat64 // Использование sql.NullFloat64 для обработки NULL значений.
		status    string
		userId    int
	)
	query := `SELECT operation, result, status, userId FROM calculations WHERE id = $1` // SQL-запрос для выборки.
	err := db.QueryRow(query, id).Scan(&operation, &result, &status, &userId)           // Выполнение запроса и считывание результатов.
	if err != nil {
		return nil, err // Возврат ошибки при возникновении.
	}

	calcResult := &models.CalculationResponse{
		ID:        id,
		Operation: operation,
		UserId:    userId,
		Status:    status,
	}

	if result.Valid {
		calcResult.Result = result.Float64 // Присвоение результата, если он не NULL.
	}

	return calcResult, nil // Возвращение ответа и nil в случае успешного выполнения функции.
}

// FetchAllCalculations извлекает все вычисления из базы данных.
func FetchAllCalculations(db *sql.DB) ([]models.OperationResponse, error) {
	var calculations []models.OperationResponse // Слайс для хранения результатов.

	query := `SELECT id, userId, operation, result, status FROM calculations` // SQL-запрос для выборки всех записей.
	rows, err := db.Query(query)                                              // Выполнение запроса.
	if err != nil {
		return nil, fmt.Errorf("querying calculations: %w", err)
	}
	defer rows.Close() // Закрытие результата запроса при выходе из функции.

	for rows.Next() { // Перебор всех полученных записей.
		var calc models.OperationResponse
		var result sql.NullFloat64 // Использование sql.NullFloat64 для обработки NULL значений.

		if err := rows.Scan(&calc.ID, &calc.UserId, &calc.Operation, &result, &calc.Status); err != nil {
			return nil, fmt.Errorf("scanning calculation: %w", err)
		}

		if result.Valid {
			calc.Result = result.Float64 // Присвоение результата, если он не NULL.
		} else {
			// При желании можно обработать NULL результаты иначе, например, присвоить значение по умолчанию или опустить поле.
		}

		calculations = append(calculations, calc) // Добавление записи в слайс.
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating over calculations results: %w", err)
	}

	return calculations, nil // Возвращение слайса с результатами и nil в случае успешного выполнения функции.
}

// FetchCalculationsByUser извлекает все вычисления для конкретного пользователя.
func FetchCalculationsByUser(db *sql.DB, userId int) ([]models.OperationResponse, error) {
	var calculations []models.OperationResponse

	query := `SELECT id, userId, operation, result, status FROM calculations WHERE userId = $1`
	rows, err := db.Query(query, userId) // Выполнение запроса с фильтрацией по userId.
	if err != nil {
		return nil, fmt.Errorf("querying calculations for user %d: %w", userId, err)
	}
	defer rows.Close()

	for rows.Next() {
		var calc models.OperationResponse
		var result sql.NullFloat64 // Для обработки NULL значений.

		if err := rows.Scan(&calc.ID, &calc.UserId, &calc.Operation, &result, &calc.Status); err != nil {
			return nil, fmt.Errorf("scanning calculation: %w", err)
		}

		if result.Valid {
			calc.Result = result.Float64
		}

		calculations = append(calculations, calc)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating over calculations results: %w", err)
	}

	return calculations, nil
}

// ClearAllCalculations удаляет все строки из таблицы 'calculations'.
func ClearAllCalculations(db *sql.DB) error {
	// SQL statement to delete all rows
	query := `DELETE FROM calculations` // SQL-запрос для удаления всех строк.
	_, err := db.Exec(query)            // Выполнение запроса.
	if err != nil {
		return fmt.Errorf("clearing all calculations: %w", err)
	}
	fmt.Println("All calculations cleared successfully.")
	return nil // Возвращение nil в случае успешного выполнения функции.
}

// CreateUserTableIfNotExists проверяет наличие в базе данных таблицы users и создает таковую при ее отсутствии
func CreateUserTableIfNotExists(db *sql.DB) error {
	var tableExists bool
	err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'users')").Scan(&tableExists)
	if err != nil {
		return err
	}

	if !tableExists {
		query := `
        CREATE TABLE users (
            id SERIAL PRIMARY KEY,
            login TEXT UNIQUE NOT NULL,
            password TEXT NOT NULL
        )`
		_, err = db.Exec(query)
		if err != nil {
			return err
		}
		fmt.Println("Table 'users' created successfully.")
	} else {
		fmt.Println("Table 'users' already exists.")
	}
	return nil
}

// RegisterUser добавляет нового юзера в базу данных с хешированным паролем
func RegisterUser(db *sql.DB, login, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	query := "INSERT INTO users (login, password) VALUES ($1, $2)"
	_, err = db.Exec(query, login, string(hashedPassword))
	if err != nil {
		return err
	}

	fmt.Println("User registered successfully.")
	return nil
}

// GetUserByLogin получает юзера по логину из базы данный
func GetUserByLogin(db *sql.DB, login string) (*models.User, error) {
	user := &models.User{}

	query := `SELECT id, login, password FROM users WHERE login = $1`
	row := db.QueryRow(query, login)
	err := row.Scan(&user.ID, &user.Login, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	return user, nil
}
