//go:build windows

package regutil

import (
	"fmt"
	"os/exec"
	"strings"

	"golang.org/x/sys/windows/registry"
)

type TargetKind int

const (
	TargetUnknown TargetKind = iota
	TargetKey
	TargetValue
)

type Target struct {
	Kind      TargetKind
	Root      registry.Key
	RootName  string
	KeyPath   string
	ValueName string
}

type ValueData struct {
	RootName  string
	KeyPath   string
	ValueName string
	Type      uint32
	Data      []byte
}

func PathExists(path string) (bool, error) {
	t, err := ResolveTarget(path)
	if err == registry.ErrNotExist {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	switch t.Kind {
	case TargetKey:
		return keyPathExists(t.Root, t.KeyPath)
	case TargetValue:
		return valuePathExists(t.Root, t.KeyPath, t.ValueName)
	default:
		return false, nil
	}
}

func KeyExists(path string) (bool, error) {
	t, err := ResolveTarget(path)
	if err == registry.ErrNotExist {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if t.Kind != TargetKey {
		return false, nil
	}
	return keyPathExists(t.Root, t.KeyPath)
}

func ValidatePath(path string) error {
	_, _, _, _, err := splitRegistryPath(path)
	return err
}

// ResolveTarget rule:
// - Path ending with '\\' => key target
// - Path not ending with '\\' => value target (last segment is value name)
func ResolveTarget(path string) (Target, error) {
	root, rootName, tail, isKeyMode, err := splitRegistryPath(path)
	if err != nil {
		return Target{}, err
	}

	if isKeyMode {
		return Target{Kind: TargetKey, Root: root, RootName: rootName, KeyPath: tail}, nil
	}

	keyPath, valueName, ok := splitValuePath(tail)
	if !ok {
		return Target{}, fmt.Errorf("값 경로는 '...\\키\\값이름' 형식이어야 합니다")
	}
	return Target{Kind: TargetValue, Root: root, RootName: rootName, KeyPath: keyPath, ValueName: valueName}, nil
}

func ReadValue(path string) (ValueData, error) {
	t, err := ResolveTarget(path)
	if err != nil {
		return ValueData{}, err
	}
	if t.Kind != TargetValue {
		return ValueData{}, fmt.Errorf("레지스트리 값 경로가 아닙니다")
	}

	k, err := registry.OpenKey(t.Root, t.KeyPath, registry.QUERY_VALUE)
	if err != nil {
		return ValueData{}, err
	}
	defer k.Close()

	n, typ, err := k.GetValue(t.ValueName, nil)
	if err != nil {
		return ValueData{}, err
	}

	buf := make([]byte, n)
	n, typ, err = k.GetValue(t.ValueName, buf)
	if err != nil {
		return ValueData{}, err
	}

	return ValueData{
		RootName:  t.RootName,
		KeyPath:   t.KeyPath,
		ValueName: t.ValueName,
		Type:      typ,
		Data:      buf[:n],
	}, nil
}

func OpenInEditor(path string) error {
	t, err := ResolveTarget(path)
	if err != nil {
		return err
	}

	keyForEditor := fmt.Sprintf("%s\\%s", t.RootName, t.KeyPath)

	// Regedit opens the location stored in LastKey.
	k, _, err := registry.CreateKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Applets\Regedit`, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	if err := k.SetStringValue("LastKey", keyForEditor); err != nil {
		return err
	}

	if err := exec.Command("regedit.exe").Start(); err == nil {
		return nil
	}

	// Some systems require elevation to launch regedit directly.
	// Retry via Shell "RunAs" so Windows can show the UAC prompt.
	ps := `Start-Process regedit.exe -Verb RunAs`
	return exec.Command("powershell", "-NoProfile", "-Command", ps).Start()
}

func splitRegistryPath(path string) (registry.Key, string, string, bool, error) {
	p := strings.TrimSpace(strings.ReplaceAll(path, "/", "\\"))
	if p == "" {
		return 0, "", "", false, fmt.Errorf("레지스트리 경로는 필수입니다")
	}

	isKeyMode := strings.HasSuffix(p, "\\")
	p = strings.TrimRight(p, "\\")

	parts := strings.SplitN(p, "\\", 2)
	rootText := strings.ToUpper(strings.TrimSpace(parts[0]))
	if len(parts) < 2 || strings.TrimSpace(parts[1]) == "" {
		return 0, "", "", false, fmt.Errorf("레지스트리 경로는 루트와 하위 경로가 필요합니다")
	}

	var root registry.Key
	var rootName string
	switch rootText {
	case "HKEY_CURRENT_USER", "HKCU":
		root = registry.CURRENT_USER
		rootName = "HKEY_CURRENT_USER"
	case "HKEY_LOCAL_MACHINE", "HKLM":
		root = registry.LOCAL_MACHINE
		rootName = "HKEY_LOCAL_MACHINE"
	case "HKEY_CLASSES_ROOT", "HKCR":
		root = registry.CLASSES_ROOT
		rootName = "HKEY_CLASSES_ROOT"
	case "HKEY_USERS", "HKU":
		root = registry.USERS
		rootName = "HKEY_USERS"
	case "HKEY_CURRENT_CONFIG", "HKCC":
		root = registry.CURRENT_CONFIG
		rootName = "HKEY_CURRENT_CONFIG"
	default:
		return 0, "", "", false, fmt.Errorf("지원하지 않는 레지스트리 루트입니다: %s", parts[0])
	}

	return root, rootName, strings.TrimSpace(parts[1]), isKeyMode, nil
}

func keyPathExists(root registry.Key, keyPath string) (bool, error) {
	k, err := registry.OpenKey(root, keyPath, registry.QUERY_VALUE)
	if err != nil {
		if err == registry.ErrNotExist {
			return false, nil
		}
		return false, err
	}
	_ = k.Close()
	return true, nil
}

func valuePathExists(root registry.Key, keyPath, valueName string) (bool, error) {
	k, err := registry.OpenKey(root, keyPath, registry.QUERY_VALUE)
	if err != nil {
		if err == registry.ErrNotExist {
			return false, nil
		}
		return false, err
	}
	defer k.Close()

	_, _, err = k.GetValue(valueName, nil)
	if err != nil {
		if err == registry.ErrNotExist {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func splitValuePath(tail string) (keyPath, valueName string, ok bool) {
	idx := strings.LastIndex(tail, "\\")
	if idx <= 0 || idx >= len(tail)-1 {
		return "", "", false
	}
	keyPath = strings.TrimSpace(tail[:idx])
	valueName = strings.TrimSpace(tail[idx+1:])
	if keyPath == "" || valueName == "" {
		return "", "", false
	}
	return keyPath, valueName, true
}
