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

	keySteamPath              = "steam_path"
	keySteamUserID            = "steam_userid"
	keySteamAccountID         = "steam_accountid"
	keyMicrosoftStoreUserID   = "microsoftstore_userid"
	keyRockstarLauncherUserID = "rockstargameslauncher_userid"
	keyUbisoftConnectPath     = "ubisoft_connect_path"
	keyUbisoftConnectUserID   = "ubisoft_connect_userid"

	tokenSteamPath              = "{{steam-path}}"
	tokenSteamUserID            = "{{steam-userid}}"
	tokenSteamAccountID         = "{{steam-accountid}}"
	tokenMicrosoftStoreUserID   = "{{microsoftstore-userid}}"
	tokenRockstarLauncherUserID = "{{rockstargameslauncher-userid}}"
	tokenUbisoftConnectPath     = "{{ubisoftconnect-path}}"
	tokenUbisoftConnectUserID   = "{{ubisoftconnect-userid}}"
)

type Settings struct {
	SteamPath              string
	SteamUserID            string
	SteamAccountID         string
	MicrosoftStoreUserID   string
	RockstarLauncherUserID string
	UbisoftConnectPath     string
	UbisoftConnectUserID   string
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
		case keySteamAccountID:
			settings.SteamAccountID = strings.TrimSpace(value)
		case keyMicrosoftStoreUserID:
			settings.MicrosoftStoreUserID = strings.TrimSpace(value)
		case keyRockstarLauncherUserID:
			settings.RockstarLauncherUserID = strings.TrimSpace(value)
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
	steamAccountIDValue := strings.TrimSpace(settings.SteamAccountID)
	microsoftStoreUserIDValue := strings.TrimSpace(settings.MicrosoftStoreUserID)
	rockstarLauncherUserIDValue := strings.TrimSpace(settings.RockstarLauncherUserID)
	ubisoftUserIDValue := strings.TrimSpace(settings.UbisoftConnectUserID)

	content := fmt.Sprintf("[settings]\n%s=%s\n%s=%s\n%s=%s\n%s=%s\n%s=%s\n%s=%s\n%s=%s\n",
		keySteamPath, steamPathValue,
		keySteamUserID, steamUserIDValue,
		keySteamAccountID, steamAccountIDValue,
		keyUbisoftConnectPath, ubisoftPathValue,
		keyUbisoftConnectUserID, ubisoftUserIDValue,
		keyRockstarLauncherUserID, rockstarLauncherUserIDValue,
		keyMicrosoftStoreUserID, microsoftStoreUserIDValue,
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
		SteamPath:              DefaultSteamPath(),
		SteamUserID:            "",
		SteamAccountID:         "",
		MicrosoftStoreUserID:   "",
		RockstarLauncherUserID: "",
		UbisoftConnectPath:     DefaultUbisoftConnectPath(),
		UbisoftConnectUserID:   "",
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
		resolved = pathutil.ReplaceTokenInsensitive(resolved, tokenSteamPath, filepath.Clean(steamPath))
	}

	if strings.Contains(strings.ToLower(resolved), tokenSteamUserID) {
		steamUserID := strings.TrimSpace(settings.SteamUserID)
		if steamUserID == "" {
			return "", fmt.Errorf("steam userid setting is empty")
		}
		resolved = pathutil.ReplaceTokenInsensitive(resolved, tokenSteamUserID, steamUserID)
	}
	if strings.Contains(strings.ToLower(resolved), tokenSteamAccountID) {
		steamAccountID := strings.TrimSpace(settings.SteamAccountID)
		if steamAccountID == "" {
			return "", fmt.Errorf("steam accountid setting is empty")
		}
		resolved = pathutil.ReplaceTokenInsensitive(resolved, tokenSteamAccountID, steamAccountID)
	}
	if strings.Contains(strings.ToLower(resolved), tokenMicrosoftStoreUserID) {
		microsoftStoreUserID := strings.TrimSpace(settings.MicrosoftStoreUserID)
		if microsoftStoreUserID == "" {
			return "", fmt.Errorf("microsoft store userid setting is empty")
		}
		resolved = pathutil.ReplaceTokenInsensitive(resolved, tokenMicrosoftStoreUserID, microsoftStoreUserID)
	}
	if strings.Contains(strings.ToLower(resolved), tokenRockstarLauncherUserID) {
		rockstarLauncherUserID := strings.TrimSpace(settings.RockstarLauncherUserID)
		if rockstarLauncherUserID == "" {
			return "", fmt.Errorf("rockstar games launcher userid setting is empty")
		}
		resolved = pathutil.ReplaceTokenInsensitive(resolved, tokenRockstarLauncherUserID, rockstarLauncherUserID)
	}

	if strings.Contains(strings.ToLower(resolved), tokenUbisoftConnectPath) {
		installPath := strings.TrimSpace(pathutil.ExpandPathVariables(settings.UbisoftConnectPath))
		if installPath == "" {
			return "", fmt.Errorf("ubisoft connect path setting is empty")
		}
		resolved = pathutil.ReplaceTokenInsensitive(resolved, tokenUbisoftConnectPath, filepath.Clean(installPath))
	}

	if strings.Contains(strings.ToLower(resolved), tokenUbisoftConnectUserID) {
		userID := strings.TrimSpace(settings.UbisoftConnectUserID)
		if userID == "" {
			return "", fmt.Errorf("ubisoft connect userid setting is empty")
		}
		resolved = pathutil.ReplaceTokenInsensitive(resolved, tokenUbisoftConnectUserID, userID)
	}

	return pathutil.ExpandPathVariables(resolved), nil
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
