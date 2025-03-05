//go:build !solution

package varfmt

import (
	"fmt"
	"strings"
)

func Sprintf(format string, args ...interface{}) string {
	if len(args) == 0 {
		return format
	}

	var result strings.Builder
	start := 0
	unnumberedIndex := 0

	for i := 0; i < len(format); i++ {
		if format[i] == '{' {
			result.WriteString(format[start:i])
			start = i + 1

			for i < len(format) && format[i] != '}' {
				i++
			}
			if i >= len(format) {
				return format
			}

			placeholder := format[start:i]
			start = i + 1

			if placeholder != "" {
				var index int
				_, err := fmt.Sscanf(placeholder, "%d", &index)
				if err == nil {
					if index >= 0 && index < len(args) {
						result.WriteString(fmt.Sprint(args[index]))
					}
					unnumberedIndex++
					continue
				}
			}

			if unnumberedIndex < len(args) {
				result.WriteString(fmt.Sprint(args[unnumberedIndex]))
			}
			unnumberedIndex++
		}
	}

	if start < len(format) {
		result.WriteString(format[start:])
	}

	return result.String()
}
