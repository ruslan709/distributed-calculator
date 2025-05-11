package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	pb "calculatorapi/proto/calculator"
	"calculatorapi/utility/calculation"
	"calculatorapi/utility/database"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	httpPort = ":8081"
	port     = ":50051"
)

type server struct {
	pb.UnimplementedCalculatorServiceServer
}

type OperationRequest struct {
	ID        int            `json:"id"`
	Operation string         `json:"operation"`
	Times     map[string]int `json:"times"`
}

var (
	maxGoroutines     = 5
	currentGoroutines = 0
	mu                sync.Mutex
	shutdownCh        = make(chan struct{})
	serverRunning     = true
)

func ConvertOperationTimes(times map[string]int) calculation.OperationTimes {
	operationTimes := calculation.OperationTimes{}
	for k, v := range times {
		switch k {
		case "add_duration":
			operationTimes["+"] = time.Duration(v) * time.Second
		case "subtract_duration":
			operationTimes["-"] = time.Duration(v) * time.Second
		case "multiply_duration":
			operationTimes["*"] = time.Duration(v) * time.Second
		case "divide_duration":
			operationTimes["/"] = time.Duration(v) * time.Second
		}
	}
	return operationTimes
}

func startCalculation(db *sql.DB, id int, operation string, times map[string]int) {
	convertedTimes := ConvertOperationTimes(times)

	go func() {
		defer func() {
			mu.Lock()
			currentGoroutines--
			mu.Unlock()
		}()

		err := database.UpdateCalculationStatusToWork(db, id)
		if err != nil {
			fmt.Printf("Error updating status to work: %v\n", err)
			return
		}

		operations, result := calculation.EvaluateOperation(operation, convertedTimes)
		for _, op := range operations {
			fmt.Println(op)
		}
		fmt.Printf("Calculation ID %d completed. Result: %.6f\n", id, result)

		err = database.UpdateCalculation(db, id, result, "completed")
		if err != nil {
			fmt.Printf("Error updating calculation record to completed: %v\n", err)
		}
	}()
}

func convertToIntMap(input map[string]int32) map[string]int {
	output := make(map[string]int)
	for key, value := range input {
		output[key] = int(value)
	}
	return output
}

func (s *server) PerformCalculation(ctx context.Context, req *pb.CalculationRequest) (*pb.CalculationResponse, error) {
	mu.Lock()

	if !serverRunning {
		mu.Unlock()
		return nil, status.Error(codes.Unavailable, "Server is shutting down")
	}

	if currentGoroutines >= maxGoroutines {
		mu.Unlock()
		return nil, status.Error(codes.ResourceExhausted, "Server max capacity reached")
	}

	currentGoroutines++

	mu.Unlock()

	db := database.GetDB()
	startCalculation(db, int(req.Id), req.Operation, convertToIntMap(req.Times))

	return &pb.CalculationResponse{Id: req.Id}, nil
}

func main() {
	database.InitializeDB()

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterCalculatorServiceServer(grpcServer, &server{})
	fmt.Printf("gRPC server is starting on port %s...\n", port)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	http.HandleFunc("/calculate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method is not supported.", http.StatusNotFound)
			return
		}

		mu.Lock()
		if !serverRunning {
			http.Error(w, "Server is shutting down", http.StatusServiceUnavailable)
			mu.Unlock()
			return
		}
		if currentGoroutines >= maxGoroutines {
			http.Error(w, "Server max capacity reached", http.StatusTooManyRequests)
			mu.Unlock()
			return
		}
		currentGoroutines++
		mu.Unlock()

		var request OperationRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		db := database.GetDB()
		startCalculation(db, request.ID, request.Operation, request.Times)
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintln(w, "Calculation started successfully.")
	})

	http.HandleFunc("/goroutines", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		fmt.Fprintf(w, "Current number of goroutines: %d\n", currentGoroutines)
		mu.Unlock()
	})

	http.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		serverRunning = false
		mu.Unlock()
		close(shutdownCh)
		fmt.Fprintln(w, "Server is shutting down...")
	})

	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		response := struct {
			Status            string `json:"status"`
			MaxGoroutines     int    `json:"maxGoroutines"`
			CurrentGoroutines int    `json:"currentGoroutines"`
		}{
			Status:            "running",
			MaxGoroutines:     maxGoroutines,
			CurrentGoroutines: currentGoroutines,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	go func() {
		<-shutdownCh
		fmt.Println("Server stopped accepting new requests. Waiting for ongoing operations to complete...")
		for {
			mu.Lock()
			if currentGoroutines == 0 {
				mu.Unlock()
				break
			}
			mu.Unlock()
			time.Sleep(1 * time.Second)
		}
		log.Fatal("Server gracefully shut down")
	}()

	fmt.Printf("Calculator server is starting on port %s...\n", httpPort)
	log.Fatal(http.ListenAndServe(httpPort, nil))
}
