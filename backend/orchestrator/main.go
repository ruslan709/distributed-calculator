package main

// Импорт необходимых пакетов
import (
	"context"       // Для работы с байтами
	"database/sql"  // Для работы с базами данных SQL
	"encoding/json" // Для кодирования и декодирования JSON
	"fmt"           // Для форматированного вывода и ввода
	"log"           // Для логирования
	"net/http"      // Для работы с HTTP
	"strconv"       // Для конвертации строк в числа и обратно
	"strings"
	"time" // Для работы со временем

	"github.com/golang-jwt/jwt/v4" // Для работы с токенами

	pb "calculatorapi/proto/calculator/calculatorapi/proto/calculator"
	"calculatorapi/utility/database" // Пакет для работы с базой данных
	"calculatorapi/utility/models"   // Пакет с моделями данных

	"golang.org/x/crypto/bcrypt" // Драйвер для хэширования паролей
	"google.golang.org/grpc"
)

// Структура для запроса калькуляции
type CalculationRequest struct {
	UserId             int    `json:"userId"`               // Идентификатор юзера
	Operation          string `json:"operation"`            // Операция для калькуляции
	AddDuration        int    `json:"add_duration"`         // Длительность операции сложения
	SubtractDuration   int    `json:"subtract_duration"`    // Длительность операции вычитания
	MultiplyDuration   int    `json:"multiply_duration"`    // Длительность операции умножения
	DivideDuration     int    `json:"divide_duration"`      // Длительность операции деления
	InactiveServerTime int    `json:"inactive_server_time"` // Время ожидания неактивного сервера
}

// Структура для ответа на запрос калькуляции, содержащая id добавленной операции в базу данных
type CalculationResponse struct {
	ID int `json:"id"` // ID калькуляции
}

// Структура для данных юзера
type Credentials struct {
	Password string `json:"password"`
	Login    string `json:"login"`
}

type Claims struct {
	Login  string `json:"login"`
	UserID int    `json:"userID"`
	jwt.StandardClaims
}

var jwtKey = []byte("secret_key") // Secret key для подписания JWT токенов

// Список развернутых серверов калькуляторов
var servers = []string{
	"http://localhost:8081",
	"http://localhost:8082",
}

// Список развернутых серверов калькуляторов
var GRPCservers = []string{
	"http://localhost:50051",
	"http://localhost:50052",
}

// Структура для статуса сервера калькулятора
type ServerStatus struct {
	URL               string `json:"url"`                     // URL сервера
	Running           bool   `json:"running"`                 // Статус работы сервера
	MaxGoroutines     int    `json:"maxGoroutines,omitempty"` // Максимальное количество горутин
	CurrentGoroutines int    `json:"currentGoroutines"`       // Текущее количество горутин
	Error             string `json:"error,omitempty"`         // Ошибка, если есть
}

// Функция для проверки статуса всех серверов калькуляторов
func pingServers() []ServerStatus {
	var statuses []ServerStatus // Список статусов серверов

	for _, serverURL := range servers {
		status := ServerStatus{URL: serverURL}
		resp, err := http.Get(fmt.Sprintf("%s/ping", serverURL))
		if err != nil {
			// Если запрос не удался, сервер считается неактивным
			status.Running = false
			status.Error = err.Error()
		} else {
			defer resp.Body.Close()
			// Если сервер ответил успешно, читаем ответ
			if resp.StatusCode == http.StatusOK {
				var serverResponse struct {
					Status            string `json:"status"`
					MaxGoroutines     int    `json:"maxGoroutines"`
					CurrentGoroutines int    `json:"currentGoroutines"`
				}
				if err := json.NewDecoder(resp.Body).Decode(&serverResponse); err != nil {
					// Если не удалось декодировать ответ, записываем ошибку
					status.Error = "Failed to decode response"
				} else {
					// Если декодирование успешно, обновляем статус сервера
					status.Running = serverResponse.Status == "running"
					status.MaxGoroutines = serverResponse.MaxGoroutines
					status.CurrentGoroutines = serverResponse.CurrentGoroutines
				}
			} else {
				// Если статус ответа не OK, сервер считается неактивным
				status.Running = false
				status.Error = fmt.Sprintf("Unexpected status code: %d", resp.StatusCode)
			}
		}
		statuses = append(statuses, status) // Добавляем статус в список
	}

	return statuses
}

