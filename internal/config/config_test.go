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
	if cfg.Plugins.Enabled {
		t.Fatal("plugins should be disabled by default")
	}
	if !cfg.Cache.Encrypt {
		t.Fatal("cache encryption should be on by default")
	}
}

func TestCachePath(t *testing.T) {
	cfg := Default()
	if cfg.CachePath() == "" {
		t.Fatal("expected cache path")
	}
}
