// Package units defines the offline unit conversions supported by cast.
package units

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

// Category groups units which can be converted between one another.
type Category string

const (
	Length Category = "length"
	Mass   Category = "mass"
	Temp   Category = "temperature"
	Time   Category = "time"
	Data   Category = "data"
	Speed  Category = "speed"
)

// Unit is a named measurement unit. Base conversions are kept inside this
// package so command handlers do not need category-specific conversion logic.
type Unit struct {
	Symbol   string
	Name     string
	Category Category
	toBase   func(float64) float64
	fromBase func(float64) float64
}

// Quantity combines a numeric value with its unit.
type Quantity struct {
	Value float64
	Unit  Unit
}

var all = []Unit{
	scaled("mm", "millimeters", Length, 0.001),
	scaled("cm", "centimeters", Length, 0.01),
	scaled("m", "meters", Length, 1),
	scaled("km", "kilometers", Length, 1000),
	scaled("in", "inches", Length, 0.0254),
	scaled("ft", "feet", Length, 0.3048),
	scaled("yd", "yards", Length, 0.9144),
	scaled("mi", "miles", Length, 1609.344),

	scaled("mg", "milligrams", Mass, 0.001),
	scaled("g", "grams", Mass, 1),
	scaled("kg", "kilograms", Mass, 1000),
	scaled("oz", "ounces", Mass, 28.349523125),
	scaled("lb", "pounds", Mass, 453.59237),

	unit("C", "Celsius", Temp, func(value float64) float64 { return value }, func(value float64) float64 { return value }),
	unit("F", "Fahrenheit", Temp, func(value float64) float64 { return (value - 32) * 5 / 9 }, func(value float64) float64 { return value*9/5 + 32 }),
	unit("K", "Kelvin", Temp, func(value float64) float64 { return value - 273.15 }, func(value float64) float64 { return value + 273.15 }),

	scaled("ms", "milliseconds", Time, 0.001),
	scaled("s", "seconds", Time, 1),
	scaled("min", "minutes", Time, 60),
	scaled("h", "hours", Time, 3600),
	scaled("day", "days", Time, 86400),

	// Data units use decimal (base-10) definitions: 1 KB is 1,000 bytes.
	scaled("B", "bytes", Data, 1),
	scaled("KB", "kilobytes", Data, 1e3),
	scaled("MB", "megabytes", Data, 1e6),
	scaled("GB", "gigabytes", Data, 1e9),
	scaled("TB", "terabytes", Data, 1e12),

	scaled("m/s", "meters per second", Speed, 1),
	scaled("km/h", "kilometers per hour", Speed, 1.0/3.6),
	scaled("mph", "miles per hour", Speed, 0.44704),
	scaled("ft/s", "feet per second", Speed, 0.3048),
}

func scaled(symbol, name string, category Category, factor float64) Unit {
	return unit(symbol, name, category, func(value float64) float64 { return value * factor }, func(value float64) float64 { return value / factor })
}

func unit(symbol, name string, category Category, toBase, fromBase func(float64) float64) Unit {
	return Unit{Symbol: symbol, Name: name, Category: category, toBase: toBase, fromBase: fromBase}
}

// Parse validates a CLI value and unit symbol.
func Parse(value, symbol string) (Quantity, error) {
	number, err := strconv.ParseFloat(value, 64)
	if err != nil || math.IsNaN(number) || math.IsInf(number, 0) {
		return Quantity{}, fmt.Errorf("%q is not a valid number", value)
	}
	selected, err := Find(symbol)
	if err != nil {
		return Quantity{}, err
	}
	return Quantity{Value: number, Unit: selected}, nil
}

// Find returns a unit by its documented symbol.
func Find(symbol string) (Unit, error) {
	for _, candidate := range all {
		if candidate.Symbol == symbol {
			return candidate, nil
		}
	}
	return Unit{}, fmt.Errorf("unsupported unit %q", symbol)
}

// Targets returns compatible conversion targets, ordered as documented and
// excluding the source unit.
func Targets(source Unit) []Unit {
	var targets []Unit
	for _, candidate := range all {
		if candidate.Category == source.Category && candidate.Symbol != source.Symbol {
			targets = append(targets, candidate)
		}
	}
	return targets
}

// Convert changes a quantity to target. The target must be in the same category.
func Convert(source Quantity, targetSymbol string) (Quantity, error) {
	target, err := Find(targetSymbol)
	if err != nil {
		return Quantity{}, err
	}
	if source.Unit.Category != target.Category {
		return Quantity{}, fmt.Errorf("cannot convert %s to %s", source.Unit.Category, target.Category)
	}
	return Quantity{Value: target.fromBase(source.Unit.toBase(source.Value)), Unit: target}, nil
}

// Format returns a compact representation for CLI output.
func Format(quantity Quantity) string {
	value := quantity.Value
	if value == 0 {
		value = 0 // avoid displaying negative zero
	}
	return strconv.FormatFloat(value, 'f', -1, 64) + " " + quantity.Unit.Symbol
}

// Categories is useful for callers that need a deterministic list of available
// unit groups without exposing internal conversion functions.
func Categories() []Category {
	unique := make(map[Category]struct{})
	for _, candidate := range all {
		unique[candidate.Category] = struct{}{}
	}
	categories := make([]Category, 0, len(unique))
	for category := range unique {
		categories = append(categories, category)
	}
	sort.Slice(categories, func(i, j int) bool { return categories[i] < categories[j] })
	return categories
}
