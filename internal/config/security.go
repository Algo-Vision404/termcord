package config

import (
	"os"
	"strings"
)

// SecurityWarnings returns non-fatal security notices for the active config.
func SecurityWarnings(cfg Config) []string {
	var out []string
	if strings.TrimSpace(cfg.General.Token) != "" {
		out = append(out, "config.toml contains a plaintext token — prefer `termcord login` (OS keyring)")
	}
	if os.Getenv("TERMCORD_TOKEN") != "" {
		out = append(out, "TERMCORD_TOKEN is set in this shell — child processes and plugins may inherit it unless disabled")
	}
	if cfg.Plugins.Enabled {
		out = append(out, "plugins are enabled — only install .toml plugins you trust (they run shell commands)")
	}
	if !cfg.Cache.Encrypt {
		out = append(out, "cache encryption is off — message bodies are stored in plaintext SQLite")
	}
	return out
}
