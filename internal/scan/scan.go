package scan

import (
	"path/filepath"
	"strings"

	"joyfuI/game-save-backup-manager/internal/model"
	"joyfuI/game-save-backup-manager/internal/pathutil"
	"joyfuI/game-save-backup-manager/internal/regutil"
)

func SaveLocations(locations []model.SaveLocation) []model.ScanResult {
	results := make([]model.ScanResult, 0, len(locations))

	for _, loc := range locations {
		switch strings.ToLower(strings.TrimSpace(loc.Type)) {
		case "reg":
			exists, err := regutil.PathExists(loc.Path)
			if err != nil {
				results = append(results, model.ScanResult{SaveLocation: loc, Installed: false, Matches: 0})
				continue
			}
			matches := 0
			if exists {
				matches = 1
			}
			results = append(results, model.ScanResult{
				SaveLocation: loc,
				Installed:    exists,
				Matches:      matches,
			})
		default:
			expandedPath := pathutil.ExpandPathVariables(loc.Path)
			matches, err := filepath.Glob(expandedPath)
			if err != nil {
				results = append(results, model.ScanResult{SaveLocation: loc, Installed: false, Matches: 0})
				continue
			}

			results = append(results, model.ScanResult{
				SaveLocation: loc,
				Installed:    len(matches) > 0,
				Matches:      len(matches),
			})
		}
	}

	return results
}
