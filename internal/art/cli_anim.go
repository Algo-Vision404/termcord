package art

import (
	"fmt"
	"io"
	"time"

	"github.com/termcord/termcord/internal/ds"
)

// RunWithSpinner runs fn while showing an animated spinner on stderr.
func RunWithSpinner(label string, fn func() error) error {
	return ds.RunWithSpinner(label, fn)
}

// PrintBootBanner prints the CLI boot splash before the TUI starts.
func PrintBootBanner(w io.Writer, version string, animated bool) {
	fmt.Fprint(w, ds.BootSplash(version))
	if animated {
		time.Sleep(120 * time.Millisecond)
	}
}

// PrintBootStep prints a bullet status line during startup.
func PrintBootStep(w io.Writer, msg string) {
	fmt.Fprint(w, ds.BootStep(msg))
}

// PrintAnimatedLogo is kept for compatibility; boot uses the figlet splash now.
func PrintAnimatedLogo(w io.Writer, version string) {
	PrintBootBanner(w, version, true)
}

// CLIDoctorProgress animates doctor checks line by line.
func CLIDoctorProgress(w io.Writer, checks []struct{ Label, Result string }) {
	for _, check := range checks {
		ds.PrintDoctorCheck(w, check.Label, check.Result)
	}
}

// CLISpinnerFrame returns one frame of a CLI progress spinner.
func CLISpinnerFrame(frame int) string {
	return ds.SpinnerFrame(frame)
}
