package insensitiveopt

import (
	"strings"
	"unicode"
)

var insensitive = true

func Insensitive(f bool) {
	insensitive = f
}

func ToLower(s string) string {
	if insensitive {
		return strings.ToLower(s)
	}

	return s
}

func ToLowerRune(s rune) rune {
	if insensitive {
		return unicode.ToLower(s)
	}

	return s
}