// Миддлвар для добавления заголовков CORS к ответам сервера
func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Устанавливаем заголовки CORS
		w.Header().Set("Access-Control-Allow-Origin", "*") // or you can specify the exact origin instead of "*"
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		// Если запрос является предварительным запросом CORS, отправляем ответ 200 OK
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Передаем запрос дальше по цепочке обработчиков
		next(w, r)
	}
}

// Функция для отправки калькуляций на серверы калькуляторов
func submitCalculations(db *sql.DB) {
	calculations, err := database.FetchCalculationsToProcess(db)
	if err != nil {
		log.Printf("Error fetching calculations to process: %v", err)
		return
	}

	for _, calc := range calculations {
		submitted := false
		for _, serverURL := range servers {
			if trySubmitCalculation(serverURL, calc) {
				submitted = true
				break // Прекращаем попытки, если успешно отправлено
			}
		}
		if !submitted {
			log.Printf("Failed to submit calculation ID %d to any server", calc.ID)
		}
	}
}

func trySubmitCalculation(serverURL string, calc models.CalculationRequest) bool {
	// Create a gRPC request from the CalculationRequest
	req := &pb.CalculationRequest{
		Id:        int32(calc.ID),
		Operation: calc.Operation,
		Times: map[string]int32{
			"add_duration":      int32(calc.AddDuration),
			"subtract_duration": int32(calc.SubtractDuration),
			"multiply_duration": int32(calc.MultiplyDuration),
			"divide_duration":   int32(calc.DivideDuration),
		},
	}

	// Get the index of the serverURL in the servers list
	index := -1
	for i, s := range servers {
		if s == serverURL {
			index = i
			break
		}
	}

	if index == -1 {
		log.Printf("Server URL %s not found in the servers list", serverURL)
		return false
	}

	// Get the corresponding gRPC server URL from the GRPCservers list
	grpcServerURL := GRPCservers[index]

	// Remove the "http://" part from the grpcServerURL
	grpcServerURL = strings.TrimPrefix(grpcServerURL, "http://")

	// Call the startCalculationGRPC function to start the calculation via gRPC
	return startCalculationGRPC(grpcServerURL, req)
}

// // Попытка отправить калькуляцию на указанный сервер
// func trySubmitCalculation(serverURL string, calc models.CalculationRequest) bool {
// 	// Формируем тело запроса
//     payload, err := json.Marshal(struct {
// 		UserId    int            `json:"userId"`
//         ID        int            `json:"id"`
//         Operation string         `json:"operation"`
//         Times     map[string]int `json:"times"`
//     }{
//         ID:        calc.ID,
// 		UserId:    calc.UserId,
//         Operation: calc.Operation,
//         Times: map[string]int{
//             "add_duration":     calc.AddDuration,
//             "subtract_duration": calc.SubtractDuration,
//             "multiply_duration": calc.MultiplyDuration,
//             "divide_duration":   calc.DivideDuration,
//         },
//     })

//     if err != nil {
//         log.Printf("Error marshaling calculation payload: %v", err)
//         return false
//     }

// 	// Отправляем запрос на сервер калькулятора
//     resp, err := http.Post(fmt.Sprintf("%s/calculate", serverURL), "application/json", bytes.NewBuffer(payload))
//     if err != nil {
//         log.Printf("Error submitting calculation to server %s: %v", serverURL, err)
//         return false
//     }
//     defer resp.Body.Close()

