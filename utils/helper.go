package utils

import "strconv"

// helper: convert string to int safely
func Atoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
