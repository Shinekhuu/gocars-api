package utils

import "unicode"

func SplitModelName(model string) string {
	letters := ""
	for _, r := range model {
		if unicode.IsLetter(r) {
			letters += string(r)
		} else {
			break
		}
	}
	return letters
}
