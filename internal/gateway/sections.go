package gateway

import (
	"github.com/termcord/termcord/internal/model"
)

func (c *Client) ChannelSections() []model.ChannelSection {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var sections []model.ChannelSection
	if len(c.dms) > 0 {
		dms := make([]model.Channel, len(c.dms))
		copy(dms, c.dms)
		sections = append(sections, model.ChannelSection{Title: "Direct Messages", Channels: dms})
	}
	for _, g := range c.guilds {
		chs := c.channels[g.ID]
		if len(chs) == 0 {
			continue
		}
		copyChs := make([]model.Channel, len(chs))
		copy(copyChs, chs)
		for i := range copyChs {
			copyChs[i].GuildName = g.Name
		}
		sections = append(sections, model.ChannelSection{
			Title:    g.Name,
			GuildID:  g.ID,
			Channels: copyChs,
		})
	}
	return sections
}

func (c *Client) FlatChannels() []model.Channel {
	sections := c.ChannelSections()
	var out []model.Channel
	for _, sec := range sections {
		out = append(out, sec.Channels...)
	}
	return out
}

func (c *Client) GuildName(guildID string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, g := range c.guilds {
		if g.ID == guildID {
			return g.Name
		}
	}
	return ""
}

func (c *Client) FindChannel(channelID string) (model.Channel, bool) {
	for _, ch := range c.FlatChannels() {
		if ch.ID == channelID {
			return ch, true
		}
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, threads := range c.threads {
		for _, th := range threads {
			if th.ID == channelID {
				return th, true
			}
		}
	}
	return model.Channel{}, false
}
