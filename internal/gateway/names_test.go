package gateway

import (
	"testing"

	"github.com/termcord/termcord/internal/model"
)

func TestLooksLikeSnowflake(t *testing.T) {
	if !looksLikeSnowflake("123456789012345678") {
		t.Fatal("expected snowflake")
	}
	if looksLikeSnowflake("general") {
		t.Fatal("channel name is not a snowflake")
	}
}

func TestFallbackChannelName(t *testing.T) {
	name := fallbackChannelName(model.Channel{Kind: model.ChannelDM})
	if name != "Direct Message" {
		t.Fatalf("got %q", name)
	}
}

func TestDisplayNameAvoidsSnowflake(t *testing.T) {
	c := &Client{}
	got := c.DisplayName(model.Channel{
		ID:   "123456789012345678",
		Name: "123456789012345678",
		Kind: model.ChannelDM,
	})
	if got == "123456789012345678" {
		t.Fatalf("display name leaked snowflake: %q", got)
	}
	if got != "Direct Message" {
		t.Fatalf("got %q", got)
	}
}
