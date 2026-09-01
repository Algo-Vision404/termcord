package ds

import (
	"fmt"
	"sync"
)

var (
	asciiOnce sync.Once
	asciiArt  string
)

// AsciiWordmark returns the Big figlet boot wordmark with a > prompt.
func AsciiWordmark() string {
	asciiOnce.Do(func() {
		asciiArt = renderWordmark()
	})
	return asciiArt
}

// BootSplash is the opening banner: wordmark + tagline.
func BootSplash(version string) string {
	return fmt.Sprintf("\n%s\n\n    %s v%s\n\n", AsciiWordmark(), BootTagline, version)
}

// BootStep formats a single startup status line.
func BootStep(msg string) string {
	return fmt.Sprintf("  • %s\n", msg)
}

// CLILogo returns the full banner for login, doctor, and version commands.
func CLILogo(version string) string {
	return BootSplash(version)
}

// CLIMark is the one-line identity used in compact headers.
func CLIMark(version string) string {
	return fmt.Sprintf("termcord v%s", version)
}
