package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/termcord/termcord/internal/security"
)

type Config struct {
	General General `toml:"general"`
	Cache   Cache   `toml:"cache"`
	UI      UI      `toml:"ui"`
	Plugins Plugins `toml:"plugins"`
}

type General struct {
	Token string `toml:"token"`
}

type Cache struct {
	Path    string `toml:"path"`
	Encrypt bool   `toml:"encrypt"`
}

type UI struct {
	Theme           string `toml:"theme"`
	SidebarWidth    int    `toml:"sidebar_width"`
	HistoryPageSize int    `toml:"history_page_size"`
	NotifyOnMention bool   `toml:"notify_on_mention"`
	ShowEmbeds      bool   `toml:"show_embeds"`
	Timestamps      bool   `toml:"timestamps"`
	ReduceMotion    bool   `toml:"reduce_motion"`
	ImageProtocol   string `toml:"image_protocol"`
	ShowMascot      bool   `toml:"show_mascot"`
}

type Plugins struct {
	Enabled bool   `toml:"enabled"`
	Dir     string `toml:"dir"`
}

func Default() Config {
	return Config{
		UI: UI{
			Theme:           "default",
			SidebarWidth:    32,
			HistoryPageSize: 50,
			NotifyOnMention: true,
			ShowEmbeds:      true,
			Timestamps:      true,
			ReduceMotion:    true,
			ShowMascot:      false,
		},
		Cache: Cache{
			Encrypt: true,
		},
		Plugins: Plugins{
			Enabled: false,
		},
	}
}

func DefaultConfigPath() string {
	if v := os.Getenv("TERMCORD_CONFIG"); v != "" {
		return v
	}
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "termcord", "config.toml")
	default:
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			return filepath.Join(xdg, "termcord", "config.toml")
		}
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".config", "termcord", "config.toml")
	}
}

func DefaultCachePath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("LOCALAPPDATA"), "termcord", "messages.db")
	default:
		if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
			return filepath.Join(xdg, "termcord", "messages.db")
		}
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".local", "share", "termcord", "messages.db")
	}
}

func Load(path string) (Config, error) {
	cfg := Default()
	if path == "" {
		path = DefaultConfigPath()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("read config: %w", err)
	}
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config: %w", err)
	}
	cfg.applyMascotDefault(string(data))
	cfg.applySecurityDefaults(string(data))
	cfg.normalize()
	return cfg, nil
}

func (c *Config) applyMascotDefault(raw string) {
	if !strings.Contains(raw, "show_mascot") {
		c.UI.ShowMascot = false
	}
	if !strings.Contains(raw, "reduce_motion") {
		c.UI.ReduceMotion = true
	}
}

func (c *Config) applySecurityDefaults(raw string) {
	if !strings.Contains(raw, "[plugins]") {
		c.Plugins.Enabled = false
		return
	}
	if strings.Contains(raw, "enabled") {
		return
	}
	c.Plugins.Enabled = false
}

func (c *Config) normalize() {
	if c.UI.SidebarWidth <= 0 {
		c.UI.SidebarWidth = 32
	}
	if c.UI.HistoryPageSize <= 0 {
		c.UI.HistoryPageSize = 50
	}
	if c.UI.Theme == "" {
		c.UI.Theme = "default"
	}
}

func (c *Config) CachePath() string {
	if c.Cache.Path != "" {
		return c.Cache.Path
	}
	return DefaultCachePath()
}

func (c *Config) PluginDir() string {
	if c.Plugins.Dir != "" {
		return c.Plugins.Dir
	}
	return pluginsDefaultDir()
}

func pluginsDefaultDir() string {
	return defaultPluginDir()
}

func defaultPluginDir() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "termcord", "plugins")
	default:
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".config", "termcord", "plugins")
	}
}

func Init(path string) (string, bool, error) {
	if path == "" {
		path = DefaultConfigPath()
	}
	if _, err := os.Stat(path); err == nil {
		return path, false, nil
	}
	if err := security.EnsureDir(filepath.Dir(path)); err != nil {
		return "", false, err
	}
	cfg := Default()
	data, err := toml.Marshal(cfg)
	if err != nil {
		return "", false, err
	}
	example := "# termcord config — see config.example.toml in the repo for all options\n\n"
	if err := os.WriteFile(path, append([]byte(example), data...), security.FilePerm); err != nil {
		return "", false, err
	}
	_ = security.SecureFile(path)
	cfgLoaded, _ := Load(path)
	_ = security.EnsureDir(cfgLoaded.PluginDir())
	return path, true, nil
}
