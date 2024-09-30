package util

import "testing"

func TestConversions(t *testing.T) {
	tests := []struct {
		col string
		num int
	}{
		{"A", 1},
		{"B", 2},
		{"Z", 26},
		{"AA", 27},
		{"AB", 28},
		{"AZ", 52},
		{"BA", 53},
		{"ZZ", 702},
		{"AAA", 703},
		{"AAB", 704},
		{"ABA", 729},
		{"BAA", 1379},
		{"JAZZ", 177138},
		{"YEET", 442930},
	}
	for _, test := range tests {
		if got := ColumnLettersToNumber(test.col); got != test.num {
			t.Errorf("ColumnLettersToNumber(%q) = %v, want %v", test.col, got, test.num)
		}
		if got := ColumnNumberToLetters(test.num); got != test.col {
			t.Errorf("ColumnNumberToLetters(%v) = %q, want %q", test.num, got, test.col)
		}
	}
}
