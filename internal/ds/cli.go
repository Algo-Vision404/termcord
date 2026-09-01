package ds

import (
	"fmt"
	"io"
	"os"
	"time"
)

var cliSpinner = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// SpinnerFrame returns one frame of the CLI braille spinner.
func SpinnerFrame(frame int) string {
	return cliSpinner[frame%len(cliSpinner)]
}

// RunWithSpinner runs fn while showing an animated spinner on stderr.
func RunWithSpinner(label string, fn func() error) error {
	done := make(chan error, 1)
	go func() { done <- fn() }()

	frame := 0
	for {
		select {
		case err := <-done:
			if err != nil {
				fmt.Fprintf(os.Stderr, "\r✗ %s\n", label)
			} else {
				fmt.Fprintf(os.Stderr, "\r✓ %s\n", label)
			}
			return err
		default:
			fmt.Fprintf(os.Stderr, "\r%s %s", SpinnerFrame(frame), label)
			frame++
			time.Sleep(80 * time.Millisecond)
		}
	}
}

// PrintDoctorCheck animates one doctor check line.
func PrintDoctorCheck(w io.Writer, label, result string) {
	for f := 0; f < 6; f++ {
		fmt.Fprintf(w, "\r  %s %s", SpinnerFrame(f), label)
		time.Sleep(50 * time.Millisecond)
	}
	fmt.Fprintf(w, "\r  ✓ %s %s\n", PadRight(label, 28), result)
}
