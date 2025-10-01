package utils

import (
	"reflect"
	"testing"
)

func TestGetOrDefaultWithNilValue(t *testing.T) {
	result := GetOrDefault(nil, "default")
	expected := "default"

	if result != expected {
		t.Errorf("GetOrDefault(nil, %q) = %v, expected %v", expected, result, expected)
	}
}

func TestGetOrDefaultWithEmptyString(t *testing.T) {
	result := GetOrDefault("", "default")
	expected := "default"

	if result != expected {
		t.Errorf("GetOrDefault(%q, %q) = %v, expected %v", "", expected, result, expected)
	}
}

func TestGetOrDefaultWithZeroInt(t *testing.T) {
	result := GetOrDefault(0, 42)
	expected := 42

	if result != expected {
		t.Errorf("GetOrDefault(0, %d) = %v, expected %v", expected, result, expected)
	}
}

func TestGetOrDefaultWithValidString(t *testing.T) {
	value := "valid_value"
	result := GetOrDefault(value, "default")
	expected := "valid_value"

	if result != expected {
		t.Errorf("GetOrDefault(%q, %q) = %v, expected %v", value, "default", result, expected)
	}
}

func TestGetOrDefaultWithValidInt(t *testing.T) {
	value := 123
	result := GetOrDefault(value, 42)
	expected := 123

	if result != expected {
		t.Errorf("GetOrDefault(%d, %d) = %v, expected %v", value, 42, result, expected)
	}
}

func TestGetOrDefaultWithValidFloat(t *testing.T) {
	value := 3.14
	result := GetOrDefault(value, 1.0)
	expected := 3.14

	if result != expected {
		t.Errorf("GetOrDefault(%f, %f) = %v, expected %v", value, 1.0, result, expected)
	}
}

func TestGetOrDefaultWithValidBool(t *testing.T) {
	value := true
	result := GetOrDefault(value, false)
	expected := true

	if result != expected {
		t.Errorf("GetOrDefault(%t, %t) = %v, expected %v", value, false, result, expected)
	}
}

func TestGetOrDefaultWithValidSlice(t *testing.T) {
	value := []string{"item1", "item2"}
	defaultValue := []string{"default"}
	result := GetOrDefault(value, defaultValue)

	if !reflect.DeepEqual(result, value) {
		t.Errorf("GetOrDefault(%v, %v) = %v, expected %v", value, defaultValue, result, value)
	}
}

func TestGetOrDefaultWithValidMap(t *testing.T) {
	value := map[string]int{"key1": 1}
	defaultValue := map[string]int{"default": 0}
	result := GetOrDefault(value, defaultValue)

	if !reflect.DeepEqual(result, value) {
		t.Errorf("GetOrDefault(%v, %v) = %v, expected %v", value, defaultValue, result, value)
	}
}

func TestGetOrDefaultWithValidStruct(t *testing.T) {
	type TestStruct struct {
		Name string
		Age  int
	}

	value := TestStruct{Name: "John", Age: 30}
	defaultValue := TestStruct{Name: "Default", Age: 0}
	result := GetOrDefault(value, defaultValue)

	if !reflect.DeepEqual(result, value) {
		t.Errorf("GetOrDefault(%v, %v) = %v, expected %v", value, defaultValue, result, value)
	}
}

func TestGetOrDefaultWithNegativeInt(t *testing.T) {
	value := -5
	result := GetOrDefault(value, 42)
	expected := -5

	if result != expected {
		t.Errorf("GetOrDefault(%d, %d) = %v, expected %v", value, 42, result, expected)
	}
}

func TestGetOrDefaultWithZeroFloat(t *testing.T) {
	value := 0.0
	result := GetOrDefault(value, 1.5)
	expected := 0.0 // 0.0 (float64) não é igual a 0 (int) na comparação interface{}, então retorna o valor original

	if result != expected {
		t.Errorf("GetOrDefault(%f, %f) = %v, expected %v", value, 1.5, result, expected)
	}
}

// Benchmark tests
func BenchmarkGetOrDefaultWithValidValue(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetOrDefault("valid_value", "default")
	}
}

func BenchmarkGetOrDefaultWithDefaultValue(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetOrDefault("", "default")
	}
}

// Table-driven tests
func TestGetOrDefaultTableDriven(t *testing.T) {
	tests := []struct {
		name         string
		value        interface{}
		defaultValue interface{}
		expected     interface{}
	}{
		{"nil value", nil, "default", "default"},
		{"empty string", "", "default", "default"},
		{"zero int", 0, 42, 42},
		{"valid string", "hello", "default", "hello"},
		{"valid int", 123, 42, 123},
		{"valid float", 3.14, 1.0, 3.14},
		{"valid bool true", true, false, true},
		{"valid bool false", false, true, false},
		{"negative int", -10, 5, -10},
		{"zero float", 0.0, 1.5, 0.0}, // 0.0 (float64) não é igual a 0 (int) na comparação interface{}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetOrDefault(tt.value, tt.defaultValue)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("GetOrDefault(%v, %v) = %v, expected %v", tt.value, tt.defaultValue, result, tt.expected)
			}
		})
	}
}
