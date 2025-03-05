//go:build !solution

package speller

import (
	"strings"
)

var (
	numbers   = []string{"zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten", "eleven", "twelve", "thirteen", "fourteen", "fifteen", "sixteen", "seventeen", "eighteen", "nineteen"}
	tens      = []string{"", "", "twenty", "thirty", "forty", "fifty", "sixty", "seventy", "eighty", "ninety"}
	thousands = []string{"", "thousand", "million", "billion"}
)

func Spell(n int64) string {
	if n == 0 {
		return "zero"
	}

	if n < 0 {
		return "minus " + Spell(-n)
	}

	words := ""
	chunkCount := 0

	for n > 0 {
		if n%1000 != 0 {
			words = spellChunk(n%1000) + thousands[chunkCount] + " " + words
		}
		n /= 1000
		chunkCount++
	}

	return strings.TrimSpace(words)
}

func spellChunk(n int64) string {
	if n == 0 {
		return ""
	} else if n < 20 {
		return numbers[n] + " "
	} else if n < 100 {
		if n%10 == 0 {
			return tens[n/10] + " "
		}
		return tens[n/10] + "-" + numbers[n%10] + " "
	} else {
		return numbers[n/100] + " hundred " + spellChunk(n%100)
	}
}
