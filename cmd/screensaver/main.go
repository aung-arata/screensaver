// Package main is the entry point for the screensaver screenshot tool.
//
// Usage:
//
//	screensaver                    # run as background daemon with global hotkey + tray icon
//	screensaver --once             # capture a single screenshot and exit
//	screensaver --select           # interactive region selection (overlay, Windows only)
//	screensaver --once --edit      # capture full screen and open annotation editor
//	screensaver --select --edit    # select a region and open annotation editor (Windows only)
//	screensaver --hotkey "ctrl+shift+p"  # use a custom hotkey
//	screensaver --once --save-dir /path  # auto-save to directory with timestamped filename
//	screensaver --select --save-dir /path  # select region and auto-save to directory
//	screensaver --format jpeg --quality 85 --once  # save as JPEG with custom quality
//	screensaver config show        # print current effective config as YAML
//	screensaver config init        # write default config to config file
//	screensaver config path        # print the config file path
package main

import (
	"flag"
	"fmt"
	"image"
	"os"
	"sync"

	"github.com/aung-arata/screensaver/internal/capture"
	"github.com/aung-arata/screensaver/internal/clipboard"
	"github.com/aung-arata/screensaver/internal/config"
	"github.com/aung-arata/screensaver/internal/editor"
	"github.com/aung-arata/screensaver/internal/hotkey"
	"github.com/aung-arata/screensaver/internal/overlay"
	"github.com/aung-arata/screensaver/internal/tray"
	"github.com/aung-arata/screensaver/internal/utils"
)

// Version is the application version, set at build time via ldflags.
var Version = "0.2.0"

// lastSavedPath holds the path of the most recently saved screenshot.
// It is updated on every successful save (from the editor or CLI flags).
var (
	lastSavedPath   string
	lastSavedPathMu sync.Mutex
)

// setLastSavedPath stores path as the most recently saved screenshot path.
func setLastSavedPath(path string) {
	lastSavedPathMu.Lock()
	lastSavedPath = path
	lastSavedPathMu.Unlock()
}

// getLastSavedPath returns the most recently saved screenshot path.
func getLastSavedPath() string {
	lastSavedPathMu.Lock()
	defer lastSavedPathMu.Unlock()
	return lastSavedPath
}

// main is the entry point for the screensaver CLI.
// It loads the config file, parses command-line flags, merges them, then
// dispatches the selected mode:
// - config <sub>: config file management subcommand.
// - --version: prints the build-time Version and exits.
// - --select: shows an interactive fullscreen region selector and captures the chosen area.
// - --once: captures a full-screen screenshot once.
// - default: starts daemon mode which registers a global hotkey (configurable via --hotkey).
// The --output flag, when provided with --once or --select, saves the captured image to the given path.
// The --edit flag, when combined with --once or --select, opens the annotation editor after capture.
func main() {
	once       := flag.Bool("once", false, "Capture one screenshot and exit (no background daemon)")
	sel        := flag.Bool("select", false, "Interactive region selection: dims the screen and lets you drag a rectangle")
	hotkeyFlag := flag.String("hotkey", "", "Global hotkey combination (e.g. 'ctrl+shift+s'); overrides config file; defaults to config value or 'ctrl+shift+s'")
	output     := flag.String("output", "", "Save screenshot to this path (only with --once or --select)")
	saveDir    := flag.String("save-dir", "", "Auto-save to this directory with a timestamped filename (only with --once or --select; --output takes precedence when both are set)")
	edit       := flag.Bool("edit", false, "Open the annotation editor after capture (use with --once or --select)")
	version    := flag.Bool("version", false, "Print version and exit")
	format     := flag.String("format", "", "Output format: png or jpeg (default from config, fallback \"png\")")
	quality    := flag.Int("quality", 0, "JPEG quality 1–100 (default from config, fallback 90; ignored for PNG)")
	configPath := flag.String("config", "", "Path to config file (overrides default location)")
	flag.Parse()

	// Load file-based config (or defaults if no file).
	var fileCfg config.Config
	var loadErr error
	if *configPath != "" {
		fileCfg, loadErr = config.LoadFrom(*configPath)
	} else {
		fileCfg, loadErr = config.Load()
	}
	if loadErr != nil {
		fmt.Fprintf(os.Stderr, "warning: could not load config: %v\n", loadErr)
		fileCfg = config.DefaultConfig()
	}

	// Build CLI overrides and merge — CLI flags always win.
	cliOverrides := config.Config{
		Hotkey:  *hotkeyFlag,
		SaveDir: *saveDir,
		Format:  *format,
		Quality: *quality,
	}
	cfg := config.Merge(fileCfg, cliOverrides)

	// Handle "config" subcommand before all other flags.
	if len(flag.Args()) > 0 && flag.Args()[0] == "config" {
		runConfigCmd(flag.Args()[1:], *configPath, cfg)
		return
	}

	if *version {
		fmt.Printf("screensaver %s\n", Version)
		os.Exit(0)
	}

	if *sel {
		runSelect(*output, *edit, cfg)
		return
	}

	if *once {
		runOnce(*output, *edit, cfg)
		return
	}

	runDaemon(cfg.Hotkey)
}

