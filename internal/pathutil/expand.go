package pathutil

import (
	"os"
	"regexp"
	"strings"
)

var windowsEnvPattern = regexp.MustCompile(`%([^%]+)%`)

// ExpandPathVariables expands both $VAR/${VAR} and %VAR% styles.
func ExpandPathVariables(input string) string {
	expanded := os.ExpandEnv(input)

	return windowsEnvPattern.ReplaceAllStringFunc(expanded, func(token string) string {
		name := strings.TrimSuffix(strings.TrimPrefix(token, "%"), "%")
		value, ok := os.LookupEnv(name)
		if !ok {
			return token
		}
		return value
	})
}
