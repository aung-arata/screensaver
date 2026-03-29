// Package main is the entry point for the screensaver screenshot tool.
//
// Usage:
//
//	screensaver              # run as background daemon with global hotkey
//	screensaver --once       # capture a single screenshot and exit
//	screensaver --select     # interactive region selection (overlay)
//	screensaver --hotkey "ctrl+shift+p"  # use a custom hotkey
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/aung-arata/screensaver/internal/capture"
	"github.com/aung-arata/screensaver/internal/clipboard"
	"github.com/aung-arata/screensaver/internal/overlay"
	"github.com/aung-arata/screensaver/internal/utils"
)

// Version is the application version, set at build time via ldflags.
var Version = "0.2.0"

func main() {
	once := flag.Bool("once", false, "Capture one screenshot and exit (no background daemon)")
	sel := flag.Bool("select", false, "Interactive region selection: dims the screen and lets you drag a rectangle")
	hotkey := flag.String("hotkey", "ctrl+shift+s", "Global hotkey combination (e.g. 'ctrl+shift+s')")
	output := flag.String("output", "", "Save screenshot to this path (only with --once or --select)")
	version := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *version {
		fmt.Printf("screensaver %s\n", Version)
		os.Exit(0)
	}

	if *sel {
		runSelect(*output)
		return
	}

	if *once {
		runOnce(*output)
		return
	}

	runDaemon(*hotkey)
}

// runOnce captures a full-screen screenshot and saves or copies it.
func runOnce(outputPath string) {
	img, err := capture.FullScreen(0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
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

// runSelect shows the fullscreen selection overlay, lets the user draw
// a rubber-band rectangle, and then captures the selected region.
func runSelect(outputPath string) {
	result, err := overlay.Show(0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if result.Cancelled {
		fmt.Println("Selection cancelled")
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

// runDaemon starts the background hotkey listener and system tray.
func runDaemon(hotkey string) {
	fmt.Printf("[screensaver] Running in the background.\n")
	fmt.Printf("  Press %s to take a screenshot.\n", hotkey)
	fmt.Printf("  Press Ctrl+C to quit.\n")

	// NOTE: Full GUI overlay, system tray, and hotkey listener require
	// platform-specific Win32 APIs or a GUI toolkit (Fyne / Walk).
	// This scaffold provides the architecture; platform-specific
	// implementations are in internal/hotkey, internal/overlay,
	// internal/editor, and internal/tray packages.
	fmt.Println("[screensaver] Daemon mode requires a GUI environment (Windows/Linux/macOS).")
	fmt.Println("[screensaver] Use --once for headless capture or --select for interactive selection.")
}
