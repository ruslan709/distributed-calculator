package models

// CalculationRequest определяет структуру запроса на вычисление.
type CalculationRequest struct {
	ID                 int    `json:"id"`                             // Идентификатор запроса, должен соответствовать схеме базы данных
	UserId             int    `json:"userId"`                         // Идентификатор юзера
	Operation          string `json:"operation"`                      // Строка операции, например "2+2"
	AddDuration        int    `json:"add_duration"`                   // Продолжительность операции сложения в секундах
	SubtractDuration   int    `json:"subtract_duration"`              // Продолжительность операции вычитания в секундах
	MultiplyDuration   int    `json:"multiply_duration"`              // Продолжительность операции умножения в секундах
	DivideDuration     int    `json:"divide_duration"`                // Продолжительность операции деления в секундах
	InactiveServerTime int    `json:"inactive_server_time,omitempty"` // Время бездействия сервера, может быть опущено
}

// CalculationResponse определяет структуру для возвращения результатов вычислений.
type CalculationResponse struct {
	ID        int     `json:"id"`               // Идентификатор запроса
	Operation string  `json:"operation"`        // Результат вычисления
	UserId    int     `json:"userId"`           // Идентификатор юзера
	Result    float64 `json:"result,omitempty"` // Результат вычисления, может быть опущен, если вычисление не завершено
	Status    string  `json:"status"`           // Статус запроса, например "completed" или "error"
}

// OperationResponse определяет структуру для возвращения информации об операции.
type OperationResponse struct {
	ID        int     `json:"id"`               // Идентификатор операции
	UserId    int     `json:"userId"`           // Идентификатор юзера
	Operation string  `json:"operation"`        // Строка операции, выполненной калькулятором
	Result    float64 `json:"result,omitempty"` // Результат операции, может быть опущен, если операция не завершена
	Status    string  `json:"status"`           // Статус операции, например "created", "work" или "completed"
}

// User определяет структуру для юзера.
type User struct {
	ID       int    `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password"`
}
