package util

import (
	"strings"
)

func Ptr[T any](v T) *T {
	return &v
}

func Deref[T any](v *T) (r T) {
	if v != nil {
		return *v
	}
	return
}

// ColumnLettersToNumber converts column letters (i.e. AB) to the corresponding number (i.e. 28).
func ColumnLettersToNumber(col string) int {
	col = strings.ToUpper(col)
	num := 0
	for i := 0; i < len(col); i++ {
		num = num*26 + int(col[i]-'A'+1)
	}
	return num
}

// ColumnNumberToLetters converts a column number (i.e. 28) to the corresponding letters (i.e. AB).
func ColumnNumberToLetters(col int) string {
	var res string
	for col > 0 {
		col--
		res = string(rune('A'+col%26)) + res
		col /= 26
	}
	return res
}
