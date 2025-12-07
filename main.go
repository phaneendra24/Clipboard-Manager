// main.go - Clipboard Manager CLI entry point
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	clipboardPkg "github/phaneendra24/goclipboard-manager/clipboard"
	"github/phaneendra24/goclipboard-manager/daemon"
	"github/phaneendra24/goclipboard-manager/storage"
	"github/phaneendra24/goclipboard-manager/ui"
)

const defaultPollMS = 300 // poll interval in ms

func printUsage() {
	fmt.Println(`Usage: clipcli <command>

Commands:
  serve [poll_ms]   Run daemon (poll_ms optional, default 300)
  save              Save current clipboard to history
  list              List history previews
  paste N           Paste history item N (0 = most recent)
  clear             Clear history
  gui               Open graphical clipboard manager`)
}

func cmdList() error {
	hist, err := storage.LoadHistory()
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

func main() {
	// Configure logging to file (if possible) else stdout
	logPath, _ := storage.LogFilePath()
	logf, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
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
		// Handle signals for graceful shutdown
		stop := make(chan struct{})
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			sig := <-sigs
			logger.Printf("received signal %v, shutting down\n", sig)
			close(stop)
		}()
		if err := daemon.Run(pollMS, logger, stop); err != nil {
			logger.Fatalf("daemon error: %v\n", err)
		}

	case "save":
		if err := clipboardPkg.Save(); err != nil {
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
		if err := clipboardPkg.PasteByIndex(i); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(2)
		}
		fmt.Printf("pasted index %d\n", i)

	case "clear":
		if err := storage.SaveHistory([]string{}); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(2)
		}
		fmt.Println("history cleared")

	case "gui":
		if err := ui.RunGUI(); err != nil {
			fmt.Fprintln(os.Stderr, "gui error:", err)
			os.Exit(2)
		}

	default:
		printUsage()
		os.Exit(1)
	}
}