//     // Считаем отправку успешной, если сервер ответил статусом 200 OK или 202 Accepted
//     if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusAccepted {
// 		log.Printf("Successfully submitted calculation ID %d for user %d to server %s", calc.ID, calc.UserId, serverURL)
// 		return true
// 	}

//     log.Printf("Server %s responded with status code %d", serverURL, resp.StatusCode)
//     return false
// }

// calculateTotalOperationTime рассчитывает общее время выполнения операции.
// Входные данные: строка операции и время выполнения для каждого типа операций.
// Возвращает общее время выполнения операции в секундах.
func calculateTotalOperationTime(operation string, addDuration, subtractDuration, multiplyDuration, divideDuration int) int {
	totalDuration := 0

	// Проходим по каждому символу в строке операции
	for _, char := range operation {
		// Определяем тип операции и добавляем соответствующее ей время к общему времени
		switch char {
		case '+':
			totalDuration += addDuration
		case '-':
			totalDuration += subtractDuration
		case '*':
			totalDuration += multiplyDuration
		case '/':
			totalDuration += divideDuration
		}
	}

	return totalDuration
}

// checkAndRestartFailedOperations проверяет и перезапускает операции, которые не были завершены в ожидаемое время.
func checkAndRestartFailedOperations(db *sql.DB) {
	log.Println("Starting checkAndRestartFailedOperations")

	// SQL-запрос для получения операций со статусом 'work'
	query := `
        SELECT id, userId, operation, start_time, add_duration, subtract_duration, multiply_duration, divide_duration
        FROM calculations
        WHERE status = 'work'
    `

	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Error querying 'work' status operations: %v", err)
		return
	}
	defer rows.Close()

	now := time.Now().UTC() // Текущее время в формате UTC
	log.Printf("Current time (UTC): %v", now)

	// Обработка каждой строки результата запроса
	for rows.Next() {
		var (
			id               int
			userId           int
			operation        string
			startTime        time.Time
			addDuration      int
			subtractDuration int
			multiplyDuration int
			divideDuration   int
		)

		if err := rows.Scan(&id, &userId, &operation, &startTime, &addDuration, &subtractDuration, &multiplyDuration, &divideDuration); err != nil {
			log.Printf("Error scanning 'work' status operation: %v", err)
			continue
		}

		operationTime := calculateTotalOperationTime(operation, addDuration, subtractDuration, multiplyDuration, divideDuration)
		expectedEndTime := startTime.Add(time.Duration(operationTime) * time.Second).Add(3 * time.Minute)

		log.Printf("Operation ID %d, User Id: %d Start time: %v, Operation time: %d seconds, Expected end time: %v", id, userId, startTime, operationTime, expectedEndTime)

		// Если текущее время превышает ожидаемое время завершения операции, обновляем статус на 'created'
		if now.After(expectedEndTime) {
			log.Printf("Operation ID %d exceeded expected end time. Resetting status to 'created'.", id)

			resetQuery := `
                UPDATE calculations
                SET status = 'created', start_time = NULL
                WHERE id = $1
            `

			if _, err := db.Exec(resetQuery, id); err != nil {
				log.Printf("Error resetting operation ID %d to 'created': %v", id, err)
			} else {
				log.Printf("Operation ID %d has been reset to 'created' due to timeout.", id)
			}
		} else {
			log.Printf("Operation ID %d is still within the expected time frame.", id)
		}
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating over 'work' status operations: %v", err)
	}

	log.Println("Completed checkAndRestartFailedOperations")
}

// Функция для отправки JSON ошибок
func sendJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func startCalculationGRPC(serverURL string, req *pb.CalculationRequest) bool {
	// Create a gRPC connection to the server
	conn, err := grpc.Dial(serverURL, grpc.WithInsecure())
	if err != nil {
		log.Printf("Failed to dial server %s: %v", serverURL, err)
		return false
	}
	defer conn.Close()

	// Create a gRPC client
	client := pb.NewCalculatorServiceClient(conn)

	// Call the PerformCalculation RPC method
	resp, err := client.PerformCalculation(context.Background(), req)
	if err != nil {
		log.Printf("Failed to start calculation on server %s: %v", serverURL, err)
		return false
	}

	// Check if the response indicates success
	if resp != nil && resp.Id == req.Id {
		log.Printf("Successfully started calculation ID %d on server %s", req.Id, serverURL)
		return true
	}

	log.Printf("Failed to start calculation ID %d on server %s", req.Id, serverURL)
	return false
}

