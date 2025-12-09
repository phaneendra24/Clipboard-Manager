// Package clipboard handles clipboard read/write operations and paste simulation.
package clipboard

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/atotto/clipboard"

	"github/phaneendra24/goclipboard-manager/storage"
)

// isWayland checks if running under Wayland
func isWayland() bool {
	return os.Getenv("WAYLAND_DISPLAY") != ""
}

// SimulatePaste simulates Ctrl+V using the appropriate tool for the display server
func SimulatePaste() error {
	time.Sleep(30 * time.Millisecond)
	if isWayland() {
		// Use wtype for Wayland
		cmd := exec.Command("wtype", "-M", "ctrl", "v", "-m", "ctrl")
		return cmd.Run()
	}
	// Use xdotool for X11
	cmd := exec.Command("xdotool", "key", "--clearmodifiers", "ctrl+v")
	return cmd.Run()
}

// Save reads the current clipboard and saves it to history.
func Save() error {
	txt, err := clipboard.ReadAll()
	if err != nil {
		return fmt.Errorf("read clipboard: %w", err)
	}
	if strings.TrimSpace(txt) == "" {
		return errors.New("clipboard empty or whitespace")
	}
	hist, err := storage.LoadHistory()
	if err != nil {
		return err
	}
	// avoid duplicate consecutive entries
	if len(hist) == 0 || hist[0] != txt {
		hist = append([]string{txt}, hist...)
	}
	return storage.SaveHistory(hist)
}

// Paste writes text to clipboard and simulates Ctrl+V.
func Paste(text string) error {
	if err := clipboard.WriteAll(text); err != nil {
		return err
	}
	return SimulatePaste()
}

// PasteByIndex pastes the history item at the given index.
func PasteByIndex(idx int) error {
	hist, err := storage.LoadHistory()
	if err != nil {
		return err
	}
	if len(hist) == 0 {
		return errors.New("history empty")
	}
	if idx < 0 || idx >= len(hist) {
		return fmt.Errorf("index out of range (0..%d)", len(hist)-1)
	}
	text := hist[idx]
	// write to system clipboard
	if err := clipboard.WriteAll(text); err != nil {
		return fmt.Errorf("write clipboard: %w", err)
	}
	if err := SimulatePaste(); err != nil {
		return fmt.Errorf("paste simulation failed: %w", err)
	}
	return nil
}

// CopyToClipboard writes text to the system clipboard.
func CopyToClipboard(text string) error {
	return clipboard.WriteAll(text)
}

// ReadClipboard reads the current system clipboard content.
func ReadClipboard() (string, error) {
	return clipboard.ReadAll()
}

