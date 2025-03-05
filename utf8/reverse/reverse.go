//go:build !solution

package reverse

import (
	"strings"
)

func readUTF8Sequence(input string, i int) (int, bool) {
	readByteCount := 0
	UTF8SequenceIsCorrect := true

	if input[i]&0b10000000 == 0b00000000 {
		readByteCount = 1
	} else if input[i]&0b11100000 == 0b11000000 {
		readByteCount = 2
	} else if input[i]&0b11110000 == 0b11100000 {
		readByteCount = 3
	} else if input[i]&0b11111000 == 0b11110000 {
		readByteCount = 4
	} else {
		return 0, false
	}

	if i+readByteCount > len(input) {
		UTF8SequenceIsCorrect = false
	}

	for j := 1; j < readByteCount; j++ {
		if i+j >= len(input) || input[i+j]&0b11000000 != 0b10000000 {
			UTF8SequenceIsCorrect = false
			break
		}
	}

	return readByteCount, UTF8SequenceIsCorrect
}

func calcReversedByteCount(input string) int {
	reversedByteCount := 0
	for i := 0; i < len(input); {
		readByteCount, UTF8SequenceIsCorrect := readUTF8Sequence(input, i)
		if UTF8SequenceIsCorrect {
			i += readByteCount
			reversedByteCount++
		} else {
			i++
			reversedByteCount += 3
		}
	}
	return reversedByteCount
}

func checkUTF8GraphemSuffix(suffix byte) bool {
	return suffix&0b11000000 == 0b10000000
}

func isASCII(char byte) bool {
	return char&0b10000000 == 0b00000000
}

func Reverse(input string) string {
	var result strings.Builder
	result.Grow(calcReversedByteCount(input))

	for i := len(input) - 1; i >= 0; {
		if isASCII(input[i]) {
			result.WriteByte(input[i])
			i--
		} else {
			start := i
			for start > 0 && checkUTF8GraphemSuffix(input[start-1]) {
				start--
			}
			if start > 0 {
				start--
			}
			readByteCount, isCorrect := readUTF8Sequence(input, start)
			if isCorrect && start+readByteCount == i+1 {
				result.WriteString(input[start : i+1])
				i = start - 1
			} else {
				result.WriteRune('\uFFFD')
				i--
			}
		}
	}
	return result.String()
}
