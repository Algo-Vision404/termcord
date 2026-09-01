package art

import (
	"strings"
	"testing"
)

func TestCLIBoxMinWidth(t *testing.T) {
	out := CLIBox("doctor", []string{"rest: ok"})
	if len(out) < 44 {
		t.Fatalf("box too narrow: %q", out)
	}
}

func TestSectionBox(t *testing.T) {
	box := SectionBox("Direct Messages", 32)
	if !strings.HasPrefix(box, "╭─ Direct Messages") {
		t.Fatalf("unexpected box header: %q", box)
	}
}

func TestHelpBlock(t *testing.T) {
	if HelpBlock() == "" {
		t.Fatal("empty help block")
	}
}
