package config

import "testing"

func TestSecurityWarnings(t *testing.T) {
	cfg := Default()
	cfg.General.Token = "secret"
	cfg.Plugins.Enabled = true
	cfg.Cache.Encrypt = false

	t.Setenv("TERMCORD_TOKEN", "from-env")
	w := SecurityWarnings(cfg)
	if len(w) < 4 {
		t.Fatalf("expected multiple warnings, got %v", w)
	}
}

func TestApplySecurityDefaultsPluginsOff(t *testing.T) {
	raw := `[plugins]
dir = "/tmp/plugins"
`
	cfg := Default()
	cfg.Plugins.Enabled = true
	cfg.applySecurityDefaults(raw)
	if cfg.Plugins.Enabled {
		t.Fatal("plugins should stay off when enabled is not set")
	}
}

func TestApplySecurityDefaultsPluginsExplicit(t *testing.T) {
	raw := `[plugins]
enabled = true
`
	cfg := Default()
	cfg.Plugins.Enabled = true
	cfg.applySecurityDefaults(raw)
	if !cfg.Plugins.Enabled {
		t.Fatal("explicit enabled = true should be preserved")
	}
}
