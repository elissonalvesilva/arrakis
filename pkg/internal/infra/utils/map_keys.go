package utils

// mapKeys extracts the keys from a map[string]string and returns them as a []string.
// This utility function is used to convert message attribute maps to the format expected by SQS.
//
// Parameters:
//   - m: Map containing message attribute names as keys
//
// Returns:
//   - []string: Slice containing all keys from the input map
func MapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
