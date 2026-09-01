package termimg

import (
	"testing"

	"github.com/termcord/termcord/internal/model"
)

func TestDetectKitty(t *testing.T) {
	t.Setenv("TERM", "xterm-kitty")
	t.Setenv("KITTY_WINDOW_ID", "1")
	if DetectTerminal() != ProtocolKitty {
		t.Fatal("expected kitty")
	}
}

func TestResolveNone(t *testing.T) {
	if Resolve("none") != ProtocolNone {
		t.Fatal("expected none")
	}
}

func TestFormatAttachmentsNonImage(t *testing.T) {
	out := FormatAttachments([]model.Attachment{
		{Filename: "doc.pdf", ContentType: "application/pdf"},
	}, ProtocolNone, "", nil)
	if out == "" || !contains(out, "doc.pdf") {
		t.Fatalf("got %q", out)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
