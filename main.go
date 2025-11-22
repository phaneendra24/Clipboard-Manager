// Command clipcli is a simple clipboard manager for the command line.
// It allows saving clipboard history, listing it, and pasting from it.
// main.go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/atotto/clipboard"
)

const (
	historyFileName = "clip_history.json"
	maxHistory      = 200
)

// historyFilePath determines the path for the history file.
// It respects the XDG_DATA_HOME environment variable and falls back to
// ~/.local/share/clipcli/clip_history.json if it's not set.
// It creates the directory if it does not exist.
func historyFilePath() (string, error) {
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
	return filepath.Join(dir, historyFileName), nil
}

// loadHistory reads the clipboard history from the JSON file.
// If the history file does not exist, it returns an empty slice.
func loadHistory() ([]string, error) {
	p, err := historyFilePath()
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

// saveHistory writes the clipboard history to the JSON file.
// It truncates the history to maxHistory items if it's longer.
func saveHistory(hist []string) error {
	p, err := historyFilePath()
	if err != nil {
		return err
	}
	if len(hist) > maxHistory {
		hist = hist[:maxHistory]
	}
	data, err := json.MarshalIndent(hist, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}

// cmdSave reads the current clipboard content and saves it to the history.
// It avoids saving empty content or consecutive duplicate entries.
func cmdSave() error {
	txt, err := clipboard.ReadAll()
	if err != nil {
		return fmt.Errorf("read clipboard: %w", err)
	}
	if strings.TrimSpace(txt) == "" {
		return fmt.Errorf("clipboard empty or whitespace")
	}
	hist, err := loadHistory()
	if err != nil {
		return err
	}
	// avoid duplicate consecutive entries
	if len(hist) == 0 || hist[0] != txt {
		hist = append([]string{txt}, hist...)
	}
	if err := saveHistory(hist); err != nil {
		return err
	}
	fmt.Println("saved to history")
	return nil
}

// cmdList displays the clipboard history with a preview of each entry.
// Previews are truncated to 120 characters.
func cmdList() error {
	hist, err := loadHistory()
	if err != nil {
		return err
	}
	if len(hist) == 0 {
		fmt.Println("(history empty)")
		return nil
	}
	for i, entry := range hist {
		preview := strings.Split(entry, "\n")[0]
		if len(preview) > 120 {
			preview = preview[:120] + "…"
		}
		fmt.Printf("[%d] %s\n", i, preview)
	}
	return nil
}

// cmdPaste copies the history item at the given index to the clipboard
// and then simulates a paste (Ctrl+V) command using xdotool.
func cmdPaste(idx int) error {
	hist, err := loadHistory()
	if err != nil {
		return err
	}
	if len(hist) == 0 {
		return fmt.Errorf("history empty")
	}
	if idx < 0 || idx >= len(hist) {
		return fmt.Errorf("index out of range")
	}
	text := hist[idx]
	// write to system clipboard
	if err := clipboard.WriteAll(text); err != nil {
		return fmt.Errorf("write clipboard: %w", err)
	}
	// use xdotool to simulate Ctrl+V (X11)
	cmd := exec.Command("xdotool", "key", "--clearmodifiers", "ctrl+v")
	// small sleep before pasting can help if invoked immediately after copying
	time.Sleep(30 * time.Millisecond)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("xdotool paste failed: %w", err)
	}
	fmt.Printf("pasted index %d\n", idx)
	return nil
}

// cmdClear removes all entries from the clipboard history.
func cmdClear() error {
	if err := saveHistory([]string{}); err != nil {
		return err
	}
	fmt.Println("cleared history")
	return nil
}

// printUsage prints the command-line usage instructions.
func printUsage() {
	fmt.Println("Usage: clipcli <command>\nCommands:\n  save        Save current clipboard to history\n  list        List history previews\n  paste N     Paste history item N (0 = most recent)\n  clear       Clear history")
}

// main is the entry point for the application.
// It parses command-line arguments and executes the corresponding command.
func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "save":
		if err := cmdSave(); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(2)
		}
	case "list":
		if err := cmdList(); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(2)
		}
	case "paste":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "paste requires index")
			os.Exit(2)
		}
		i, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, "invalid index")
			os.Exit(2)
		}
		if err := cmdPaste(i); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(2)
		}
	case "clear":
		if err := cmdClear(); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(2)
		}
	default:
		printUsage()
		os.Exit(1)
	}
}
