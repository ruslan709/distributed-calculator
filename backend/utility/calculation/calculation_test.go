package calculation

import (
	"strings"
	"testing"
)

func TestParseOperation(t *testing.T) {
	tests := []struct {
		name          string
		operation     string
		wantOperands  []string
		wantOperators []string
	}{
		{
			name:          "Simple Addition",
			operation:     "3 + 4",
			wantOperands:  []string{"3", "4"},
			wantOperators: []string{"+"},
		},
		{
			name:          "Complex Operation",
			operation:     "5 + 6 * 3",
			wantOperands:  []string{"5", "6", "3"},
			wantOperators: []string{"+", "*"},
		},
		{
			name:          "Operation With Spaces",
			operation:     "12    /   4 - 1",
			wantOperands:  []string{"12", "4", "1"},
			wantOperators: []string{"/", "-"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operands, operators := parseOperation(tt.operation)
			trimmedOperands := trimSpacesFromOperands(operands)
			if !equalSlices(trimmedOperands, tt.wantOperands) {
				t.Errorf("parseOperation() got operands = %v, want %v", operands, tt.wantOperands)
			}
			if !equalSlices(operators, tt.wantOperators) {
				t.Errorf("parseOperation() got operators = %v, want %v", operators, tt.wantOperators)
			}
		})
	}
}

// Helper function to compare slices
func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Helper function to remove spaces from operands for testing
func trimSpacesFromOperands(operands []string) []string {
	trimmed := make([]string, len(operands))
	for i, op := range operands {
		trimmed[i] = strings.TrimSpace(op)
	}
	return trimmed
}
