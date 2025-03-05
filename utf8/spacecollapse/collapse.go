//go:build !solution

package spacecollapse

import "strings"

func skipSpace(input string, i int) bool {
	return i+1 < len(input) &&
		(input[i] == ' ' || input[i] == '\t' || input[i] == '\n' || input[i] == '\r') &&
		(input[i+1] == ' ' || input[i+1] == '\t' || input[i+1] == '\n' || input[i+1] == '\r')
}

func replaceSpace(input string, i int) bool {
	return input[i] == '\t' || input[i] == '\n' || input[i] == '\r'
}

func writeUTF8Sequence(input string, i int, result *strings.Builder) int {
	expectedUTFCount := 0
	UTF8SequenceIsCorrect := true
	if input[i]&0b10000000 == 0b00000000 {
		expectedUTFCount = 1
	} else if input[i]&0b11100000 == 0b11000000 {
		expectedUTFCount = 2
	} else if input[i]&0b11110000 == 0b11100000 {
		expectedUTFCount = 3
	} else if input[i]&0b11111000 == 0b11110000 {
		expectedUTFCount = 4
	} else {
		expectedUTFCount = 1
		UTF8SequenceIsCorrect = false
	}

	if i+expectedUTFCount > len(input) {
		UTF8SequenceIsCorrect = false
	}

	for j := 1; j < expectedUTFCount; j++ {
		if i+j >= len(input) || input[i+j]&0b11000000 != 0b10000000 {
			UTF8SequenceIsCorrect = false
			break
		}
	}

	if UTF8SequenceIsCorrect {
		result.WriteString(input[i : i+expectedUTFCount])
	} else {
		result.WriteRune('\uFFFD')
	}

	return expectedUTFCount
}

func CollapseSpaces(input string) string {
	var result strings.Builder
	result.Grow(len(input))

	for i := 0; i < len(input); {
		if skipSpace(input, i) {
			i++
			continue
		}

		if replaceSpace(input, i) {
			result.WriteByte(' ')
			i++
			continue
		}

		i += writeUTF8Sequence(input, i, &result)
	}

	return result.String()
}
