package art

import (
	"testing"

	"github.com/mattn/go-runewidth"
)

func TestAnimConnecting(t *testing.T) {
	out := RenderConnecting(0, "0.1.0", SceneOpts{Mode: MascotConnecting, ShowMascot: false, ReduceMotion: true})
	if out == "" || !contains(out, "Connecting") {
		t.Fatalf("unexpected connecting scene: %q", out)
	}
}

func TestScanBeamBounds(t *testing.T) {
	for frame := 0; frame < 200; frame++ {
		out := ScanBeam(frame, 46)
		if runewidth.StringWidth(out) != 46 {
			t.Fatalf("frame %d: width %d want 46: %q", frame, runewidth.StringWidth(out), out)
		}
	}
}

func TestProgressBar(t *testing.T) {
	bar := ProgressBar(10, 0.5, 3)
	if len(bar) < 12 {
		t.Fatalf("bar too short: %q", bar)
	}
}

func TestAnimHeaderWidth(t *testing.T) {
	out := AnimHeader(0, "0.1.0", "ready", false, SceneOpts{ShowMascot: true})
	for _, line := range stringsLines(out) {
		if runewidth.StringWidth(line) != boxWidth {
			t.Fatalf("header line width %d want %d: %q", runewidth.StringWidth(line), boxWidth, line)
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func stringsLines(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}
