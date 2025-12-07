// Package ui provides the graphical user interface for the clipboard manager.
package ui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	clipboardPkg "github/phaneendra24/goclipboard-manager/clipboard"
	"github/phaneendra24/goclipboard-manager/storage"
)

// RunGUI starts the Fyne-based graphical clipboard manager.
func RunGUI() error {
	fmt.Println("Starting Clipboard Manager GUI...")
	
	// Use app ID for better window manager recognition
	a := app.NewWithID("com.clipcli.manager")
	a.Settings().SetTheme(theme.DarkTheme())

	w := a.NewWindow("Clipboard Manager")
	w.Resize(fyne.NewSize(600, 500))
	w.CenterOnScreen()
	
	// Request focus to bring window to front
	w.RequestFocus()
	
	fmt.Println("Window created, loading history...")

	hist, err := storage.LoadHistory()
	if err != nil {
		fmt.Printf("Error loading history: %v\n", err)
		return err
	}
	fmt.Printf("Loaded %d history items\n", len(hist))

	// State
	originalHist := hist
	filtered := make([]int, len(originalHist))
	for i := range originalHist {
		filtered[i] = i
	}

	// UI Components
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search clipboard history...")

	statusLabel := widget.NewLabel(fmt.Sprintf("Showing %d of %d items", len(filtered), len(originalHist)))

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
				if idx < len(originalHist) {
					preview := strings.Split(originalHist[idx], "\n")[0]
					if len(preview) > 80 {
						preview = preview[:80] + "…"
					}
					o.(*widget.Label).SetText(fmt.Sprintf("[%d] %s", idx, preview))
				}
			}
		},
	)

	var selectedIndex int = -1
	list.OnSelected = func(id widget.ListItemID) {
		selectedIndex = id
	}

	// Refresh helper
	refreshList := func() {
		list.Refresh()
		statusLabel.SetText(fmt.Sprintf("Showing %d of %d items", len(filtered), len(originalHist)))
	}

	// Filter function
	applyFilter := func(query string) {
		query = strings.TrimSpace(strings.ToLower(query))
		if query == "" {
			filtered = make([]int, len(originalHist))
			for i := range originalHist {
				filtered[i] = i
			}
		} else {
			tmp := []int{}
			for i, v := range originalHist {
				if strings.Contains(strings.ToLower(v), query) {
					tmp = append(tmp, i)
				}
			}
			filtered = tmp
		}
		selectedIndex = -1
		refreshList()
	}

	searchEntry.OnChanged = applyFilter

	// Action buttons
	copyBtn := widget.NewButton("Copy", func() {
		if selectedIndex >= 0 && selectedIndex < len(filtered) {
			idx := filtered[selectedIndex]
			if idx < len(originalHist) {
				if err := clipboardPkg.CopyToClipboard(originalHist[idx]); err != nil {
					dialog.ShowError(err, w)
				} else {
					statusLabel.SetText(fmt.Sprintf("Copied item %d to clipboard", idx))
				}
			}
		} else {
			statusLabel.SetText("No item selected")
		}
	})

	pasteBtn := widget.NewButton("Paste", func() {
		if selectedIndex >= 0 && selectedIndex < len(filtered) {
			idx := filtered[selectedIndex]
			if idx < len(originalHist) {
				if err := clipboardPkg.Paste(originalHist[idx]); err != nil {
					dialog.ShowError(err, w)
				} else {
					statusLabel.SetText(fmt.Sprintf("Pasted item %d", idx))
				}
			}
		} else {
			statusLabel.SetText("No item selected")
		}
	})

	deleteBtn := widget.NewButton("Delete", func() {
		if selectedIndex >= 0 && selectedIndex < len(filtered) {
			idx := filtered[selectedIndex]
			if idx < len(originalHist) {
				// Remove from history
				originalHist = append(originalHist[:idx], originalHist[idx+1:]...)

				// Rebuild filtered
				filtered = make([]int, len(originalHist))
				for i := range originalHist {
					filtered[i] = i
				}

				// Save to disk
				if err := storage.SaveHistory(originalHist); err != nil {
					dialog.ShowError(err, w)
					return
				}

				selectedIndex = -1
				applyFilter(searchEntry.Text)
				statusLabel.SetText(fmt.Sprintf("Deleted item %d", idx))
			}
		} else {
			statusLabel.SetText("No item selected")
		}
	})

	clearBtn := widget.NewButton("Clear All", func() {
		dialog.ShowConfirm("Confirm Clear", "Delete all clipboard history?", func(confirm bool) {
			if confirm {
				originalHist = []string{}
				filtered = []int{}
				if err := storage.SaveHistory(originalHist); err != nil {
					dialog.ShowError(err, w)
					return
				}
				refreshList()
				statusLabel.SetText("History cleared")
			}
		}, w)
	})

	refreshBtn := widget.NewButton("Refresh", func() {
		newHist, err := storage.LoadHistory()
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		originalHist = newHist
		filtered = make([]int, len(originalHist))
		for i := range originalHist {
			filtered[i] = i
		}
		searchEntry.SetText("")
		selectedIndex = -1
		refreshList()
		statusLabel.SetText(fmt.Sprintf("Refreshed: %d items", len(originalHist)))
	})

	// Button toolbar
	toolbar := container.NewHBox(
		copyBtn,
		pasteBtn,
		deleteBtn,
		widget.NewSeparator(),
		refreshBtn,
		clearBtn,
	)

	// Layout
	content := container.NewBorder(
		container.NewVBox(searchEntry, toolbar), // top
		statusLabel,                              // bottom
		nil,                                      // left
		nil,                                      // right
		list,                                     // center
	)

	w.SetContent(content)
	
	fmt.Println("Window setup complete, showing window...")
	w.Show()
	fmt.Println("Window should be visible now. Running event loop...")
	a.Run()
	fmt.Println("Application closed.")
	return nil
}
