package tui

import "testing"

func TestScoreMatch(t *testing.T) {
	if scoreMatch("gen", "Acme › #general") == 0 {
		t.Fatal("expected match for subsequence")
	}
	if scoreMatch("xyz", "Acme › #general") != 0 {
		t.Fatal("expected no match")
	}
}

func TestNormalizeSpotlightQuery(t *testing.T) {
	if got := normalizeSpotlightQuery("  #general "); got != "general" {
		t.Fatalf("got %q", got)
	}
}
