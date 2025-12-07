// Package storage handles persistence of clipboard history to disk.
package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	// HistoryFileName is the name of the history JSON file.
	HistoryFileName = "clip_history.json"
	// LogFileName is the name of the log file.
	LogFileName = "clipcli.log"
	// MaxHistory is the maximum number of history entries to keep.
	MaxHistory = 500
)

// DataDir returns the data directory for clipcli, creating it if necessary.
func DataDir() (string, error) {
	xdg := os.Getenv("XDG_DATA_HOME")
	if xdg == "" {
		home := os.Getenv("HOME")
		if home == "" {
			return "", fmt.Errorf("no HOME or XDG_DATA_HOME set")
		}
		xdg = filepath.Join(home, ".local", "share")
	}
	dir := filepath.Join(xdg, "clipcli")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

// HistoryFilePath returns the path to the history file.
func HistoryFilePath() (string, error) {
	dir, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, HistoryFileName), nil
}

// LogFilePath returns the path to the log file.
func LogFilePath() (string, error) {
	dir, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, LogFileName), nil
}

// LoadHistory loads the clipboard history from disk.
func LoadHistory() ([]string, error) {
	p, err := HistoryFilePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	var hist []string
	if err := json.Unmarshal(data, &hist); err != nil {
		return nil, err
	}
	return hist, nil
}

// SaveHistory writes the history to disk atomically.
func SaveHistory(hist []string) error {
	p, err := HistoryFilePath()
	if err != nil {
		return err
	}
	if len(hist) > MaxHistory {
		hist = hist[:MaxHistory]
	}
	dir := filepath.Dir(p)
	tmp := filepath.Join(dir, fmt.Sprintf(".%s.tmp", HistoryFileName))
	data, err := json.MarshalIndent(hist, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	// rename is atomic on most Unix filesystems
	return os.Rename(tmp, p)
}

// Preview returns a truncated preview of a string.
func Preview(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
