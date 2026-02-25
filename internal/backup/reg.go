package backup

import (
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode/utf16"

	"golang.org/x/sys/windows/registry"
	"joyfuI/game-save-backup-manager/internal/model"
	"joyfuI/game-save-backup-manager/internal/regutil"
)

func CreateRegBackup(loc model.SaveLocation, backupDir string) (string, error) {
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return "", fmt.Errorf("backup dir create failed: %w", err)
	}

	regPath := filepath.Join(backupDir, loc.FileName+".reg")
	target, err := regutil.ResolveTarget(loc.Path)
	if err != nil {
		return "", err
	}

	switch target.Kind {
	case regutil.TargetKey:
		targetPath := fmt.Sprintf("%s\\%s", target.RootName, target.KeyPath)
		cmd := exec.Command("reg", "export", targetPath, regPath, "/y")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("reg export failed: %w (%s)", err, string(out))
		}
		return regPath, nil
	case regutil.TargetValue:
		value, err := regutil.ReadValue(loc.Path)
		if err != nil {
			return "", err
		}
		content, err := buildRegValueFileContent(value)
		if err != nil {
			return "", err
		}
		if err := writeUTF16LEWithBOM(regPath, content); err != nil {
			return "", err
		}
		return regPath, nil
	default:
		return "", fmt.Errorf("지원하지 않는 레지스트리 백업 대상입니다")
	}
}

func buildRegValueFileContent(v regutil.ValueData) (string, error) {
	valueLine, err := formatRegValueLine(v.ValueName, v.Type, v.Data)
	if err != nil {
		return "", err
	}

	keyLine := fmt.Sprintf("[%s\\%s]", v.RootName, v.KeyPath)
	lines := []string{
		"Windows Registry Editor Version 5.00",
		"",
		keyLine,
		valueLine,
		"",
	}
	return strings.Join(lines, "\r\n"), nil
}

func formatRegValueLine(name string, typ uint32, data []byte) (string, error) {
	namePart := fmt.Sprintf("\"%s\"", escapeRegString(name))

	switch typ {
	case registry.SZ:
		return fmt.Sprintf("%s=\"%s\"", namePart, escapeRegString(decodeUTF16String(data))), nil
	case registry.DWORD:
		if len(data) < 4 {
			return "", fmt.Errorf("dword value length is invalid")
		}
		v := binary.LittleEndian.Uint32(data[:4])
		return fmt.Sprintf("%s=dword:%08x", namePart, v), nil
	case registry.QWORD:
		if len(data) < 8 {
			return "", fmt.Errorf("qword value length is invalid")
		}
		return fmt.Sprintf("%s=hex(b):%s", namePart, formatHexBytes(data[:8])), nil
	case registry.BINARY:
		return fmt.Sprintf("%s=hex:%s", namePart, formatHexBytes(data)), nil
	case registry.EXPAND_SZ:
		return fmt.Sprintf("%s=hex(2):%s", namePart, formatHexBytes(data)), nil
	case registry.MULTI_SZ:
		return fmt.Sprintf("%s=hex(7):%s", namePart, formatHexBytes(data)), nil
	default:
		return fmt.Sprintf("%s=hex(%x):%s", namePart, typ, formatHexBytes(data)), nil
	}
}

func formatHexBytes(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	parts := make([]string, len(data))
	for i, b := range data {
		parts[i] = fmt.Sprintf("%02x", b)
	}
	return strings.Join(parts, ",")
}

func escapeRegString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}

func decodeUTF16String(data []byte) string {
	if len(data) < 2 {
		return ""
	}
	u := make([]uint16, 0, len(data)/2)
	for i := 0; i+1 < len(data); i += 2 {
		u = append(u, binary.LittleEndian.Uint16(data[i:i+2]))
	}

	for len(u) > 0 && u[len(u)-1] == 0 {
		u = u[:len(u)-1]
	}
	if len(u) == 0 {
		return ""
	}
	return string(utf16.Decode(u))
}

func writeUTF16LEWithBOM(path, content string) error {
	runes := utf16.Encode([]rune(content))
	buf := make([]byte, 2+len(runes)*2)
	buf[0] = 0xFF
	buf[1] = 0xFE
	for i, r := range runes {
		binary.LittleEndian.PutUint16(buf[2+i*2:2+i*2+2], r)
	}
	if err := os.WriteFile(path, buf, 0o644); err != nil {
		return fmt.Errorf("reg file write failed: %w", err)
	}
	return nil
}
