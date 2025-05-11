# Distributed calculator - The final part of the project
### Overview
- Persistence - the ability to save the program state and restore it after reboot.
- Multi-user mode - all operations are performed in the context of a specific user, data is stored in a DBMS.

### Installation and launch recommendations
For a convenient and comfortable development and launch of the project, I recommend:

- Install [Go](https://go.dev/doc/install ) is the official programming language in which the backend is written.
- Install [Visual Studio Code](https://code .visualstudio.com /) is a convenient Go-enabled code editor.
- Install [PostgreSQL](postgresql.org ) is a database management system that is used in the project for storing and processing data.
- Install [Git](https://git-scm.com/downloads ) - to clone the project repository.After installation, check the correctness with the command:
`git --version`
- Install [Postman](https://www.postman.com/downloads /) - for convenient testing and debugging of our project's API
- Install the **Live Server** extension in Visual Studio Code to automatically start the local server and update the page when code changes.

### Installing Go
1. Go to the official website: [https://go.dev/doc/install ]
2. Select your operating system and download the installation package.
3. Follow the instructions of the installer.
  4. After installation, add the path `/usr/local/go/bin` (Linux/macOS/Windows) or the Go installation path to the environment variable `PATH`.
5. Check the installation by opening the terminal and running the command:
     ```
     go version
     ```
     You should see the installed version of Go.

 ### Installing PostgreSQL
1. Go to the official website: [https://www.postgresql.org/download /]
2. Select your operating system.
  3. Download and install PostgreSQL following the instructions of the installer.
  4. After installation, configure the user and database for the project.
5. Make sure that the PostgreSQL server is running.    
 
  ### Installing Git
1. Go to the official website: [https://git-scm.com/downloads ](https://git-scm.com/downloads
2. Download and install Git for your OS.
  3. Check the installation by opening the terminal and running the command:
     ```
     git --version
     ```
     You should see the installed version of Git.

### Installing a Live Server in VS Code
1. Open VS Code.
2. Go to the extensions section (the square icon on the left or `Ctrl+Shift+X`).
3. Enter `Live Server` in the search.
4. Press the button **Install**.
5. After installation, the **Go Live** button will appear in the lower right corner - click it to start the server.

### Postman Installation
1. Go to the official website: [https://www.postman.com/downloads /]
2. Select the version for your operating system (Windows, macOS, Linux).
3. Download and install the application.
4. Launch Postman and create an account (recommended for syncing).

## Project structure
The project consists of two main parts:

- **[Frontend](./frontend/README(ru).md)** is the client part implemented in HTML, CSS and JavaScript. The frontend folder contains detailed documentation on installing and running the frontend.

- **Backend** is a server part written in Go that provides application logic, database management, and API.

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
### Main components(./frontend/README(ru).md)**
### The main components of the backend:
- **calc1/** and **calc2/**  
  Two separate calculator services that perform arithmetic calculations. Each is implemented as a separate application with a `main.go` file.
- **orchestrator/**  
   The orchestrator works as the main dispatcher: it accepts examples from users, distributes them between calculators, monitors the load, saves the results and checks access rights.
 - **utility/**  
  Auxiliary packages used in all backend services:
  - `calculation/` - contains the logic of calculations and algorithms for processing arithmetic expressions.
  - `database/` - implements work with the PostgreSQL database, including storing and updating data about calculations and users.
  - `models/` - describes the data structures and models used in the project (for example, models of calculations, users, etc.).

### Getting started
### Copying a project from GitHub

To start working with the project, you need to clone the repository to a local computer.  
Open a terminal and run the command:
`git clone https://github.com/ruslan709/distributed-calculator.git`

After cloning the repository, go to the project folder to execute the following commands:
`cd distributed-calculator`

### Installing dependencies
To install all the dependencies of the project, run the command in the project directory:
`go mod tidy`
This command will download and install all the necessary dependencies specified in the `go.mod` file.

### Instructions for launching the project 

### Launching backend services
The project includes several backend services: an orchestrator and two calculators (`calc1` and `calc2`).  
To work correctly, they must be run **in separate terminal windows or tabs**.

#### Starting the orchestrator
Open a new terminal, navigate to the orchestrator folder, and launch the service.:
`cd backend/orchestrator`
`go run main.go`

#### Launching the first calculator
In another terminal, go to the folder of the first calculator and run it:
`cd backend/calc1`
`go run main.go`

#### Launching the second calculator
In the third terminal, go to the folder of the second calculator and run it:
`cd backend/calc2`
`go run main.go`

After launching all the services, the backend will be ready to work.  
To launch the frontend, use the instructions from the **[Frontend] folder(./frontend/README(ru).md)**

#### Getting the status of an orchestrator

```bash
curl -X GET http://localhost:8080/orchestrator-status```

Server response example:
``json
{
"running": true,
  "message": "Orchestrator is running"
}
```

#### Getting calculator server statuses

```bash
curl -X GET http://localhost:8080/ping-servers
```

Server response example:
``json
[
{
"url": "http://localhost:8081 ",
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

To send a computational task to the calculator's server, use the following `curl` request:

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

Server response example:
``json
{
"message": "Calculation started successfully."
}
``

#### Getting the current number of goroutines
```bash
curl -X GET http://localhost:8081/goroutines
```

Server response example:
```plaintext
Current number of goroutines: 1
```

#### Checking the status of the calculator server
```bash
curl -X GET http://localhost:8081/ping
```

Server response example:
``json
{
"status": "running",
  "maxGoroutines": 5,
  "currentGoroutines": 1
}
```

#### Stopping the calculator server
```bash
curl -X POST http://localhost:8081/shutdown
```

Server response example:
```plaintext
Server is shutting down...
```
## Interface
![Illustration for the project]()
# Thank you for your interest in the Distributed Calculator project! # 
This project demonstrates the principles of building scalable, fault-tolerant and multi-user distributed systems in the Go language using modern technologies: gRPC, PostgreSQL, JWT authorization and microservice architecture.
# Special thanks to the Go development course team for their support and useful materials.