// openEditorAndExit opens the annotation editor for img.
// On error it writes to stderr and exits with status 1.
func openEditorAndExit(img image.Image) {
	e := editor.New(img)
	e.OnSave = func(p string) { setLastSavedPath(p) }
	if err := e.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error running editor: %v\n", err)
		os.Exit(1)
	}
}

// runOnce captures a full-screen screenshot and saves it to outputPath or copies it to the clipboard.
// If openEditor is true the annotation editor is opened instead of the default copy/save behaviour.
// If outputPath is non-empty the image is written there; otherwise if cfg.SaveDir is non-empty the image
// is auto-saved with a timestamped filename; otherwise the image is copied to the clipboard.
// On error it writes a message to stderr and exits the process with status 1.
func runOnce(outputPath string, openEditor bool, cfg config.Config) {
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
		if err := utils.SaveImage(img, outputPath, cfg.Quality); err != nil {
			fmt.Fprintf(os.Stderr, "error saving image: %v\n", err)
			os.Exit(1)
		}
		setLastSavedPath(outputPath)
		fmt.Printf("Screenshot saved to %s\n", outputPath)
		return
	}

	if cfg.SaveDir != "" {
		path, err := utils.GenerateFilename(cfg.SaveDir, cfg.Format)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error generating filename: %v\n", err)
			os.Exit(1)
		}
		if err := utils.SaveImage(img, path, cfg.Quality); err != nil {
			fmt.Fprintf(os.Stderr, "error saving image: %v\n", err)
			os.Exit(1)
		}
		setLastSavedPath(path)
		fmt.Printf("Screenshot saved to %s\n", path)
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
// saves the resulting image to outputPath, auto-saves to cfg.SaveDir, or copies it to the clipboard.
// If outputPath is non-empty the image is written to that path; otherwise if cfg.SaveDir is non-empty
// the image is auto-saved with a timestamped filename; otherwise it is copied to the clipboard.
// If the user cancels selection the function returns without producing an image; on capture, save,
// or clipboard errors the process prints an error to stderr and exits with status 1.
func runSelect(outputPath string, openEditor bool, cfg config.Config) {
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
		if err := utils.SaveImage(img, outputPath, cfg.Quality); err != nil {
			fmt.Fprintf(os.Stderr, "error saving image: %v\n", err)
			os.Exit(1)
		}
		setLastSavedPath(outputPath)
		fmt.Printf("Screenshot saved to %s\n", outputPath)
		return
	}

	if cfg.SaveDir != "" {
		path, err := utils.GenerateFilename(cfg.SaveDir, cfg.Format)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error generating filename: %v\n", err)
			os.Exit(1)
		}
		if err := utils.SaveImage(img, path, cfg.Quality); err != nil {
			fmt.Fprintf(os.Stderr, "error saving image: %v\n", err)
			os.Exit(1)
		}
		setLastSavedPath(path)
		fmt.Printf("Screenshot saved to %s\n", path)
		return
	}

	if err := clipboard.CopyImage(img); err != nil {
		fmt.Fprintf(os.Stderr, "error copying to clipboard: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Region screenshot copied to clipboard")
}

