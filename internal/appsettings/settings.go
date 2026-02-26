package appsettings

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"joyfuI/game-save-backup-manager/internal/pathutil"
)

const (
	defaultUbisoftConnectPathRaw = `%PROGRAMFILES(X86)%\Ubisoft\Ubisoft Game Launcher`
	keyUbisoftConnectPath        = "ubisoft_connect_path"
	keyUbisoftConnectUserID      = "ubisoft_connect_user_id"
)

type Settings struct {
	UbisoftConnectPath string
	UbisoftConnectUserID string
}

func DefaultUbisoftConnectPath() string {
	return filepath.Clean(pathutil.ExpandPathVariables(defaultUbisoftConnectPathRaw))
}

func Load() (Settings, error) {
	settings := Settings{
		UbisoftConnectPath: DefaultUbisoftConnectPath(),
	}

	filePath, err := filePath()
	if err != nil {
		return settings, err
	}

	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return settings, nil
		}
		return settings, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "[") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		if strings.EqualFold(strings.TrimSpace(key), keyUbisoftConnectPath) {
			cleaned := filepath.Clean(strings.TrimSpace(value))
			if cleaned != "" {
				settings.UbisoftConnectPath = cleaned
			}
			continue
		}

		if strings.EqualFold(strings.TrimSpace(key), keyUbisoftConnectUserID) {
			settings.UbisoftConnectUserID = strings.TrimSpace(value)
		}
	}

	if err := scanner.Err(); err != nil {
		return settings, err
	}

	return settings, nil
}

func Save(settings Settings) error {
	filePath, err := filePath()
	if err != nil {
		return err
	}

	pathValue := filepath.Clean(strings.TrimSpace(pathutil.ExpandPathVariables(settings.UbisoftConnectPath)))
	if pathValue == "" {
		pathValue = DefaultUbisoftConnectPath()
	}

	userIDValue := strings.TrimSpace(settings.UbisoftConnectUserID)

	content := fmt.Sprintf("[settings]\n%s=%s\n%s=%s\n",
		keyUbisoftConnectPath, pathValue,
		keyUbisoftConnectUserID, userIDValue,
	)
	return os.WriteFile(filePath, []byte(content), 0644)
}

func EnsureInitialized() error {
	filePath, err := filePath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(filePath); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	return Save(Settings{
		UbisoftConnectPath: DefaultUbisoftConnectPath(),
	})
}

func filePath() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}

	base := strings.TrimSuffix(filepath.Base(exePath), filepath.Ext(exePath))
	if base == "" {
		base = "game-save-backup-manager"
	}

	return filepath.Join(filepath.Dir(exePath), base+".ini"), nil
}
