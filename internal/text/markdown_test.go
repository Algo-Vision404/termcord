package text

import (
	"strings"
	"testing"
)

func TestRenderDiscordMentions(t *testing.T) {
	in := "hey <@123456789012345678> check <#999>"
	out := RenderDiscord(in, 0)
	if out != "hey @user check #channel" {
		t.Fatalf("got %q", out)
	}
}

func TestFormatEmbeds(t *testing.T) {
	s := FormatEmbeds([]Embed{{Title: "Build", Description: "passed"}})
	if !strings.Contains(s, "Build") || !strings.Contains(s, "passed") {
		t.Fatalf("unexpected embed output: %q", s)
	}
}
