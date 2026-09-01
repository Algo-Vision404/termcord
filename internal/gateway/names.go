package gateway

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/termcord/termcord/internal/model"
)

func looksLikeSnowflake(s string) bool {
	s = strings.TrimSpace(s)
	if len(s) < 15 || len(s) > 22 {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func (c *Client) rememberUser(id, name string) {
	if id == "" || name == "" {
		return
	}
	c.mu.Lock()
	if c.users == nil {
		c.users = make(map[string]string)
	}
	if c.users[id] != name {
		c.users[id] = name
	}
	c.mu.Unlock()
}

func (c *Client) UserName(userID string) string {
	if userID == "" {
		return ""
	}
	c.mu.RLock()
	name := c.users[userID]
	c.mu.RUnlock()
	if name != "" {
		return name
	}
	u, err := c.session.User(userID)
	if err != nil || u == nil || u.Username == "" {
		return ""
	}
	c.rememberUser(u.ID, u.Username)
	return u.Username
}

func (c *Client) ResolveUserName(userID string) string {
	if name := c.UserName(userID); name != "" {
		return name
	}
	return "user"
}

func (c *Client) ResolveChannelName(channelID string) string {
	if channelID == "" {
		return "channel"
	}
	if ch, ok := c.FindChannel(channelID); ok {
		if label := c.DisplayName(ch); !looksLikeSnowflake(label) {
			return label
		}
	}
	if ch, err := c.session.Channel(channelID); err == nil && ch != nil {
		name := channelDisplayName(ch)
		if looksLikeSnowflake(name) {
			name = c.resolveDMFromRecipients(ch)
		}
		if name != "" && !looksLikeSnowflake(name) {
			c.setDMName(channelID, name)
			return name
		}
	}
	return "channel"
}

func (c *Client) ResolveChannelMention(channelID string) string {
	name := c.ResolveChannelName(channelID)
	if name == "" || name == "channel" {
		return "#channel"
	}
	return "#" + strings.TrimPrefix(name, "#")
}

// DisplayName returns a human-readable label for a channel, never a raw snowflake ID.
func (c *Client) DisplayName(ch model.Channel) string {
	name := strings.TrimSpace(ch.Name)
	if name == "" || looksLikeSnowflake(name) {
		if resolved := c.resolvedDMName(ch.ID); resolved != "" {
			return resolved
		}
		return fallbackChannelName(ch)
	}
	return name
}

func (c *Client) DisplayLabel(ch model.Channel) string {
	switch ch.Kind {
	case model.ChannelDM:
		return "@" + c.DisplayName(ch)
	case model.ChannelGroupDM:
		return "group · " + c.DisplayName(ch)
	default:
		if ch.GuildName != "" {
			return ch.GuildName + " › #" + c.DisplayName(ch)
		}
		return "#" + c.DisplayName(ch)
	}
}

func (c *Client) resolvedDMName(channelID string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, dm := range c.dms {
		if dm.ID == channelID && dm.Name != "" && !looksLikeSnowflake(dm.Name) {
			return dm.Name
		}
	}
	return ""
}

func (c *Client) setDMName(channelID, name string) {
	if channelID == "" || name == "" || looksLikeSnowflake(name) {
		return
	}
	c.mu.Lock()
	changed := false
	for i := range c.dms {
		if c.dms[i].ID == channelID && c.dms[i].Name != name {
			c.dms[i].Name = name
			changed = true
		}
	}
	c.mu.Unlock()
	if changed {
		c.notifyChannelsUpdated()
	}
}

func (c *Client) notifyChannelsUpdated() {
	if c.onChannelsUpdated != nil {
		c.onChannelsUpdated()
	}
}

func (c *Client) SetOnChannelsUpdated(fn func()) {
	c.mu.Lock()
	c.onChannelsUpdated = fn
	c.mu.Unlock()
}

func (c *Client) resolveDMNames() {
	c.mu.RLock()
	dms := append([]model.Channel(nil), c.dms...)
	c.mu.RUnlock()

	for _, dm := range dms {
		if dm.Name != "" && !looksLikeSnowflake(dm.Name) {
			continue
		}
		ch, err := c.session.Channel(dm.ID)
		if err != nil || ch == nil {
			continue
		}
		name := channelDisplayName(ch)
		if looksLikeSnowflake(name) {
			name = c.resolveDMFromRecipients(ch)
		}
		c.setDMName(dm.ID, name)
	}
}

func (c *Client) resolveDMFromRecipients(ch *discordgo.Channel) string {
	if ch == nil {
		return ""
	}
	if len(ch.Recipients) > 0 {
		return channelDisplayName(ch)
	}
	return ""
}

func (c *Client) enrichChannelFromMessage(channelID string, authorID, authorName string) {
	if authorName != "" {
		c.rememberUser(authorID, authorName)
	}
	c.mu.RLock()
	isDM := false
	for _, dm := range c.dms {
		if dm.ID == channelID {
			isDM = true
			if dm.Name != "" && !looksLikeSnowflake(dm.Name) {
				c.mu.RUnlock()
				return
			}
			break
		}
	}
	c.mu.RUnlock()
	if !isDM || authorID == "" || authorID == c.UserID() {
		return
	}
	if name := c.UserName(authorID); name != "" {
		c.setDMName(channelID, name)
	}
}

func (c *Client) resolveTypingUser(userID string) {
	if userID == "" {
		return
	}
	if name := c.UserName(userID); name != "" {
		c.typing.rememberName(userID, name)
	}
}

func fallbackChannelName(ch model.Channel) string {
	switch ch.Kind {
	case model.ChannelDM:
		return "Direct Message"
	case model.ChannelGroupDM:
		return "Group chat"
	default:
		if ch.Name != "" && !looksLikeSnowflake(ch.Name) {
			return ch.Name
		}
		return "channel"
	}
}
