package util

import (
	"strconv"
)

// Conv provides type-conversion helpers.  All functions return zero-values
// on parse failure rather than panicking.

// Atoi is a shorthand for strconv.Atoi that returns 0 on error.
func Atoi(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}

// Atoi64 is strconv.ParseInt with a default of 0.
func Atoi64(s string) int64 {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return n
}

// Atoui64 is strconv.ParseUint with a default of 0.
func Atoui64(s string) uint64 {
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0
	}
	return n
}

// Atof64 is strconv.ParseFloat with a default of 0.
func Atof64(s string) float64 {
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return n
}

// Atob is strconv.ParseBool.  Returns false on error.
func Atob(s string) bool {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return false
	}
	return b
}
