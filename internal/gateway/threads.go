package gateway

import (
	"github.com/bwmarrin/discordgo"
	"github.com/termcord/termcord/internal/model"
)

func (c *Client) FetchActiveThreads(parentChannelID string) ([]model.Channel, error) {
	list, err := c.session.ThreadsActive(parentChannelID)
	if err != nil {
		return nil, err
	}

	entries := make([]model.Channel, 0, len(list.Threads))
	for _, ch := range list.Threads {
		if ch == nil {
			continue
		}
		entries = append(entries, mapThreadChannel(ch))
	}

	c.mu.Lock()
	c.threads[parentChannelID] = entries
	c.mu.Unlock()

	return entries, nil
}

func (c *Client) CachedThreads(parentChannelID string) []model.Channel {
	c.mu.RLock()
	defer c.mu.RUnlock()
	chs := c.threads[parentChannelID]
	out := make([]model.Channel, len(chs))
	copy(out, chs)
	return out
}

func (c *Client) ParentChannelID(channelID string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, chs := range c.channels {
		for _, ch := range chs {
			if ch.ID == channelID {
				return ""
			}
		}
	}
	for parentID, threads := range c.threads {
		for _, t := range threads {
			if t.ID == channelID {
				return parentID
			}
		}
	}
	return ""
}

func (c *Client) handleThreadCreate(_ *discordgo.Session, t *discordgo.ThreadCreate) {
	if t.Channel == nil {
		return
	}
	entry := mapThreadChannel(t.Channel)
	parentID := t.ParentID
	if parentID == "" {
		parentID = t.Channel.ParentID
	}

	c.mu.Lock()
	c.threads[parentID] = upsertChannel(c.threads[parentID], entry)
	c.mu.Unlock()
}

func mapThreadChannel(ch *discordgo.Channel) model.Channel {
	return model.Channel{
		ID:       ch.ID,
		Name:     ch.Name,
		GuildID:  ch.GuildID,
		Kind:     model.ChannelThread,
		ParentID: ch.ParentID,
	}
}

func upsertChannel(list []model.Channel, ch model.Channel) []model.Channel {
	for i, existing := range list {
		if existing.ID == ch.ID {
			list[i] = ch
			return list
		}
	}
	return append(list, ch)
}
