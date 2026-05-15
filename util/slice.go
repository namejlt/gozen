package util

// Slice provides generic-style slice helpers.

// SliceContain checks if a string slice contains a value.
func SliceContain(slice []string, val string) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}

// SliceIntContain checks if an int slice contains a value.
func SliceIntContain(slice []int, val int) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}

// SliceInt64Contain checks if an int64 slice contains a value.
func SliceInt64Contain(slice []int64, val int64) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}

// SliceUnique deduplicates a string slice while preserving order.
func SliceUnique(slice []string) []string {
	seen := make(map[string]struct{}, len(slice))
	out := make([]string, 0, len(slice))
	for _, v := range slice {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			out = append(out, v)
		}
	}
	return out
}

// SliceMapString applies fn to each element and returns the results.
func SliceMapString(slice []string, fn func(string) string) []string {
	out := make([]string, len(slice))
	for i, v := range slice {
		out[i] = fn(v)
	}
	return out
}
