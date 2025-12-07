// Package ui provides the graphical user interface for the clipboard manager.
package ui

import (
	"fmt"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	clipboardPkg "github/phaneendra24/goclipboard-manager/clipboard"
	"github/phaneendra24/goclipboard-manager/storage"
)

// searchEntryWidget extends Entry to forward navigation shortcuts
type searchEntryWidget struct {
	widget.Entry
	onUp      func()
	onDown    func()
	onEscape  func()
	onDelete  func()
	onPin     func()
	onPaste   func()
}

func (e *searchEntryWidget) TypedKey(key *fyne.KeyEvent) {
	switch key.Name {
	case fyne.KeyUp:
		if e.onUp != nil {
			e.onUp()
		}
		return
	case fyne.KeyDown:
		if e.onDown != nil {
			e.onDown()
		}
		return
	case fyne.KeyEscape:
		if e.onEscape != nil {
			e.onEscape()
		}
		return
	case fyne.KeyDelete:
		if e.onDelete != nil {
			e.onDelete()
		}
		return
	}
	e.Entry.TypedKey(key)
}

func (e *searchEntryWidget) TypedShortcut(s fyne.Shortcut) {
	if cs, ok := s.(*desktop.CustomShortcut); ok {
		if cs.Modifier == fyne.KeyModifierControl {
			switch cs.KeyName {
			case fyne.KeyJ:
				if e.onDown != nil {
					e.onDown()
				}
				return
			case fyne.KeyK:
				if e.onUp != nil {
					e.onUp()
				}
				return
			case fyne.KeyP:
				if e.onPin != nil {
					e.onPin()
				}
				return
			case fyne.KeyD:
				if e.onDelete != nil {
					e.onDelete()
				}
				return
			case fyne.KeyReturn:
				if e.onPaste != nil {
					e.onPaste()
				}
				return
			}
		}
	}
	e.Entry.TypedShortcut(s)
}

