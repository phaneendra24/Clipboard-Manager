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
)

// MaxHistory is the maximum number of history entries to keep (configurable).
var MaxHistory = 500

// ClipboardData represents the complete clipboard storage with history and pinned items.
type ClipboardData struct {
	History []string          `json:"history"`
	Pinned  map[string]bool   `json:"pinned"` // Map of content hash to pinned status
}

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

// LoadClipboardData loads the complete clipboard data (history + pinned) from disk.
func LoadClipboardData() (*ClipboardData, error) {
	p, err := HistoryFilePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return &ClipboardData{History: []string{}, Pinned: make(map[string]bool)}, nil
		}
		return nil, err
	}

	// Try to parse as new format first
	var clipData ClipboardData
	if err := json.Unmarshal(data, &clipData); err == nil && clipData.History != nil {
		if clipData.Pinned == nil {
			clipData.Pinned = make(map[string]bool)
		}
		return &clipData, nil
	}

	// Fallback: try to parse as old format (just array of strings)
	var hist []string
	if err := json.Unmarshal(data, &hist); err != nil {
		return nil, err
	}
	return &ClipboardData{History: hist, Pinned: make(map[string]bool)}, nil
}

// SaveClipboardData writes the complete clipboard data to disk atomically.
func SaveClipboardData(clipData *ClipboardData) error {
	p, err := HistoryFilePath()
	if err != nil {
		return err
	}
	// Preserve pinned items when trimming - only trim unpinned items
	if len(clipData.History) > MaxHistory {
		var pinned, unpinned []string
		for _, item := range clipData.History {
			if clipData.Pinned[item] {
				pinned = append(pinned, item)
			} else {
				unpinned = append(unpinned, item)
			}
		}
		// Keep all pinned + as many unpinned as will fit
		maxUnpinned := MaxHistory - len(pinned)
		if maxUnpinned < 0 {
			maxUnpinned = 0
		}
		if len(unpinned) > maxUnpinned {
			unpinned = unpinned[:maxUnpinned]
		}
		clipData.History = append(pinned, unpinned...)
	}
	dir := filepath.Dir(p)
	tmp := filepath.Join(dir, fmt.Sprintf(".%s.tmp", HistoryFileName))
	data, err := json.MarshalIndent(clipData, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

// LoadHistory loads only the history (backward compatible).
func LoadHistory() ([]string, error) {
	clipData, err := LoadClipboardData()
	if err != nil {
		return nil, err
	}
	return clipData.History, nil
}

// SaveHistory saves only the history (backward compatible).
func SaveHistory(hist []string) error {
	clipData, err := LoadClipboardData()
	if err != nil {
		// If loading fails, create new data
		clipData = &ClipboardData{Pinned: make(map[string]bool)}
	}
	clipData.History = hist
	return SaveClipboardData(clipData)
}

// IsPinned checks if an item is pinned.
func IsPinned(text string) bool {
	clipData, err := LoadClipboardData()
	if err != nil {
		return false
	}
	return clipData.Pinned[text]
}

// TogglePin toggles the pinned status of an item.
func TogglePin(text string) (bool, error) {
	clipData, err := LoadClipboardData()
	if err != nil {
		return false, err
	}
	if clipData.Pinned == nil {
		clipData.Pinned = make(map[string]bool)
	}
	newStatus := !clipData.Pinned[text]
	if newStatus {
		clipData.Pinned[text] = true
	} else {
		delete(clipData.Pinned, text)
	}
	return newStatus, SaveClipboardData(clipData)
}

// GetPinnedItems returns only the pinned items from history.
func GetPinnedItems(clipData *ClipboardData) []string {
	var pinned []string
	for _, item := range clipData.History {
		if clipData.Pinned[item] {
			pinned = append(pinned, item)
		}
	}
	return pinned
}

// GetUnpinnedItems returns only the unpinned items from history.
func GetUnpinnedItems(clipData *ClipboardData) []string {
	var unpinned []string
	for _, item := range clipData.History {
		if !clipData.Pinned[item] {
			unpinned = append(unpinned, item)
		}
	}
	return unpinned
}

// Preview returns a truncated preview of a string.
func Preview(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