// Основная функция, запускающая сервер
func main() {
	// Инициализация соединения с базой данных на старте приложения
	database.InitializeDB()
	database.SetupDatabase()

	// Определение канала для управления выключением
	shutdownCh := make(chan struct{})

	// Горутина периодической отправки задач на калькуляторы
	go func() {
		db := database.GetDB()                     // Получение глобального объекта базы данных
		ticker := time.NewTicker(30 * time.Second) // Таймер для периодической проверки
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				submitCalculations(db) // Отправка вычислений на обработку
			case <-shutdownCh:
				log.Println("Stopping submission of new calculations.")
				return
			}
		}
	}()

	// Обработчик для эндпоинта /submit-calculation.
	// Принимает запросы на добавление новых вычислений.
	http.HandleFunc("/submit-calculation", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		// Возвращаем ошибку, если метод запроса не POST
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req CalculationRequest
		// Декодирование тела запроса в структуру CalculationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			// В случае ошибки декодирования возвращаем ошибку Bad Request
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// fmt.Println("AddDuration:", req.AddDuration)
		// fmt.Println("SubtractDuration:", req.SubtractDuration)
		// fmt.Println("MultiplyDuration:", req.MultiplyDuration)
		// fmt.Println("DivideDuration:", req.DivideDuration)
		// fmt.Println("InactiveServerTime:", req.InactiveServerTime)

		// Вставка данных о вычислении в базу данных
		db := database.GetDB()
		id, err := database.InsertCalculation(db, req.UserId, req.Operation, req.AddDuration, req.SubtractDuration, req.MultiplyDuration, req.DivideDuration, req.InactiveServerTime)
		// В случае ошибки при записи в базу данных возвращаем ошибку сервера
		if err != nil {
			log.Fatal("Error writing data to database:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		type CalculationResponse struct {
			ID        int    `json:"id"`
			UserId    int    `json:"userId"`
			Status    string `json:"status"`
			Operation string `json:"operation"`
		}

		// Создаем ответ сервера с ID созданного вычисления
		status := "created"
		resp := CalculationResponse{ID: id, UserId: req.UserId, Status: status, Operation: req.Operation}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))

	// Обработчик для проверки статуса серверов калькуляторов.
	http.HandleFunc("/ping-servers", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		statuses := pingServers() // Получение статусов серверов
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(statuses)
	}))

	// Обработчик для получения статуса оркестратора.
	http.HandleFunc("/orchestrator-status", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		status := struct {
			Running bool   `json:"running"`
			Message string `json:"message"`
		}{
			Running: true,
			Message: "Orchestrator is running",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	}))

	// Обработчик для получения результата вычисления по ID.
	http.HandleFunc("/get-calculation-result", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters
		idParam := r.URL.Query().Get("id")
		if idParam == "" {
			http.Error(w, "Missing id parameter", http.StatusBadRequest)
			return
		}

		id, err := strconv.Atoi(idParam)
		if err != nil {
			http.Error(w, "Invalid id parameter", http.StatusBadRequest)
			return
		}

		db := database.GetDB()
		if err != nil {
			log.Printf("Error setting up the database: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		result, err := database.GetCalculationResultByID(db, id)
		if err != nil {
			log.Printf("Error fetching calculation result: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}))

	// Обработчик для получения всех вычислений из базы данных.
	http.HandleFunc("/get-all-calculations", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		db := database.GetDB()

		calculations, err := database.FetchAllCalculations(db)
		if err != nil {
			log.Printf("Error fetching all calculations: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(calculations)
	}))

	// Обработчик для получения всех вычислений по userId.
	http.HandleFunc("/get-calculations-by-user", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		db := database.GetDB()
		userIdParam := r.URL.Query().Get("userId") // Получение userId из параметров запроса.

		if userIdParam == "" {
			http.Error(w, "User ID is required", http.StatusBadRequest)
			return
		}

		userId, err := strconv.Atoi(userIdParam) // Преобразование userId в int.
		if err != nil {
			http.Error(w, "Invalid User ID", http.StatusBadRequest)
			return
		}

		calculations, err := database.FetchCalculationsByUser(db, userId)
		if err != nil {
			log.Printf("Error fetching calculations for user %d: %v", userId, err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(calculations)
	}))

	// Обработчик для очистки всех вычислений из базы данных.
	http.HandleFunc("/clear-all-calculations", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		db := database.GetDB()

		if err := database.ClearAllCalculations(db); err != nil {
			log.Printf("Error clearing all calculations: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "All calculations have been cleared successfully.")
	}))

	// Обработчик для регистрации нового пользователя по логину и паролю.
	http.HandleFunc("/api/v1/register", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			sendJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var newUser models.User
		err := json.NewDecoder(r.Body).Decode(&newUser)
		if err != nil {
			sendJSONError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Call the database function to insert the new user
		err = database.RegisterUser(database.GetDB(), newUser.Login, newUser.Password)
		if err != nil {
			log.Printf("Error registering user: %v", err)
			sendJSONError(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	}))

	// Обработчик для получения пользователя по логину через POST-запрос.
	http.HandleFunc("/get-user", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		// Определение структуры для разбора логина из тела запроса
		var requestData struct {
			Login string `json:"login"`
		}

		// Декодирование JSON-тела запроса в структуру requestData
		err := json.NewDecoder(r.Body).Decode(&requestData)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if requestData.Login == "" {
			http.Error(w, "Missing login field", http.StatusBadRequest)
			return
		}

		db := database.GetDB()
		user, err := database.GetUserByLogin(db, requestData.Login)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "User not found", http.StatusNotFound)
				return
			}
			log.Printf("Error fetching user: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	}))

	// Обработчик для процесса входа в систему, логина
	http.HandleFunc("/api/v1/login", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		var creds Credentials
		err := json.NewDecoder(r.Body).Decode(&creds)
		if err != nil {
			sendJSONError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Получение пользователя из базы данных
		db := database.GetDB()
		user, err := database.GetUserByLogin(db, creds.Login)
		if err != nil {
			sendJSONError(w, "Login failed", http.StatusUnauthorized)
			return
		}

		// Сравнение хэшированного пароля, предполагая, что он захэширован с использованием bcrypt
		if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
			sendJSONError(w, "Login failed, Incorrect Password", http.StatusUnauthorized)
			return
		}

		// Определение времени истечения токена
		expirationTime := time.Now().Add(24 * time.Hour)
		claims := &Claims{
			Login:  creds.Login,
			UserID: user.ID,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: expirationTime.Unix(),
			},
		}

		// Объявление токена с использованием алгоритма подписи и утверждений
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		// Создание строки JWT
		tokenString, err := token.SignedString(jwtKey)
		if err != nil {
			sendJSONError(w, "Error creating the JWT token", http.StatusInternalServerError)
			return
		}

		// Установка клиентского куки для "token" как только что сгенерированного JWT,
		// устанавливаем время истечения, которое совпадает с временем истечения токена
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"jwt": tokenString})
	}))

	// Горутина для периодической проверки и перезапуска неудачных операций.
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				db := database.GetDB()
				checkAndRestartFailedOperations(db)
			case <-shutdownCh:
				log.Println("Shutting down check and restart operations.")
				return
			}
		}
	}()

	// Запуск HTTP-сервера на порту 8080.
	fmt.Println("Server is running on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Error starting server:", err)
		// Закрытие канала при остановке сервера
		close(shutdownCh)
	}
}
