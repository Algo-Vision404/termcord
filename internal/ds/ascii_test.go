package ds

import (
	"strings"
	"testing"
)

func TestAsciiWordmark(t *testing.T) {
	art := AsciiWordmark()
	if strings.Contains(art, "888") {
		t.Fatal("wordmark should not use 888 block styling")
	}
	if !strings.Contains(art, "_ __ ___") {
		t.Fatalf("wordmark missing big figlet art: %q", art)
	}
	if !strings.Contains(art, ">") {
		t.Fatal("wordmark should include > prompt")
	}
	lines := strings.Split(strings.TrimRight(art, "\n"), "\n")
	if len(lines) != len(wordmarkLines) {
		t.Fatalf("expected %d lines, got %d", len(wordmarkLines), len(lines))
	}
}

func TestBootSplash(t *testing.T) {
	s := BootSplash("0.1.0")
	if !strings.Contains(s, BootTagline) {
		t.Fatalf("missing tagline: %q", s)
	}
	if !strings.Contains(s, "v0.1.0") {
		t.Fatal("missing version")
	}
}
