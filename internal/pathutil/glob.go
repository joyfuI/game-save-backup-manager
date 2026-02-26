package pathutil

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Glob expands simple brace patterns like "{a,b}" and returns merged glob matches.
func Glob(pattern string) ([]string, error) {
	expanded, err := expandBraces(pattern)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})
	out := make([]string, 0)
	for _, p := range expanded {
		matches, globErr := filepath.Glob(p)
		if globErr != nil {
			return nil, globErr
		}
		for _, m := range matches {
			if _, ok := seen[m]; ok {
				continue
			}
			seen[m] = struct{}{}
			out = append(out, m)
		}
	}

	return out, nil
}

func expandBraces(pattern string) ([]string, error) {
	start := strings.IndexByte(pattern, '{')
	if start < 0 {
		return []string{pattern}, nil
	}

	end := findBraceEnd(pattern, start)
	if end < 0 {
		return nil, fmt.Errorf("invalid brace pattern: %s", pattern)
	}

	inside := pattern[start+1 : end]
	parts := splitTopLevel(inside)
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid brace pattern: %s", pattern)
	}

	prefix := pattern[:start]
	suffix := pattern[end+1:]

	out := make([]string, 0, len(parts))
	for _, part := range parts {
		expandedSuffix, err := expandBraces(suffix)
		if err != nil {
			return nil, err
		}
		for _, s := range expandedSuffix {
			out = append(out, prefix+part+s)
		}
	}

	return out, nil
}

func findBraceEnd(s string, open int) int {
	depth := 0
	for i := open; i < len(s); i++ {
		switch s[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

func splitTopLevel(s string) []string {
	parts := make([]string, 0, 2)
	depth := 0
	start := 0

	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '{':
			depth++
		case '}':
			depth--
		case ',':
			if depth == 0 {
				parts = append(parts, s[start:i])
				start = i + 1
			}
		}
	}

	parts = append(parts, s[start:])
	return parts
}
