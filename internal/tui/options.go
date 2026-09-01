package tui

import (
	"github.com/termcord/termcord/internal/art"
	"github.com/termcord/termcord/internal/config"
)

type Options struct {
	Theme           string
	SidebarWidth    int
	HistoryPageSize int
	NotifyOnMention bool
	ShowEmbeds      bool
	Timestamps      bool
	ReduceMotion    bool
	ImageProtocol   string
	ShowMascot      bool
	PluginsEnabled  bool
	PluginDir       string
}

// TypingEvent reports who is typing in a channel.
type TypingEvent struct {
	ChannelID string
	Users     []string
}

func OptionsFromConfig(cfg config.Config) Options {
	return Options{
		Theme:           cfg.UI.Theme,
		SidebarWidth:    cfg.UI.SidebarWidth,
		HistoryPageSize: cfg.UI.HistoryPageSize,
		NotifyOnMention: cfg.UI.NotifyOnMention,
		ShowEmbeds:      cfg.UI.ShowEmbeds,
		Timestamps:      cfg.UI.Timestamps,
		ReduceMotion:    cfg.UI.ReduceMotion,
		ImageProtocol:   cfg.UI.ImageProtocol,
		ShowMascot:      cfg.UI.ShowMascot,
		PluginsEnabled:  cfg.Plugins.Enabled,
		PluginDir:       cfg.PluginDir(),
	}
}

func HelpText() string {
	return art.HelpBlock()
}
