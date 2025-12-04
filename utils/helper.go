package utils

import (
	"strconv"
)

// helper: convert string to int safely
func Atoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

// AtoiUint converts a string to uint safely
func AtoiUint(s string) uint {
	i, err := strconv.ParseUint(s, 10, 64) // parse as unsigned integer
	if err != nil {
		return 0
	}
	return uint(i)
}

func IntValue(i *int) int {
	if i != nil {
		return *i
	}
	return 0
}

func SafeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
