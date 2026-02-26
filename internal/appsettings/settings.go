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
	defaultSteamPathRaw          = `%PROGRAMFILES(X86)%\Steam`
	defaultUbisoftConnectPathRaw = `%PROGRAMFILES(X86)%\Ubisoft\Ubisoft Game Launcher`

	keySteamPath            = "steam_path"
	keySteamUserID          = "steam_userid"
	keyMicrosoftStoreUserID = "microsoftstore_userid"
	keyUbisoftConnectPath   = "ubisoft_connect_path"
	keyUbisoftConnectUserID = "ubisoft_connect_userid"

	tokenSteamPath            = "{{steam-path}}"
	tokenSteamUserID          = "{{steam-userid}}"
	tokenMicrosoftStoreUserID = "{{microsoftstore-userid}}"
	tokenUbisoftConnectPath   = "{{ubisoftconnect-path}}"
	tokenUbisoftConnectUserID = "{{ubisoftconnect-userid}}"
)

type Settings struct {
	SteamPath            string
	SteamUserID          string
	MicrosoftStoreUserID string
	UbisoftConnectPath   string
	UbisoftConnectUserID string
}

func DefaultSteamPath() string {
	return filepath.Clean(pathutil.ExpandPathVariables(defaultSteamPathRaw))
}

func DefaultUbisoftConnectPath() string {
	return filepath.Clean(pathutil.ExpandPathVariables(defaultUbisoftConnectPathRaw))
}

func Load() (Settings, error) {
	settings := Settings{
		SteamPath:          DefaultSteamPath(),
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

		switch strings.ToLower(strings.TrimSpace(key)) {
		case keySteamPath:
			cleaned := filepath.Clean(strings.TrimSpace(value))
			if cleaned != "" {
				settings.SteamPath = cleaned
			}
		case keySteamUserID:
			settings.SteamUserID = strings.TrimSpace(value)
		case keyMicrosoftStoreUserID:
			settings.MicrosoftStoreUserID = strings.TrimSpace(value)
		case keyUbisoftConnectPath:
			cleaned := filepath.Clean(strings.TrimSpace(value))
			if cleaned != "" {
				settings.UbisoftConnectPath = cleaned
			}
		case keyUbisoftConnectUserID:
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

	steamPathValue := filepath.Clean(strings.TrimSpace(pathutil.ExpandPathVariables(settings.SteamPath)))
	if steamPathValue == "" {
		steamPathValue = DefaultSteamPath()
	}

	ubisoftPathValue := filepath.Clean(strings.TrimSpace(pathutil.ExpandPathVariables(settings.UbisoftConnectPath)))
	if ubisoftPathValue == "" {
		ubisoftPathValue = DefaultUbisoftConnectPath()
	}

	steamUserIDValue := strings.TrimSpace(settings.SteamUserID)
	microsoftStoreUserIDValue := strings.TrimSpace(settings.MicrosoftStoreUserID)
	ubisoftUserIDValue := strings.TrimSpace(settings.UbisoftConnectUserID)

	content := fmt.Sprintf("[settings]\n%s=%s\n%s=%s\n%s=%s\n%s=%s\n%s=%s\n",
		keySteamPath, steamPathValue,
		keySteamUserID, steamUserIDValue,
		keyMicrosoftStoreUserID, microsoftStoreUserIDValue,
		keyUbisoftConnectPath, ubisoftPathValue,
		keyUbisoftConnectUserID, ubisoftUserIDValue,
	)
	return os.WriteFile(filePath, []byte(content), 0o644)
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
		SteamPath:            DefaultSteamPath(),
		SteamUserID:          "",
		MicrosoftStoreUserID: "",
		UbisoftConnectPath:   DefaultUbisoftConnectPath(),
		UbisoftConnectUserID: "",
	})
}

func ResolveSavePath(path string) (string, error) {
	settings, err := Load()
	if err != nil {
		return "", err
	}
	return resolveSavePathWithSettings(path, settings)
}

func resolveSavePathWithSettings(path string, settings Settings) (string, error) {
	resolved := path

	if strings.Contains(strings.ToLower(resolved), tokenSteamPath) {
		steamPath := strings.TrimSpace(pathutil.ExpandPathVariables(settings.SteamPath))
		if steamPath == "" {
			return "", fmt.Errorf("steam path setting is empty")
		}
		resolved = replaceTokenInsensitive(resolved, tokenSteamPath, filepath.Clean(steamPath))
	}

	if strings.Contains(strings.ToLower(resolved), tokenSteamUserID) {
		steamUserID := strings.TrimSpace(settings.SteamUserID)
		if steamUserID == "" {
			return "", fmt.Errorf("steam userid setting is empty")
		}
		resolved = replaceTokenInsensitive(resolved, tokenSteamUserID, steamUserID)
	}
	if strings.Contains(strings.ToLower(resolved), tokenMicrosoftStoreUserID) {
		microsoftStoreUserID := strings.TrimSpace(settings.MicrosoftStoreUserID)
		if microsoftStoreUserID == "" {
			return "", fmt.Errorf("microsoft store userid setting is empty")
		}
		resolved = replaceTokenInsensitive(resolved, tokenMicrosoftStoreUserID, microsoftStoreUserID)
	}

	if strings.Contains(strings.ToLower(resolved), tokenUbisoftConnectPath) {
		installPath := strings.TrimSpace(pathutil.ExpandPathVariables(settings.UbisoftConnectPath))
		if installPath == "" {
			return "", fmt.Errorf("ubisoft connect path setting is empty")
		}
		resolved = replaceTokenInsensitive(resolved, tokenUbisoftConnectPath, filepath.Clean(installPath))
	}

	if strings.Contains(strings.ToLower(resolved), tokenUbisoftConnectUserID) {
		userID := strings.TrimSpace(settings.UbisoftConnectUserID)
		if userID == "" {
			return "", fmt.Errorf("ubisoft connect userid setting is empty")
		}
		resolved = replaceTokenInsensitive(resolved, tokenUbisoftConnectUserID, userID)
	}

	return pathutil.ExpandPathVariables(resolved), nil
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