// runDaemon registers a global hotkey and a system tray icon then blocks
// until the user quits via the tray menu or sends SIGTERM / Ctrl+C.
//
// Pressing the hotkey or choosing "Take Screenshot" from the tray opens the
// annotation editor with a full-screen capture.  Choosing "Select Region"
// shows the overlay selection first.
func runDaemon(combo string) {
	quit := make(chan struct{})
	var once sync.Once
	closeQuit := func() { once.Do(func() { close(quit) }) }

	// Start the hotkey listener in a background goroutine.
	hl := hotkey.NewListener(combo, func() { go captureAndEdit() })
	go func() {
		if err := hl.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "[hotkey] %v\n", err)
		}
	}()

	// Run the system tray (blocks until Quit is selected).
	cfg := tray.Config{
		Tooltip:    fmt.Sprintf("Screensaver – press %s to capture", combo),
		OnCapture:  func() { go captureAndEdit() },
		OnSelect:   func() { go selectAndEdit() },
		OnOpenLast: func() { go openLastScreenshot(getLastSavedPath()) },
		OnQuit:     closeQuit,
	}
	go func() {
		if err := tray.Run(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "tray: %v\n", err)
		}
		closeQuit()
	}()

	<-quit
	hl.Stop()
}

// captureAndEdit takes a full-screen screenshot and opens the annotation
// editor.  Errors are printed to stderr; the process is not terminated.
func captureAndEdit() {
	img, err := capture.FullScreen(0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "capture: %v\n", err)
		return
	}
	e := editor.New(img)
	e.OnSave = func(p string) { setLastSavedPath(p) }
	if err := e.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "editor: %v\n", err)
	}
}

// selectAndEdit shows the region-selection overlay and, if the user draws a
// rectangle, captures that region and opens the annotation editor.
func selectAndEdit() {
	result, err := overlay.Show(0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "overlay: %v\n", err)
		return
	}
	if result.Cancelled || result.Region.Dx() == 0 || result.Region.Dy() == 0 {
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
		fmt.Fprintf(os.Stderr, "capture region: %v\n", err)
		return
	}
	e := editor.New(img)
	e.OnSave = func(p string) { setLastSavedPath(p) }
	if err := e.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "editor: %v\n", err)
	}
}

// runConfigCmd handles the "config" subcommand.
//
// Supported sub-subcommands:
//   - show  — print the current effective config as YAML to stdout
//   - init  — write DefaultConfig() to the config file (errors if file exists, unless --force)
//   - path  — print the config file path
func runConfigCmd(args []string, configFilePath string, effectiveCfg config.Config) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: screensaver config <show|init|path>")
		os.Exit(1)
	}

	resolvedPath := configFilePath
	if resolvedPath == "" {
		p, err := config.ConfigPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error resolving config path: %v\n", err)
			os.Exit(1)
		}
		resolvedPath = p
	}

	switch args[0] {
	case "show":
		data, err := config.EncodeYAML(effectiveCfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error encoding config: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(string(data))

	case "init":
		force := len(args) > 1 && args[1] == "--force"
		if _, err := os.Stat(resolvedPath); err == nil && !force {
			fmt.Fprintf(os.Stderr, "config file already exists at %s (use --force to overwrite)\n", resolvedPath)
			os.Exit(1)
		}
		if err := config.SaveTo(config.DefaultConfig(), resolvedPath); err != nil {
			fmt.Fprintf(os.Stderr, "error writing config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Config written to %s\n", resolvedPath)

	case "path":
		fmt.Println(resolvedPath)

	default:
		fmt.Fprintf(os.Stderr, "unknown config subcommand %q\n", args[0])
		fmt.Fprintln(os.Stderr, "usage: screensaver config <show|init|path>")
		os.Exit(1)
	}
}
