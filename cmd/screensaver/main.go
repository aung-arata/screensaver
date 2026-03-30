// Package main is the entry point for the screensaver screenshot tool.
//
// Usage:
//
//	screensaver              # run as background daemon with global hotkey
//	screensaver --once       # capture a single screenshot and exit
//	screensaver --select     # interactive region selection (overlay)
//	screensaver --edit       # open annotation editor after capture (with --once or --select)
//	screensaver --hotkey "ctrl+shift+p"  # use a custom hotkey
package main

import (
	"flag"
	"fmt"
	"image"
	"os"

	"github.com/aung-arata/screensaver/internal/capture"
	"github.com/aung-arata/screensaver/internal/clipboard"
	"github.com/aung-arata/screensaver/internal/editor"
	"github.com/aung-arata/screensaver/internal/overlay"
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

// runDaemon prints user-facing daemon-mode instructions and the configured hotkey.
// The function does not start any GUI, hotkey listener, or tray; those platform-specific
// components are implemented elsewhere and must run in a GUI environment.
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
