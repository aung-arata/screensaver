// Package main is the entry point for the screensaver screenshot tool.
//
// Usage:
//
//	screensaver                    # run as background daemon with global hotkey
//	screensaver --once             # capture a single screenshot and exit
//	screensaver --select           # interactive region selection (overlay, Windows only)
//	screensaver --once --edit      # capture full screen and open annotation editor
//	screensaver --select --edit    # select a region and open annotation editor (Windows only)
//	screensaver --hotkey "ctrl+shift+p"  # use a custom hotkey
package main

import (
	"flag"
	"fmt"
	"image"
	"os"
	"os/signal"
	"sync"

	"github.com/aung-arata/screensaver/internal/capture"
	"github.com/aung-arata/screensaver/internal/clipboard"
	"github.com/aung-arata/screensaver/internal/editor"
	"github.com/aung-arata/screensaver/internal/hotkey"
	"github.com/aung-arata/screensaver/internal/overlay"
	"github.com/aung-arata/screensaver/internal/tray"
	"github.com/aung-arata/screensaver/internal/utils"
)

// Version is the application version, set at build time via ldflags.
var Version = "0.2.0"

// main is the entry point for the screensaver CLI.
// It parses command-line flags and dispatches the selected mode:
// - --version: prints the build-time Version and exits.
// - --select: shows an interactive fullscreen region selector and captures the chosen area.
// - --once: captures a full-screen screenshot once.
// - default: starts daemon mode which registers a global hotkey (configurable via --hotkey).
// The --output flag, when provided with --once or --select, saves the captured image to the given path.
// The --edit flag, when combined with --once or --select, opens the annotation editor after capture.
func main() {
	once := flag.Bool("once", false, "Capture one screenshot and exit (no background daemon)")
	sel := flag.Bool("select", false, "Interactive region selection: dims the screen and lets you drag a rectangle")
	hotkey := flag.String("hotkey", "ctrl+shift+s", "Global hotkey combination (e.g. 'ctrl+shift+s')")
	output := flag.String("output", "", "Save screenshot to this path (only with --once or --select)")
	edit := flag.Bool("edit", false, "Open the annotation editor after capture (use with --once or --select)")
	version := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *version {
		fmt.Printf("screensaver %s\n", Version)
		os.Exit(0)
	}

	if *sel {
		runSelect(*output, *edit)
		return
	}

	if *once {
		runOnce(*output, *edit)
		return
	}

	runDaemon(*hotkey)
}

// openEditorAndExit opens the annotation editor for img.
// On error it writes to stderr and exits with status 1.
func openEditorAndExit(img image.Image) {
	e := editor.New(img)
	if err := e.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error running editor: %v\n", err)
		os.Exit(1)
	}
}

// runOnce captures a full-screen screenshot and saves it to outputPath or copies it to the clipboard.
// If openEditor is true the annotation editor is opened instead of the default copy/save behaviour.
// If outputPath is non-empty the image is written there; otherwise the image is copied to the clipboard.
// On error it writes a message to stderr and exits the process with status 1.
func runOnce(outputPath string, openEditor bool) {
	img, err := capture.FullScreen(0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if openEditor {
		openEditorAndExit(img)
		return
	}

	if outputPath != "" {
		if err := utils.SaveImage(img, outputPath); err != nil {
			fmt.Fprintf(os.Stderr, "error saving image: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Screenshot saved to %s\n", outputPath)
		return
	}

	// Default: copy to clipboard.
	if err := clipboard.CopyImage(img); err != nil {
		fmt.Fprintf(os.Stderr, "error copying to clipboard: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Screenshot copied to clipboard")
}

// runSelect displays a fullscreen region-selection overlay that lets the user draw a rectangle,
// captures the selected area, and either opens the annotation editor (if openEditor is true),
// saves the resulting image to outputPath, or copies it to the clipboard. If outputPath is
// non-empty the image is written to that path; otherwise it is copied to the clipboard. If the
// user cancels selection the function returns without producing an image; on capture, save, or
// clipboard errors the process prints an error to stderr and exits with status 1.
func runSelect(outputPath string, openEditor bool) {
	result, err := overlay.Show(0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if result.Cancelled {
		fmt.Println("Selection cancelled")
		return
	}

	if result.Region.Dx() == 0 || result.Region.Dy() == 0 {
		fmt.Println("No selection")
		return
	}

	region := capture.Region{
		X:      result.Region.Min.X,
		Y:      result.Region.Min.Y,
		Width:  result.Region.Dx(),
		Height: result.Region.Dy(),
	}
	img, err := capture.CaptureRegion(region, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error capturing region: %v\n", err)
		os.Exit(1)
	}

	if openEditor {
		openEditorAndExit(img)
		return
	}

	if outputPath != "" {
		if err := utils.SaveImage(img, outputPath); err != nil {
			fmt.Fprintf(os.Stderr, "error saving image: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Screenshot saved to %s\n", outputPath)
		return
	}

	if err := clipboard.CopyImage(img); err != nil {
		fmt.Fprintf(os.Stderr, "error copying to clipboard: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Region screenshot copied to clipboard")
}

// runDaemon starts the background daemon: it registers a global hotkey,
// starts a system tray icon (if available), and waits until a termination
// signal is received or the tray "Quit" item is clicked.
//
// On non-Windows platforms the hotkey listener is not yet available and
// the function exits with an error message suggesting --once mode instead.
func runDaemon(hotkeyCombo string) {
	fmt.Printf("[screensaver] Running in the background.\n")
	fmt.Printf("  Press %s to take a screenshot.\n", hotkeyCombo)
	fmt.Printf("  Press Ctrl+C to quit.\n")

	// Shared callback: capture the full screen and copy to clipboard.
	captureAndCopy := func() {
		img, err := capture.FullScreen(0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[screensaver] capture error: %v\n", err)
			return
		}
		if err := clipboard.CopyImage(img); err != nil {
			fmt.Fprintf(os.Stderr, "[screensaver] clipboard error: %v\n", err)
			return
		}
		fmt.Println("[screensaver] Screenshot copied to clipboard")
	}

	// Channel closed when the tray "Quit" item is clicked.
	quit := make(chan struct{})
	var quitOnce sync.Once

	// Start system tray (optional — not available on all platforms).
	trayCfg := tray.DefaultConfig()
	go func() {
		err := tray.Run(trayCfg, tray.Callbacks{
			OnCapture: captureAndCopy,
			OnQuit: func() {
				quitOnce.Do(func() { close(quit) })
			},
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "[screensaver] System tray: %v\n", err)
		}
	}()

	// Start global hotkey listener.
	listener := hotkey.NewListener(hotkeyCombo, captureAndCopy)
	errCh := make(chan error, 1)
	go func() {
		errCh <- listener.Start()
	}()

	// Handle OS signals for graceful shutdown.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	select {
	case err := <-errCh:
		if err != nil {
			fmt.Fprintf(os.Stderr, "[screensaver] %v\n", err)
			fmt.Println("[screensaver] Use --once for headless capture or --select for interactive selection.")
			os.Exit(1)
		}
	case sig := <-sigCh:
		fmt.Printf("\n[screensaver] Received %v, shutting down...\n", sig)
		listener.Stop()
	case <-quit:
		fmt.Println("[screensaver] Quit requested via tray, shutting down...")
		listener.Stop()
	}
}
