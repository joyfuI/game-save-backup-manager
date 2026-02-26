package backup

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"joyfuI/game-save-backup-manager/internal/appsettings"
	"joyfuI/game-save-backup-manager/internal/model"
	"joyfuI/game-save-backup-manager/internal/pathutil"
)

type Result struct {
	ZipPath string
	Matched int
	Written int
	Skipped int
}

func CreateZipBackup(loc model.SaveLocation, backupDir string) (Result, error) {
	resolvedPath, err := appsettings.ResolveSavePath(loc.Path)
	if err != nil {
		return Result{}, err
	}

	matches, err := pathutil.Glob(resolvedPath)
	if err != nil {
		return Result{}, fmt.Errorf("glob failed: %w", err)
	}
	if len(matches) == 0 {
		return Result{}, fmt.Errorf("백업할 세이브 파일을 찾지 못했습니다")
	}

	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return Result{}, fmt.Errorf("backup dir create failed: %w", err)
	}

	zipPath := filepath.Join(backupDir, loc.FileName+".zip")
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return Result{}, fmt.Errorf("zip file create failed: %w", err)
	}
	defer func() {
		_ = zipFile.Close()
	}()

	zw := zip.NewWriter(zipFile)

	baseRoot := fixedPrefixPath(resolvedPath)
	logicalPattern := normalizeLogicalPattern(loc.Path)
	logicalPrefix := restoreLogicalTokens(fixedPrefixPath(logicalPattern))
	if !hasGlobWildcards(logicalPattern) {
		baseRoot = filepath.Dir(resolvedPath)
		logicalPrefix = filepath.Dir(loc.Path)
	}
	written := 0
	skipped := 0

	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			skipped++
			continue
		}

		if info.IsDir() {
			err = filepath.Walk(match, func(path string, fi os.FileInfo, walkErr error) error {
				if walkErr != nil {
					skipped++
					return nil
				}
				if fi.IsDir() {
					return nil
				}
				entryName := archivePath(baseRoot, logicalPrefix, path)
				if addErr := addFileToZip(zw, path, entryName); addErr != nil {
					return addErr
				}
				written++
				return nil
			})
			if err != nil {
				return Result{}, err
			}
			continue
		}

		entryName := archivePath(baseRoot, logicalPrefix, match)
		if err := addFileToZip(zw, match, entryName); err != nil {
			return Result{}, err
		}
		written++
	}

	if written == 0 {
		return Result{}, fmt.Errorf("백업 대상 파일이 없어 zip을 만들지 못했습니다")
	}

	if err := zw.Close(); err != nil {
		return Result{}, fmt.Errorf("zip close failed: %w", err)
	}

	return Result{
		ZipPath: zipPath,
		Matched: len(matches),
		Written: written,
		Skipped: skipped,
	}, nil
}

func addFileToZip(zw *zip.Writer, srcPath, entryName string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("open source file failed (%s): %w", srcPath, err)
	}
	defer src.Close()

	w, err := zw.Create(entryName)
	if err != nil {
		return fmt.Errorf("create zip entry failed (%s): %w", entryName, err)
	}

	if _, err := io.Copy(w, src); err != nil {
		return fmt.Errorf("copy source file failed (%s): %w", srcPath, err)
	}

	return nil
}

func fixedPrefixPath(pattern string) string {
	clean := filepath.Clean(pattern)
	volume := filepath.VolumeName(clean)
	rest := strings.TrimPrefix(clean, volume)
	rest = strings.TrimPrefix(rest, string(filepath.Separator))

	parts := strings.Split(rest, string(filepath.Separator))
	fixed := make([]string, 0, len(parts))
	for _, p := range parts {
		if strings.ContainsAny(p, "*?[{") {
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

func archivePath(baseRoot, logicalPrefix, fullPath string) string {
	rel, err := filepath.Rel(baseRoot, fullPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		rel = filepath.Base(fullPath)
	}
	if rel == "." || rel == "" {
		rel = filepath.Base(fullPath)
	}

	rel = filepath.ToSlash(rel)
	logicalPrefix = strings.Trim(filepath.ToSlash(logicalPrefix), "/")
	if logicalPrefix == "" {
		return rel
	}

	return logicalPrefix + "/" + rel
}

func hasGlobWildcards(path string) bool {
	return strings.ContainsAny(path, "*?[{")
}

func normalizeLogicalPattern(path string) string {
	replaced := replaceTokenInsensitive(path, "{{ubisoftconnect-path}}", "__UBISOFTCONNECT_PATH__")
	return replaceTokenInsensitive(replaced, "{{ubisoftconnect-userid}}", "__UBISOFTCONNECT_USERID__")
}

func restoreLogicalTokens(path string) string {
	replaced := strings.ReplaceAll(path, "__UBISOFTCONNECT_PATH__", "{{ubisoftconnect-path}}")
	return strings.ReplaceAll(replaced, "__UBISOFTCONNECT_USERID__", "{{ubisoftconnect-userid}}")
}

func replaceTokenInsensitive(input, token, replacement string) string {
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
