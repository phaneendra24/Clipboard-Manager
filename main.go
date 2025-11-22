// main.go
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/atotto/clipboard"
)

const (
	historyFileName = "clip_history.json"
	logFileName     = "clipcli.log"
	maxHistory      = 500
	defaultPollMS   = 300 // poll interval in ms
)

// historyFilePath returns path under XDG_DATA_HOME or fallback to ~/.local/share/clipcli/
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

func logFilePath() (string, error) {
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
	return filepath.Join(dir, logFileName), nil
}

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

// atomicSaveHistory writes to temp file then renames for atomicity
func atomicSaveHistory(hist []string) error {
	p, err := historyFilePath()
	if err != nil {
		return err
	}
	if len(hist) > maxHistory {
		hist = hist[:maxHistory]
	}
	dir := filepath.Dir(p)
	tmp := filepath.Join(dir, fmt.Sprintf(".%s.tmp", historyFileName))
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

func cmdSave() error {
	txt, err := clipboard.ReadAll()
	if err != nil {
		return fmt.Errorf("read clipboard: %w", err)
	}
	if strings.TrimSpace(txt) == "" {
		return errors.New("clipboard empty or whitespace")
	}
	hist, err := loadHistory()
	if err != nil {
		return err
	}
	// avoid duplicate consecutive entries
	if len(hist) == 0 || hist[0] != txt {
		hist = append([]string{txt}, hist...)
	}
	if err := atomicSaveHistory(hist); err != nil {
		return err
	}
	fmt.Println("saved to history")
	return nil
}

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
		if len(preview) > 200 {
			preview = preview[:200] + "…"
		}
		fmt.Printf("[%d] %s\n", i, preview)
	}
	return nil
}

func cmdPaste(idx int) error {
	hist, err := loadHistory()
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
	// tiny sleep helps if paste is triggered right after copying
	time.Sleep(30 * time.Millisecond)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("xdotool paste failed: %w", err)
	}
	fmt.Printf("pasted index %d\n", idx)
	return nil
}

func cmdClear() error {
	if err := atomicSaveHistory([]string{}); err != nil {
		return err
	}
	fmt.Println("cleared history")
	return nil
}

// runDaemon polls clipboard and saves new entries
func runDaemon(pollMS int, logger *log.Logger, stopCh <-chan struct{}) error {
	logger.Printf("daemon starting (poll %dms)\n", pollMS)
	ticker := time.NewTicker(time.Duration(pollMS) * time.Millisecond)
	defer ticker.Stop()

	var lastSeen string
	for {
		select {
		case <-stopCh:
			logger.Println("daemon stopping (received stop)")
			return nil
		case <-ticker.C:
			txt, err := clipboard.ReadAll()
			if err != nil {
				// keep running, just log error
				logger.Printf("clipboard read error: %v\n", err)
				continue
			}
			// sanitize: ignore empty strings
			if strings.TrimSpace(txt) == "" {
				continue
			}
			if txt == lastSeen {
				continue // no change
			}
			// update lastSeen only after successful save
			hist, err := loadHistory()
			if err != nil {
				logger.Printf("load history error: %v\n", err)
				continue
			}
			// skip if already the most recent (protect against races)
			if len(hist) > 0 && hist[0] == txt {
				lastSeen = txt
				continue
			}
			// push front
			newHist := append([]string{txt}, hist...)
			if len(newHist) > maxHistory {
				newHist = newHist[:maxHistory]
			}
			if err := atomicSaveHistory(newHist); err != nil {
				logger.Printf("save history error: %v\n", err)
				continue
			}
			lastSeen = txt
			logger.Printf("captured clipboard (len=%d) preview: %q\n", len(newHist), preview(txt, 80))
		}
	}
}

func preview(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func printUsage() {
	fmt.Println("Usage: clipcli <command>\nCommands:\n  serve [poll_ms]   Run daemon (poll_ms optional, default 300)\n  save              Save current clipboard to history\n  list              List history previews\n  paste N           Paste history item N (0 = most recent)\n  clear             Clear history")
}

func main() {
	// configure logging to file (if possible) else stdout
	logPath, _ := logFilePath()
	logf, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		// fallback to stdout
		log.SetOutput(os.Stdout)
	} else {
		log.SetOutput(logf)
		defer logf.Close()
	}
	logger := log.Default()

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "serve":
		pollMS := defaultPollMS
		if len(os.Args) >= 3 {
			if v, err := strconv.Atoi(os.Args[2]); err == nil && v > 0 {
				pollMS = v
			}
		}
		// handle signals for graceful shutdown
		stop := make(chan struct{})
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			sig := <-sigs
			logger.Printf("received signal %v, shutting down\n", sig)
			close(stop)
		}()
		if err := runDaemon(pollMS, logger, stop); err != nil {
			logger.Fatalf("daemon error: %v\n", err)
		}
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
