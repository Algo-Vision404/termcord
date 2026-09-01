package config

import "testing"

func TestDefaultConfig(t *testing.T) {
	cfg := Default()
	if cfg.UI.SidebarWidth != 32 {
		t.Fatalf("sidebar width %d", cfg.UI.SidebarWidth)
	}
	if !cfg.UI.NotifyOnMention {
		t.Fatal("expected notify on mention")
	}
}

func TestCachePath(t *testing.T) {
	cfg := Default()
	if cfg.CachePath() == "" {
		t.Fatal("expected cache path")
	}
}
