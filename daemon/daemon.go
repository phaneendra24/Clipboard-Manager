// Package daemon provides the background clipboard polling service.
package daemon

import (
	"log"
	"strings"
	"time"

	"github.com/atotto/clipboard"

	"github/phaneendra24/goclipboard-manager/storage"
)

// Run starts the daemon that polls the clipboard at the given interval.
// It saves new clipboard contents to history and logs activity.
// The daemon runs until stopCh is closed.
func Run(pollMS int, logger *log.Logger, stopCh <-chan struct{}) error {
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
				logger.Printf("clipboard read error: %v\n", err)
				continue
			}
			// Ignore empty strings
			if strings.TrimSpace(txt) == "" {
				continue
			}
			if txt == lastSeen {
				continue // no change
			}

			clipData, err := storage.LoadClipboardData()
			if err != nil {
				logger.Printf("load history error: %v\n", err)
				continue
			}

			// Check if already exists anywhere in history
			existsAt := -1
			for i, item := range clipData.History {
				if item == txt {
					existsAt = i
					break
				}
			}

			if existsAt == 0 {
				// Already at top, no change needed
				lastSeen = txt
				continue
			}

			if existsAt > 0 {
				// Remove from old position
				clipData.History = append(clipData.History[:existsAt], clipData.History[existsAt+1:]...)
			}

			// Add to front
			clipData.History = append([]string{txt}, clipData.History...)
			if len(clipData.History) > storage.MaxHistory {
				clipData.History = clipData.History[:storage.MaxHistory]
			}

			if err := storage.SaveClipboardData(clipData); err != nil {
				logger.Printf("save history error: %v\n", err)
				continue
			}
			lastSeen = txt
			logger.Printf("captured clipboard (len=%d) preview: %q\n", len(clipData.History), storage.Preview(txt, 80))
		}
	}
}

