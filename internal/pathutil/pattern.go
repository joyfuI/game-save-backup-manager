package pathutil

import (
	"path/filepath"
	"strings"
)

// HasGlobWildcards reports whether path contains filepath.Match wildcards.
func HasGlobWildcards(path string) bool {
	return strings.ContainsAny(path, "*?[{")
}

// FixedPrefixPath returns the non-wildcard prefix directory of a glob pattern.
func FixedPrefixPath(pattern string) string {
	clean := filepath.Clean(pattern)
	volume := filepath.VolumeName(clean)
	rest := strings.TrimPrefix(clean, volume)
	rest = strings.TrimPrefix(rest, string(filepath.Separator))

	parts := strings.Split(rest, string(filepath.Separator))
	fixed := make([]string, 0, len(parts))
	for _, p := range parts {
		if HasGlobWildcards(p) {
			break
		}
		fixed = append(fixed, p)
	}

	if len(fixed) == 0 {
		if volume != "" {
			return volume + string(filepath.Separator)
		}
		return string(filepath.Separator)
	}

	base := filepath.Join(fixed...)
	if volume != "" {
		return filepath.Join(volume+string(filepath.Separator), base)
	}
	return filepath.Join(string(filepath.Separator), base)
}
