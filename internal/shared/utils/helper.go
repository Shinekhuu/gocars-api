package utils

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

func Atoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func AtoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return i
}

func AtoiUint(s string) uint {
	i, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0
	}
	return uint(i)
}

func StringToUintPtr(s string) *uint {
	if s == "" {
		return nil
	}
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return nil
	}
	u := uint(v)
	return &u
}

func UintPtrOrNilFromUint(v uint) *uint {
	if v == 0 {
		return nil
	}
	return &v
}

func UintToIntPtr(u uint) *int {
	i := int(u)
	return &i
}

func UintPtrToIntPtr(u *uint) *int {
	if u == nil {
		return nil
	}
	i := int(*u)
	return &i
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

func ContainsChinese(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

func ContainsSlashAndSpace(s string) bool {
	return strings.Contains(s, "/") && strings.Contains(s, " ")
}

func SplitIntoEngines(modelStr string) []string {
	engineNames := strings.FieldsFunc(modelStr, func(r rune) bool {
		return r == ' ' || r == '/'
	})
	for i := range engineNames {
		engineNames[i] = strings.TrimSpace(engineNames[i])
		engineNames[i] = strings.ReplaceAll(engineNames[i], "#", "")
	}
	if len(engineNames) == 0 {
		engineNames = []string{modelStr}
	}
	return engineNames
}

func SplitIntoCategoryNames(categoryStr string) []string {
	categories := strings.Split(categoryStr, ">")
	for i := range categories {
		categories[i] = strings.TrimSpace(categories[i])
	}
	return categories
}

func Unique(input []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0)
	for _, v := range input {
		v = strings.TrimSpace(strings.ToUpper(v))
		if v == "" || seen[v] {
			continue
		}
		seen[v] = true
		result = append(result, v)
	}
	return result
}

var cleanRegex = regexp.MustCompile(`[^\p{L}\p{N}\s]`)

func NormalizeString(input string) string {
	if input == "" {
		return ""
	}
	s := strings.TrimSpace(input)
	replacer := strings.NewReplacer("/", " ", "-", " ", "_", " ")
	s = replacer.Replace(s)
	s = cleanRegex.ReplaceAllString(s, "")
	s = strings.ToUpper(s)
	s = strings.Join(strings.Fields(s), " ")
	return s
}

func NormalizeOEM(oem string) (base string, full string) {
	oem = strings.ToUpper(strings.TrimSpace(oem))
	oem = strings.ReplaceAll(oem, " ", "")
	re := regexp.MustCompile(`^(\d{5}-\d{5})`)
	match := re.FindStringSubmatch(oem)
	if len(match) > 1 {
		base = match[1]
		return base, oem
	}
	return oem, oem
}

func IsOEM(oem string) bool {
	return regexp.MustCompile(`^\d{5}-\d{5}`).MatchString(oem)
}
