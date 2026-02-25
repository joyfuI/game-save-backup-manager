//go:build !windows

package regutil

import "fmt"

type TargetKind int

const (
	TargetUnknown TargetKind = iota
	TargetKey
	TargetValue
)

type Target struct {
	Kind      TargetKind
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
	return false, fmt.Errorf("registry checks are only supported on windows")
}

func KeyExists(path string) (bool, error) {
	return false, fmt.Errorf("registry checks are only supported on windows")
}

func ValidatePath(path string) error {
	return fmt.Errorf("registry validation is only supported on windows")
}

func ResolveTarget(path string) (Target, error) {
	return Target{}, fmt.Errorf("registry target resolve is only supported on windows")
}

func ReadValue(path string) (ValueData, error) {
	return ValueData{}, fmt.Errorf("registry value read is only supported on windows")
}

func OpenInEditor(path string) error {
	return fmt.Errorf("registry editor open is only supported on windows")
}
