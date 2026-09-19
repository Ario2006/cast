package units

import (
	"math"
	"testing"
)

func TestConvertAcrossCategories(t *testing.T) {
	tests := []struct {
		fromValue string
		fromUnit  string
		toUnit    string
		want      float64
	}{
		{"10", "km", "mi", 6.2137119223733395},
		{"1", "kg", "lb", 2.2046226218487757},
		{"100", "C", "F", 212},
		{"2", "h", "min", 120},
		{"2.5", "GB", "MB", 2500},
		{"72", "mph", "km/h", 115.872768},
	}

	for _, test := range tests {
		t.Run(test.fromUnit+"_to_"+test.toUnit, func(t *testing.T) {
			input, err := Parse(test.fromValue, test.fromUnit)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			got, err := Convert(input, test.toUnit)
			if err != nil {
				t.Fatalf("Convert() error = %v", err)
			}
			if math.Abs(got.Value-test.want) > 1e-9 {
				t.Fatalf("Convert() = %v, want %v", got.Value, test.want)
			}
		})
	}
}

func TestConvertRejectsIncompatibleUnits(t *testing.T) {
	input, err := Parse("10", "km")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Convert(input, "kg"); err == nil {
		t.Fatal("Convert() succeeded for incompatible units")
	}
}

func TestTargetsExcludesSource(t *testing.T) {
	input, err := Parse("10", "km")
	if err != nil {
		t.Fatal(err)
	}
	for _, target := range Targets(input.Unit) {
		if target.Symbol == "km" {
			t.Fatal("Targets() included source unit")
		}
	}
}
