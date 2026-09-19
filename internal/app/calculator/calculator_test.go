package calculator

import (
	"math"
	"testing"
)

func TestEvaluate(t *testing.T) {
	tests := []struct {
		expression string
		want       float64
	}{
		{"15 * 32", 480},
		{"(12 + 8) * 3", 60},
		{"100 / 4 + 2.5", 27.5},
		{"20 % 6", 2},
		{"-2 * (4 + 1)", -10},
		{"2 ^ 3 ^ 2", 512},
	}

	for _, test := range tests {
		t.Run(test.expression, func(t *testing.T) {
			got, err := Evaluate(test.expression)
			if err != nil {
				t.Fatalf("Evaluate() error = %v", err)
			}
			if math.Abs(got-test.want) > 1e-9 {
				t.Fatalf("Evaluate() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestEvaluateRejectsInvalidExpressions(t *testing.T) {
	for _, expression := range []string{"", "1 +", "(1 + 2", "1 / 0", "hello"} {
		t.Run(expression, func(t *testing.T) {
			if _, err := Evaluate(expression); err == nil {
				t.Fatalf("Evaluate(%q) succeeded", expression)
			}
		})
	}
}
