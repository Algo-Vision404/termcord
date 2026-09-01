package tui

import "github.com/termcord/termcord/internal/config"

type Options struct {
	Theme           string
	SidebarWidth    int
	HistoryPageSize int
	NotifyOnMention bool
	ShowEmbeds      bool
	Timestamps      bool
	PluginsEnabled  bool
}

func OptionsFromConfig(cfg config.Config) Options {
	return Options{
		Theme:           cfg.UI.Theme,
		SidebarWidth:    cfg.UI.SidebarWidth,
		HistoryPageSize: cfg.UI.HistoryPageSize,
		NotifyOnMention: cfg.UI.NotifyOnMention,
		ShowEmbeds:      cfg.UI.ShowEmbeds,
		Timestamps:      cfg.UI.Timestamps,
		PluginsEnabled:  cfg.Plugins.Enabled,
	}
}

const helpText = `termcord commands
  /join #name      switch channel
  /threads         list active threads
  /thread name     open a thread
  /parent          leave thread
  /history         load older messages
  /reply           reply to selected message
  /react emoji     toggle reaction
  /search query    search local cache
  /plugins         list loaded plugins
  /clear           clear view
  /quit            exit

keys: ctrl+b sidebar · ctrl+t threads · ctrl+r react
      ctrl+↑/↓ select msg · pgup/pgdn scroll · ctrl+p/n channel`
