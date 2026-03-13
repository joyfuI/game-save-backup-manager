package pathutil

import "strings"

// ReplaceTokenInsensitive replaces all token matches in input, case-insensitively.
func ReplaceTokenInsensitive(input, token, replacement string) string {
	lowerInput := strings.ToLower(input)
	lowerToken := strings.ToLower(token)

	if !strings.Contains(lowerInput, lowerToken) {
		return input
	}

	var builder strings.Builder
	for {
		idx := strings.Index(lowerInput, lowerToken)
		if idx < 0 {
			builder.WriteString(input)
			break
		}

		builder.WriteString(input[:idx])
		builder.WriteString(replacement)
		input = input[idx+len(token):]
		lowerInput = lowerInput[idx+len(token):]
	}

	return builder.String()
}
