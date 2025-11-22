// main.go
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	historyFileName = "clip_history.json"
	maxHistory      = 500
	defaultPollMS   = 300 // for serve if you use daemon
)

// --- file paths / persistence (same as earlier steps) --- //
func dataDir() (string, error) {
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

func historyFilePath() (string, error) {
	dir, err := dataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, historyFileName), nil
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

func saveHistory(hist []string) error {
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
	return os.Rename(tmp, p)
}

// --- clipboard + paste helper --- //
func doPaste(text string) error {
	// write to clipboard
	if err := clipboard.WriteAll(text); err != nil {
		return err
	}
	// simulate paste using xdotool (X11). Small sleep helps timing.
	time.Sleep(30 * time.Millisecond)
	cmd := exec.Command("xdotool", "key", "--clearmodifiers", "ctrl+v")
	return cmd.Run()
}

// --- CLI commands (save/list/paste/clear) --- //
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
	if len(hist) == 0 || hist[0] != txt {
		hist = append([]string{txt}, hist...)
	}
	return saveHistory(hist)
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
	return doPaste(hist[idx])
}

func cmdClear() error {
	return saveHistory([]string{})
}

// --- TUI: browsing and actions --- //
func runTUI() error {
	hist, err := loadHistory()
	if err != nil {
		return err
	}

	app := tview.NewApplication()

	// State
	orig := hist
	filtered := make([]int, len(orig))
	for i := range orig {
		filtered[i] = i
	}

	// UI components
	list := tview.NewList().ShowSecondaryText(false)
	info := tview.NewTextView().SetDynamicColors(true)
	searchInput := tview.NewInputField().SetLabel("/ ").SetFieldWidth(0)

	// --- Layout must be defined BEFORE any handler uses it ---
	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(searchInput, 1, 0, false).
		AddItem(list, 0, 1, true).
		AddItem(info, 2, 0, false)

	// Helper to refresh list
	refreshList := func() {
		list.Clear()
		for _, idx := range filtered {
			item := orig[idx]
			preview := strings.Split(item, "\n")[0]
			if len(preview) > 200 {
				preview = preview[:200] + "…"
			}
			list.AddItem(preview, "", 0, nil)
		}
		info.Clear()
		fmt.Fprintf(info, "[yellow](Enter) Paste  [green]c(copy)  [red]d(delete)  [blue]/(search)  [white]q(quit)\n")
		fmt.Fprintf(info, "Showing %d of %d items\n", len(filtered), len(orig))
	}

	refreshList()

	// --- Actions ---
	pasteSelected := func() {
		if list.GetItemCount() == 0 {
			return
		}
		idx := list.GetCurrentItem()
		origIdx := filtered[idx]
		text := orig[origIdx]

		if err := doPaste(text); err != nil {
			modal := tview.NewModal().
				SetText("Paste failed: " + err.Error()).
				AddButtons([]string{"OK"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					app.SetRoot(layout, true).SetFocus(list)
				})
			app.SetRoot(modal, false)
			return
		}

		info.Clear()
		fmt.Fprintf(info, "[green]Pasted item %d\n", origIdx)
	}

	copySelected := func() {
		if list.GetItemCount() == 0 {
			return
		}
		idx := list.GetCurrentItem()
		origIdx := filtered[idx]

		if err := clipboard.WriteAll(orig[origIdx]); err != nil {
			info.Clear()
			fmt.Fprintf(info, "[red]Copy failed: %v\n", err)
			return
		}
		info.Clear()
		fmt.Fprintf(info, "[green]Copied item %d to clipboard\n", origIdx)
	}

	deleteSelected := func() {
		if list.GetItemCount() == 0 {
			return
		}
		idx := list.GetCurrentItem()
		origIdx := filtered[idx]

		orig = append(orig[:origIdx], orig[origIdx+1:]...)

		// rebuild filtered
		filtered = make([]int, len(orig))
		for i := range orig {
			filtered[i] = i
		}

		if err := saveHistory(orig); err != nil {
			info.Clear()
			fmt.Fprintf(info, "[red]Failed to save: %v\n", err)
		}

		refreshList()
		app.SetFocus(list)
	}

	applyFilter := func(q string) {
		q = strings.TrimSpace(q)
		if q == "" {
			filtered = make([]int, len(orig))
			for i := range orig {
				filtered[i] = i
			}
		} else {
			q = strings.ToLower(q)
			tmp := []int{}
			for i, v := range orig {
				if strings.Contains(strings.ToLower(v), q) {
					tmp = append(tmp, i)
				}
			}
			filtered = tmp
		}
		refreshList()
		if len(filtered) > 0 {
			list.SetCurrentItem(0)
		}
	}

	// Key handling
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			if app.GetFocus() == searchInput {
				searchInput.SetText("")
				applyFilter("")
				app.SetFocus(list)
				return nil
			}
			app.Stop()
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case '/':
				searchInput.SetText("")
				app.SetFocus(searchInput)
				return nil
			case 'q', 'Q':
				app.Stop()
				return nil
			case 'c', 'C':
				copySelected()
				return nil
			case 'd', 'D':
				deleteSelected()
				return nil
			}
		}
		return event
	})

	list.SetSelectedFunc(func(i int, main, sec string, r rune) {
		pasteSelected()
	})

	searchInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			applyFilter(searchInput.GetText())
			app.SetFocus(list)
		} else if key == tcell.KeyEsc {
			searchInput.SetText("")
			applyFilter("")
			app.SetFocus(list)
		}
	})

	return app.SetRoot(layout, true).EnableMouse(false).Run()
}

// --- simple usage/help --- //
func printUsage() {
	fmt.Println("Usage: clipcli <command>\nCommands:\n  serve [poll_ms]   Run daemon (not required for TUI)\n  save              Save current clipboard to history\n  list              List history previews\n  paste N           Paste history item N (0 = most recent)\n  clear             Clear history\n  tui               Open terminal UI to browse history")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "serve":
		// shortcut: reuse earlier serve if you want. Not implemented in this snippet to keep focus on TUI.
		fmt.Println("serve not implemented in this build; run previous serve binary if you have one.")
		os.Exit(0)
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
	case "tui":
		if err := runTUI(); err != nil {
			fmt.Fprintln(os.Stderr, "tui error:", err)
			os.Exit(2)
		}
	default:
		printUsage()
		os.Exit(1)
	}
}
