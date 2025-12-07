// Package clipboard handles clipboard read/write operations and paste simulation.
package clipboard

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/atotto/clipboard"

	"github/phaneendra24/goclipboard-manager/storage"
)

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

// Paste writes text to clipboard and simulates Ctrl+V using xdotool.
func Paste(text string) error {
	if err := clipboard.WriteAll(text); err != nil {
		return err
	}
	// simulate paste using xdotool (X11). Small sleep helps timing.
	time.Sleep(30 * time.Millisecond)
	cmd := exec.Command("xdotool", "key", "--clearmodifiers", "ctrl+v")
	return cmd.Run()
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
	// use xdotool to simulate Ctrl+V (X11)
	cmd := exec.Command("xdotool", "key", "--clearmodifiers", "ctrl+v")
	time.Sleep(30 * time.Millisecond)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("xdotool paste failed: %w", err)
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
