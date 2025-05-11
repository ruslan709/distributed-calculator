package calculation

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type OperationTimes map[string]time.Duration

func EvaluateOperation(operation string, operationTimes OperationTimes) ([]string, float64) {
	var operations []string
	operands, operators := parseOperation(operation)

	expression := make([]string, 0, len(operands)+len(operators))
	for i, op := range operands {
		expression = append(expression, op)
		if i < len(operators) {
			expression = append(expression, operators[i])
		}
	}

	for i := 0; i < len(expression)-1; i++ {
		if expression[i] == "*" || expression[i] == "/" {
			left, _ := strconv.ParseFloat(expression[i-1], 64)
			right, _ := strconv.ParseFloat(expression[i+1], 64)
			result := performOperation(left, right, expression[i], operationTimes)
			operations = append(operations, fmt.Sprintf("%s %s %s = %.6f", expression[i-1], expression[i], expression[i+1], result))

			expression[i+1] = fmt.Sprintf("%.6f", result)
			expression = append(expression[:i-1], expression[i+1:]...)
			i = i - 2
		}
	}

	var result float64
	if len(expression) > 0 {
		result, _ = strconv.ParseFloat(expression[0], 64)
	}
	for i := 1; i < len(expression); i += 2 {
		right, _ := strconv.ParseFloat(expression[i+1], 64)
		result = performOperation(result, right, expression[i], operationTimes)
		operations = append(operations, fmt.Sprintf("%.6f %s %s = %.6f", result, expression[i], expression[i+1], result))
	}

	return operations, result
}

func parseOperation(operation string) ([]string, []string) {
	operands := strings.FieldsFunc(operation, func(c rune) bool {
		return c == '+' || c == '-' || c == '*' || c == '/'
	})

	operators := make([]string, 0)
	for _, c := range operation {
		if strings.ContainsRune("+-*/", c) {
			operators = append(operators, string(c))
		}
	}

	return operands, operators
}

func performOperation(left, right float64, operator string, operationTimes OperationTimes) float64 {
	if duration, ok := operationTimes[operator]; ok {
		fmt.Printf("Performing %s operation, waiting for %v\n", operator, duration)
		time.Sleep(duration)
	} else {
		fmt.Println("Unknown operation, no delay applied")
	}

	switch operator {
	case "+":
		return left + right
	case "-":
		return left - right
	case "*":
		return left * right
	case "/":
		if right == 0 {
			fmt.Println("Error: Division by zero")
			return 0
		}
		return left / right
	default:
		fmt.Println("Unknown operator", operator)
		return 0
	}
}
