package storage

import (
	"database/sql"
	"fmt"
	"path/filepath"

	"joyfuI/game-save-backup-manager/internal/model"
	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS savelocation (
	name TEXT NOT NULL,
	type TEXT NOT NULL,
	filename TEXT NOT NULL,
	path TEXT NOT NULL
);
`

func OpenAndInit(dbPath string) (*sql.DB, error) {
	absPath, err := filepath.Abs(dbPath)
	if err != nil {
		return nil, fmt.Errorf("resolve db path: %w", err)
	}

	db, err := sql.Open("sqlite", absPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("create schema: %w", err)
	}

	return db, nil
}

func LoadSaveLocations(db *sql.DB) ([]model.SaveLocation, error) {
	rows, err := db.Query(`SELECT rowid, name, type, filename, path FROM savelocation ORDER BY name COLLATE NOCASE`)
	if err != nil {
		return nil, fmt.Errorf("query savelocation: %w", err)
	}
	defer rows.Close()

	locations := make([]model.SaveLocation, 0)
	for rows.Next() {
		var loc model.SaveLocation
		if err := rows.Scan(&loc.RowID, &loc.Name, &loc.Type, &loc.FileName, &loc.Path); err != nil {
			return nil, fmt.Errorf("scan savelocation: %w", err)
		}
		locations = append(locations, loc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate savelocation: %w", err)
	}

	return locations, nil
}

func InsertSaveLocation(db *sql.DB, loc model.SaveLocation) error {
	_, err := db.Exec(
		`INSERT INTO savelocation (name, type, filename, path) VALUES (?, ?, ?, ?)`,
		loc.Name,
		loc.Type,
		loc.FileName,
		loc.Path,
	)
	if err != nil {
		return fmt.Errorf("insert savelocation: %w", err)
	}
	return nil
}

func UpdateSaveLocation(db *sql.DB, loc model.SaveLocation) error {
	result, err := db.Exec(
		`UPDATE savelocation SET name = ?, type = ?, filename = ?, path = ? WHERE rowid = ?`,
		loc.Name,
		loc.Type,
		loc.FileName,
		loc.Path,
		loc.RowID,
	)
	if err != nil {
		return fmt.Errorf("update savelocation: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected on update: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("update savelocation: no row found for rowid=%d", loc.RowID)
	}

	return nil
}

func DeleteSaveLocation(db *sql.DB, rowID int64) error {
	result, err := db.Exec(`DELETE FROM savelocation WHERE rowid = ?`, rowID)
	if err != nil {
		return fmt.Errorf("delete savelocation: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected on delete: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("delete savelocation: no row found for rowid=%d", rowID)
	}

	return nil
}
