# Распределённый калькулятор - Финальная часть проекта
### Обзор
- Персистентность - возможность сохранять состояние программы и восстанавливать его после перезагрузки.
- Многопользовательский режим - все операции выполняются в контексте конкретного пользователя, данные хранятся в СУБД.

### Рекомендации по установке и запуску
Для удобной и комфортной разработки и запуска проекта рекомендую:

- Установить [Go](https://go.dev/doc/install) - официальный язык программирования, на котором написан backend.
- Установить [Visual Studio Code](https://code.visualstudio.com/) - удобный редактор кода с поддержкой Go.
- Установить [PostgreSQL](postgresql.org)- Система управления базами данных, которая используется в проекте для хранения и обработки данных.
- Установить [Git](https://git-scm.com/downloads) - для клонирования репозитория проекта.После установки проверьте корректность командой:
`git --version`
- Установить [Postman](https://www.postman.com/downloads/) - для удобного тестирования и отладки API нашего проекта
- В Visual Studio Code установить расширение **Live Server** для автоматического запуска локального сервера и обновления страницы при изменениях в коде.

### Установка Go
  1. Перейдите на официальный сайт: [https://go.dev/doc/install]
  2. Выберите свою операционную систему и скачайте установочный пакет.
  3. Следуйте инструкциям установщика.
  4. После установки добавьте путь `/usr/local/go/bin` (Linux/macOS/Windows) или путь установки Go в переменную окружения `PATH`.
  5. Проверьте установку, открыв терминал и выполнив команду:
     ```
     go version
     ```
     Вы должны увидеть установленную версию Go.

 ### Установка PostgreSQL
  1. Перейдите на официальный сайт: [https://www.postgresql.org/download/]
  2. Выберите свою операционную систему.
  3. Скачайте и установите PostgreSQL, следуя инструкциям установщика.
  4. После установки настройте пользователя и базу данных для проекта.
  5. Убедитесь, что сервер PostgreSQL запущен.    
 
  ### Установка Git
  1. Перейдите на официальный сайт: [https://git-scm.com/downloads](https://git-scm.com/downloads)
  2. Скачайте и установите Git для вашей ОС.
  3. Проверьте установку, открыв терминал и выполнив команду:
     ```
     git --version
     ```
     Вы должны увидеть установленную версию Git.

### Установка Live Server в VS Code
1. Откройте VS Code.
2. Перейдите в раздел расширений (иконка квадратов слева или `Ctrl+Shift+X`).
3. В поиске введите `Live Server`.
4. Нажмите кнопку **Установить**.
5. После установки в правом нижнем углу появится кнопка **Go Live** - нажмите её для запуска сервера.

### Установка Postman
1. Перейдите на официальный сайт: [https://www.postman.com/downloads/]
2. Выберите версию для вашей операционной системы (Windows, macOS, Linux).
3. Скачайте и установите приложение.
4. Запустите Postman и создайте учётную запись (рекомендуется для синхронизации).

## Структура проекта
Проект состоит из двух основных частей:

- ![Frontend](https://github.com/ruslan709/distributed-calculator/blob/main/frontend/Readme(ru).md) - клиентская часть, реализованная на HTML, CSS и JavaScript. В папке `frontend` находится подробная документация по установке и запуску фронтенда.

- **Backend** - серверная часть, написанная на Go, обеспечивающая логику приложения, работу с базой данных и API.

```css
distributed-calculator
│
├── backend
│   ├── calc1
│   │   └── main.go
│   ├── calc2
│   │   └── main.go
│   ├── orchestrator
│   │   └── main.go
│   └── utility
│       ├── calculation
│       │   └── calculation.go
│       ├── database
│       │   └── database.go
│       └── models
│           └── calculations.go
│
├── frontend
│   ├── index.html
│   ├── script.js
│   ├── styles.css
│   ├── README(ru).md       
│   └── README(eng).md       
│
├── go.mod
├── go.sum
├── README(ru).md          
└── README(eng).md          
```

![ Основные компоненты  frontend](https://github.com/ruslan709/distributed-calculator/blob/main/frontend/Readme(ru).md)
### Основные компоненты backend:
- **calc1/** и **calc2/**  
  Два отдельных сервиса-калькулятора, которые выполняют арифметические вычисления. Каждый реализован как отдельное приложение с файлом `main.go`.
- **orchestrator/**  
   Оркестратор работает как главный диспетчер: принимает примеры от пользователей, распределяет их между калькуляторами, следит за нагрузкой, сохраняет результаты и проверяет права доступа.
 - **utility/**  
  Вспомогательные пакеты, используемые во всех сервисах backend:
  - `calculation/` - содержит логику вычислений и алгоритмы обработки арифметических выражений.
  - `database/` - реализует работу с базой данных PostgreSQL, включая хранение и обновление данных о вычислениях и пользователях.
  - `models/` - описывает структуры данных и модели, используемые в проекте (например, модели вычислений, пользователей и т.д.).

### Начало работы
### Копирование проекта с GitHub

Для начала работы с проектом необходимо клонировать репозиторий на локальный компьютер.  
Откройте терминал и выполните команду:
`git clone https://github.com/ruslan709/distributed-calculator.git`

После клонирования репозитория перейдите в папку проекта для выполнения последующих команд:
`cd distributed-calculator`

### Установка зависимостей
Чтобы установить все зависимости проекта, выполните в директории проекта команду:
`go mod tidy`
Эта команда скачает и установит все необходимые зависимости, указанные в файле `go.mod`.

### Инструкция по запуску проекта 

### Запуск backend-сервисов
Проект включает несколько backend-сервисов: оркестратор и два калькулятора (`calc1` и `calc2`).  
Для корректной работы их необходимо запускать **в отдельных терминальных окнах или вкладках**.

#### Запуск оркестратора
Откройте новый терминал, перейдите в папку оркестратора и запустите сервис:
`cd backend/orchestrator`
`go run main.go`

#### Запуск первого калькулятора
В другом терминале перейдите в папку первого калькулятора и запустите его:
`cd backend/calc1`
`go run main.go`

#### Запуск второго калькулятора
В третьем терминале перейдите в папку второго калькулятора и запустите его:
`cd backend/calc2`
`go run main.go`

После запуска всех сервисов backend будет готов к работе.  
Для запуска фронтенда используйте инструкции из папки **[Frontend](./frontend/README(ru).md)**

#### Получение статуса оркестратора

```bash
curl -X GET http://localhost:8080/orchestrator-status```

Пример ответа сервера:
```json
{
  "running": true,
  "message": "Orchestrator is running"
}
```

#### Получение статусов серверов калькулятора

```bash
curl -X GET http://localhost:8080/ping-servers
```

Пример ответа сервера:
```json
[
  {
    "url": "http://localhost:8081",
    "running": true,
    "maxGoroutines": 5,
    "currentGoroutines": 2
  },
  {
    "url": "http://localhost:8082",
    "running": false,
    "error": "Connection refused"
  }
]
```

Для отправки вычислительной задачи на сервер калькулятора, используйте следующий `curl` запрос:

```bash
curl -X POST http://localhost:8081/calculate -H "Content-Type: application/json" -d '{
    "id": 1,
    "userId": "1",
    "operation": "2+2",
    "times": {
        "add_duration": 1,
        "subtract_duration": 1,
        "multiply_duration": 1,
        "divide_duration": 1
    }
}'
```

Пример ответа сервера:
```json
{
  "message": "Calculation started successfully."
}
```

#### Получение текущего количества горутин
```bash
curl -X GET http://localhost:8081/goroutines
```

Пример ответа сервера:
```plaintext
Current number of goroutines: 1
```

#### Проверка состояния сервера калькулятора
```bash
curl -X GET http://localhost:8081/ping
```

Пример ответа сервера:
```json
{
  "status": "running",
  "maxGoroutines": 5,
  "currentGoroutines": 1
}
```

#### Остановка сервера калькулятора
```bash
curl -X POST http://localhost:8081/shutdown
```

Пример ответа сервера:
```plaintext
Server is shutting down...
```
## Интерфейс
![Иллюстрация для проекта](https://github.com/ruslan709/distributed-calculator/tree/main/frontend/interface)
# Спасибо за интерес к проекту «Распределённый калькулятор»! # 
Данный проект демонстрирует принципы построения масштабируемых, отказоустойчивых и многопользовательских распределённых систем на языке Go с использованием современных технологий: gRPC, PostgreSQL, JWT-авторизации и микросервисной архитектуры.
# Особая благодарность команде курса  Go-разработчиков за поддержку и полезные материалы.  
