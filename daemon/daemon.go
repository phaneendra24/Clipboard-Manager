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
			hist, err := storage.LoadHistory()
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
			if len(newHist) > storage.MaxHistory {
				newHist = newHist[:storage.MaxHistory]
			}
			if err := storage.SaveHistory(newHist); err != nil {
				logger.Printf("save history error: %v\n", err)
				continue
			}
			lastSeen = txt
			logger.Printf("captured clipboard (len=%d) preview: %q\n", len(newHist), storage.Preview(txt, 80))
		}
	}
}
