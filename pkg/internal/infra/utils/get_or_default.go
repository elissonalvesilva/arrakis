// Package utils provides utility functions for common operations.
package utils

// GetOrDefault returns the default value if the given value is nil, empty string, or zero.
// This function is particularly useful for providing fallback values in configurations.
//
// Parameters:
//   - value: The value to check. Can be any type.
//   - defaultValue: The fallback value to return if value is considered "empty".
//
// Returns:
//   - interface{}: Either the original value or the default value.
//
// Example:
//
//	result := GetOrDefault(userInput, "default_value")
//	timeout := GetOrDefault(configTimeout, 30).(int)
func GetOrDefault(value, defaultValue interface{}) interface{} {
	if value == nil || value == "" || value == 0 {
		return defaultValue
	}
	return value
}
