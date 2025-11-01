package utils

import "unicode"

func SplitModelName(model string) string {
	letters := ""
	for _, r := range model {
		if unicode.IsLetter(r) {
			letters += string(r)
		} else {
			// Once first non-letter appears, take the rest as numbers
			break
		}
	}
	return letters
}
