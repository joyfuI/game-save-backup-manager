package model

// SaveLocation maps to savelocation.db rows.
type SaveLocation struct {
	RowID    int64
	Name     string
	Type     string
	FileName string
	Path     string
}

// ScanResult is computed runtime data for UI.
type ScanResult struct {
	SaveLocation
	Installed bool
	Matches   int
}
