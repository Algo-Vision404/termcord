package art

import "testing"

func TestCalmEmpty(t *testing.T) {
	if CalmEmpty() == "" {
		t.Fatal("expected calm empty")
	}
}

func TestRenderEmptyCalm(t *testing.T) {
	out := RenderEmpty(0, SceneOpts{ReduceMotion: true})
	if !contains(out, "No messages") {
		t.Fatalf("got %q", out)
	}
}

func TestFriendlyFooter(t *testing.T) {
	if FriendlyFooter() == "" {
		t.Fatal("expected footer")
	}
}
