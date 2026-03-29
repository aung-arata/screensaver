// Package clipboard provides image-to-clipboard functionality.
//
// It wraps the atotto/clipboard library for text clipboard operations
// and uses platform-specific approaches for image data.
package clipboard

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"
	"os/exec"
	"runtime"
)

// CopyImage copies an image to the system clipboard.
//
// On Linux this requires xclip to be installed. On macOS it uses osascript.
// On Windows it uses PowerShell to interact with the clipboard.
func CopyImage(img image.Image) error {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return fmt.Errorf("encoding image: %w", err)
	}

	switch runtime.GOOS {
	case "linux":
		return copyImageLinux(buf.Bytes())
	case "darwin":
		return copyImageMacOS(buf.Bytes())
	case "windows":
		return copyImageWindows(buf.Bytes())
	default:
		return fmt.Errorf("clipboard: unsupported platform %s", runtime.GOOS)
	}
}

func copyImageLinux(pngData []byte) error {
	// Try xclip first, then xsel.
	for _, args := range [][]string{
		{"xclip", "-selection", "clipboard", "-t", "image/png"},
		{"xsel", "--clipboard", "--input"},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdin = bytes.NewReader(pngData)
		if err := cmd.Run(); err == nil {
			return nil
		}
	}
	return fmt.Errorf("clipboard: xclip or xsel required on Linux")
}

func copyImageMacOS(pngData []byte) error {
	// Write to a temp file, then use osascript.
	f, err := os.CreateTemp("", "screensaver-*.png")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())

	if _, err := f.Write(pngData); err != nil {
		f.Close()
		return err
	}
	f.Close()

	script := fmt.Sprintf(`set the clipboard to (read (POSIX file %q) as «class PNGf»)`, f.Name())
	cmd := exec.Command("osascript", "-e", script)
	return cmd.Run()
}

func copyImageWindows(pngData []byte) error {
	// Write to a temp file, then use PowerShell to set clipboard.
	f, err := os.CreateTemp("", "screensaver-*.png")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())

	if _, err := f.Write(pngData); err != nil {
		f.Close()
		return err
	}
	f.Close()

	script := fmt.Sprintf(
		`Add-Type -AssemblyName System.Windows.Forms; `+
			`[System.Windows.Forms.Clipboard]::SetImage([System.Drawing.Image]::FromFile('%s'))`,
		f.Name(),
	)
	cmd := exec.Command("powershell", "-Command", script)
	return cmd.Run()
}