// RunGUI starts the Fyne-based graphical clipboard manager.
func RunGUI() error {
	// Use app ID for better window manager recognition
	a := app.NewWithID("com.clipcli.manager")
	
	// Apply Catppuccin theme
	a.Settings().SetTheme(&CatppuccinTheme{})

	w := a.NewWindow("📋 Clipboard Manager")
	w.Resize(fyne.NewSize(700, 500))
	w.CenterOnScreen()

	// Load clipboard data (history + pinned)
	clipData, err := storage.LoadClipboardData()
	if err != nil {
		return err
	}

	// Build sorted list: pinned items first, then unpinned
	buildSortedHistory := func() []string {
		pinned := storage.GetPinnedItems(clipData)
		unpinned := storage.GetUnpinnedItems(clipData)
		return append(pinned, unpinned...)
	}

	// State
	sortedHist := buildSortedHistory()
	filtered := make([]int, len(sortedHist))
	for i := range sortedHist {
		filtered[i] = i
	}

	// Custom Entry that forwards navigation shortcuts
	var moveUp, moveDown, deleteSelected func()
	var closeWindow func()
	var togglePinSelected, pasteSelected func()

	searchEntry := &searchEntryWidget{
		Entry:    widget.Entry{},
		onUp:     func() { moveUp() },
		onDown:   func() { moveDown() },
		onEscape: func() { closeWindow() },
		onDelete: func() { deleteSelected() },
		onPin:    func() { togglePinSelected() },
		onPaste:  func() { pasteSelected() },
	}
	searchEntry.ExtendBaseWidget(searchEntry)
	searchEntry.SetPlaceHolder("  Search clipboard...")

	// Clean, minimal status bar
	statusLabel := widget.NewLabel(fmt.Sprintf("⏎ Copy  •  Ctrl+⏎ Paste  •  Ctrl+P Pin  •  Del Remove  │  %d items", len(sortedHist)))
	statusLabel.Importance = widget.LowImportance

	// List widget
	list := widget.NewList(
		func() int {
			return len(filtered)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			if i < len(filtered) {
				idx := filtered[i]
				if idx < len(sortedHist) {
					item := sortedHist[idx]
					preview := strings.Split(item, "\n")[0]
					if len(preview) > 90 {
						preview = preview[:90] + "…"
					}
					// Add pin indicator
					prefix := "  "
					if clipData.Pinned[item] {
						prefix = "📌"
					}
					o.(*widget.Label).SetText(fmt.Sprintf("%s %s", prefix, preview))
				}
			}
		},
	)

	var selectedIndex int = 0
	if len(filtered) > 0 {
		list.Select(0)
	}
	list.OnSelected = func(id widget.ListItemID) {
		selectedIndex = id
	}

	// Refresh helper
	refreshAll := func() {
		sortedHist = buildSortedHistory()
		list.Refresh()
		statusLabel.SetText(fmt.Sprintf("⏎ Copy  •  Ctrl+⏎ Paste  •  Ctrl+P Pin  •  Del Remove  │  %d items", len(sortedHist)))
	}

	// Fuzzy match function - returns score (higher = better match), -1 = no match
	fuzzyMatch := func(pattern, text string) int {
		pattern = strings.ToLower(pattern)
		text = strings.ToLower(text)
		
		if pattern == "" {
			return 0
		}
		
		// Check for exact substring match first (highest priority)
		if strings.Contains(text, pattern) {
			return 1000 + len(pattern)*10
		}
		
		// Fuzzy matching: characters must appear in order
		pIdx := 0
		score := 0
		lastMatchIdx := -1
		wordStart := true
		
		for i := 0; i < len(text) && pIdx < len(pattern); i++ {
			if text[i] == pattern[pIdx] {
				pIdx++
				// Bonus for consecutive matches
				if lastMatchIdx == i-1 {
					score += 15
				} else {
					score += 5
				}
				// Bonus for matching at word start
				if wordStart {
					score += 10
				}
				lastMatchIdx = i
			}
			// Track word boundaries
			wordStart = text[i] == ' ' || text[i] == '/' || text[i] == '_' || text[i] == '-'
		}
		
		// All pattern characters must be found
		if pIdx < len(pattern) {
			return -1
		}
		
		return score
	}

	// Filter function with fuzzy search
	applyFilter := func(query string) {
		query = strings.TrimSpace(query)
		if query == "" {
			filtered = make([]int, len(sortedHist))
			for i := range sortedHist {
				filtered[i] = i
			}
		} else {
			// Collect matches with scores
			type matchResult struct {
				index int
				score int
			}
			matches := []matchResult{}
			
			for i, v := range sortedHist {
				score := fuzzyMatch(query, v)
				if score >= 0 {
					matches = append(matches, matchResult{index: i, score: score})
				}
			}
			
			// Sort by score (higher first)
			sort.Slice(matches, func(a, b int) bool {
				return matches[a].score > matches[b].score
			})
			
			// Extract indices
			filtered = make([]int, len(matches))
			for i, m := range matches {
				filtered[i] = m.index
			}
		}
		selectedIndex = 0
		list.Refresh()
		if len(filtered) > 0 {
			list.Select(0)
		}
	}

	// Action functions
	var copySelected func()
	var copyAndClose func()
	var refreshHistory func()
	var clearAll func()

	searchEntry.OnChanged = applyFilter

	searchEntry.OnSubmitted = func(s string) {
		if len(filtered) > 0 {
			selectedIndex = 0
			list.Select(0)
			copyAndClose()
		}
	}

	copySelected = func() {
		if selectedIndex >= 0 && selectedIndex < len(filtered) {
			idx := filtered[selectedIndex]
			if idx < len(sortedHist) {
				if err := clipboardPkg.CopyToClipboard(sortedHist[idx]); err != nil {
					dialog.ShowError(err, w)
				} else {
					statusLabel.SetText("✓ Copied to clipboard")
				}
			}
		}
	}

	copyAndClose = func() {
		copySelected()
		w.Close()
	}

	pasteSelected = func() {
		if selectedIndex >= 0 && selectedIndex < len(filtered) {
			idx := filtered[selectedIndex]
			if idx < len(sortedHist) {
				if err := clipboardPkg.Paste(sortedHist[idx]); err != nil {
					dialog.ShowError(err, w)
				} else {
					w.Close()
				}
			}
		}
	}

	togglePinSelected = func() {
		if selectedIndex >= 0 && selectedIndex < len(filtered) {
			idx := filtered[selectedIndex]
			if idx < len(sortedHist) {
				item := sortedHist[idx]
				isPinned, err := storage.TogglePin(item)
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				// Reload data
				clipData, _ = storage.LoadClipboardData()
				if isPinned {
					statusLabel.SetText("📌 Pinned")
				} else {
					statusLabel.SetText("📌 Unpinned")
				}
				refreshAll()
				applyFilter(searchEntry.Text)
			}
		}
	}

	deleteSelected = func() {
		if selectedIndex >= 0 && selectedIndex < len(filtered) {
			idx := filtered[selectedIndex]
			if idx < len(sortedHist) {
				item := sortedHist[idx]
				// Remove from history
				newHist := []string{}
				for _, h := range clipData.History {
					if h != item {
						newHist = append(newHist, h)
					}
				}
				clipData.History = newHist
				// Remove from pinned if present
				delete(clipData.Pinned, item)
				
				if err := storage.SaveClipboardData(clipData); err != nil {
					dialog.ShowError(err, w)
					return
				}
				statusLabel.SetText("✓ Deleted")
				refreshAll()
				applyFilter(searchEntry.Text)
			}
		}
	}

	refreshHistory = func() {
		newData, err := storage.LoadClipboardData()
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		clipData = newData
		searchEntry.SetText("")
		selectedIndex = 0
		refreshAll()
		if len(filtered) > 0 {
			list.Select(0)
		}
		statusLabel.SetText(fmt.Sprintf("✓ Refreshed: %d items", len(sortedHist)))
	}

	clearAll = func() {
		dialog.ShowConfirm("Clear All", "Delete all clipboard history (including pinned)?", func(confirm bool) {
			if confirm {
				clipData.History = []string{}
				clipData.Pinned = make(map[string]bool)
				if err := storage.SaveClipboardData(clipData); err != nil {
					dialog.ShowError(err, w)
					return
				}
				refreshAll()
				statusLabel.SetText("✓ Cleared")
			}
		}, w)
	}

	// Navigation helper functions - assign to variables for searchEntry callbacks
	moveUp = func() {
		if selectedIndex > 0 {
			selectedIndex--
			list.Select(selectedIndex)
		}
	}
	moveDown = func() {
		if selectedIndex < len(filtered)-1 {
			selectedIndex++
			list.Select(selectedIndex)
		}
	}
	closeWindow = func() {
		w.Close()
	}

	// Keyboard shortcuts (for non-modified keys)
	w.Canvas().SetOnTypedKey(func(ev *fyne.KeyEvent) {
		switch ev.Name {
		case fyne.KeyUp:
			moveUp()
		case fyne.KeyDown:
			moveDown()
		case fyne.KeyDelete:
			deleteSelected()
		}
	})

	// Desktop shortcuts - these work regardless of focus
	shortcutPaste := &desktop.CustomShortcut{KeyName: fyne.KeyReturn, Modifier: fyne.KeyModifierControl}
	shortcutPin := &desktop.CustomShortcut{KeyName: fyne.KeyP, Modifier: fyne.KeyModifierControl}
	shortcutRefresh := &desktop.CustomShortcut{KeyName: fyne.KeyR, Modifier: fyne.KeyModifierControl}
	shortcutDelete := &desktop.CustomShortcut{KeyName: fyne.KeyD, Modifier: fyne.KeyModifierControl}
	shortcutBackspace := &desktop.CustomShortcut{KeyName: fyne.KeyBackspace, Modifier: fyne.KeyModifierControl}
	shortcutClear := &desktop.CustomShortcut{KeyName: fyne.KeyL, Modifier: fyne.KeyModifierControl}
	shortcutEscape := &desktop.CustomShortcut{KeyName: fyne.KeyEscape, Modifier: fyne.KeyModifierShift} // Shift+Esc always quits
	// Vim-style navigation: Ctrl+J (down), Ctrl+K (up)
	shortcutNavDown := &desktop.CustomShortcut{KeyName: fyne.KeyJ, Modifier: fyne.KeyModifierControl}
	shortcutNavUp := &desktop.CustomShortcut{KeyName: fyne.KeyK, Modifier: fyne.KeyModifierControl}

	w.Canvas().AddShortcut(shortcutPaste, func(s fyne.Shortcut) { pasteSelected() })
	w.Canvas().AddShortcut(shortcutPin, func(s fyne.Shortcut) { togglePinSelected() })
	w.Canvas().AddShortcut(shortcutRefresh, func(s fyne.Shortcut) { refreshHistory() })
	w.Canvas().AddShortcut(shortcutDelete, func(s fyne.Shortcut) { deleteSelected() })
	w.Canvas().AddShortcut(shortcutBackspace, func(s fyne.Shortcut) { deleteSelected() })
	w.Canvas().AddShortcut(shortcutClear, func(s fyne.Shortcut) { clearAll() })
	w.Canvas().AddShortcut(shortcutEscape, func(s fyne.Shortcut) { w.Close() })
	w.Canvas().AddShortcut(shortcutNavDown, func(s fyne.Shortcut) { moveDown() })
	w.Canvas().AddShortcut(shortcutNavUp, func(s fyne.Shortcut) { moveUp() })

	// Layout
	content := container.NewBorder(
		searchEntry,  // top
		statusLabel,  // bottom
		nil, nil,
		list,         // center
	)

	w.SetContent(content)
	w.Show()

	// Auto-focus search
	go func() {
		w.Canvas().Focus(searchEntry)
	}()

	a.Run()
	return nil
}
