package gateway

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/termcord/termcord/internal/model"
)

func (c *Client) FetchMessage(channelID, messageID string) (*model.Message, error) {
	msg, err := c.session.ChannelMessage(channelID, messageID)
	if err != nil {
		return nil, err
	}
	out := c.mapMessage(msg)
	return &out, nil
}

func (c *Client) AddReaction(channelID, messageID, emoji string) error {
	emoji = normalizeEmoji(emoji)
	return c.session.MessageReactionAdd(channelID, messageID, emoji)
}

func (c *Client) RemoveReaction(channelID, messageID, emoji string) error {
	emoji = normalizeEmoji(emoji)
	return c.session.MessageReactionRemove(channelID, messageID, emoji, c.UserID())
}

func (c *Client) ToggleReaction(channelID, messageID, emoji string) error {
	msg, err := c.FetchMessage(channelID, messageID)
	if err != nil {
		return err
	}
	emoji = normalizeEmoji(emoji)
	for _, r := range msg.Reactions {
		if r.Emoji == emoji && r.Me {
			return c.RemoveReaction(channelID, messageID, emoji)
		}
	}
	return c.AddReaction(channelID, messageID, emoji)
}

func (c *Client) handleReactionAdd(_ *discordgo.Session, e *discordgo.MessageReactionAdd) {
	if e.MessageReaction == nil || c.onMessage == nil {
		return
	}
	c.refreshMessageReactions(e.ChannelID, e.MessageID)
}

func (c *Client) handleReactionRemove(_ *discordgo.Session, e *discordgo.MessageReactionRemove) {
	if e.MessageReaction == nil || c.onMessage == nil {
		return
	}
	c.refreshMessageReactions(e.ChannelID, e.MessageID)
}

func (c *Client) refreshMessageReactions(channelID, messageID string) {
	msg, err := c.FetchMessage(channelID, messageID)
	if err != nil {
		if c.onError != nil {
			c.onError(fmt.Errorf("refresh reactions: %w", err))
		}
		return
	}
	c.onMessage(*msg)
}

func mapReactions(reactions []*discordgo.MessageReactions) []model.Reaction {
	if len(reactions) == 0 {
		return nil
	}
	out := make([]model.Reaction, 0, len(reactions))
	for _, r := range reactions {
		if r == nil || r.Emoji == nil {
			continue
		}
		out = append(out, model.Reaction{
			Emoji: formatEmoji(r.Emoji),
			Count: r.Count,
			Me:    r.Me,
		})
	}
	return out
}

func formatEmoji(e *discordgo.Emoji) string {
	if e.ID != "" {
		if e.Animated {
			return fmt.Sprintf("<a:%s:%s>", e.Name, e.ID)
		}
		return fmt.Sprintf("<:%s:%s>", e.Name, e.ID)
	}
	return e.Name
}

func normalizeEmoji(emoji string) string {
	emoji = strings.TrimSpace(emoji)
	if strings.HasPrefix(emoji, "<") {
		// Custom emoji mention -> name:id for API
		if i := strings.LastIndex(emoji, ":"); i >= 0 {
			inner := strings.Trim(emoji, "<>")
			parts := strings.Split(inner, ":")
			if len(parts) >= 2 {
				return parts[len(parts)-2] + ":" + parts[len(parts)-1]
			}
		}
	}
	return emoji
}
