package gateway

import (
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/termcord/termcord/internal/model"
)

type typingTracker struct {
	mu    sync.Mutex
	users map[string]map[string]time.Time // channelID -> userID -> expiry
	names map[string]string              // userID -> display name
	onUpdate func(channelID string, names []string)
}

func newTypingTracker(onUpdate func(string, []string)) *typingTracker {
	return &typingTracker{
		users: make(map[string]map[string]time.Time),
		names: make(map[string]string),
		onUpdate: onUpdate,
	}
}

func (t *typingTracker) rememberName(id, name string) {
	if id == "" || name == "" {
		return
	}
	t.mu.Lock()
	t.names[id] = name
	t.mu.Unlock()
}

func (t *typingTracker) handle(ev *discordgo.TypingStart) {
	if ev == nil || ev.UserID == "" || ev.ChannelID == "" {
		return
	}
	t.mu.Lock()
	if t.users[ev.ChannelID] == nil {
		t.users[ev.ChannelID] = make(map[string]time.Time)
	}
	t.users[ev.ChannelID][ev.UserID] = time.Now().Add(10 * time.Second)
	channel := ev.ChannelID
	names := t.activeNamesLocked(channel)
	t.mu.Unlock()
	if t.onUpdate != nil {
		t.onUpdate(channel, names)
	}
}

func (t *typingTracker) activeNamesLocked(channelID string) []string {
	now := time.Now()
	byUser := t.users[channelID]
	var names []string
	for id, exp := range byUser {
		if exp.Before(now) {
			delete(byUser, id)
			continue
		}
		name := t.names[id]
		if name == "" {
			name = "someone"
		}
		names = append(names, name)
	}
	return names
}

func (t *typingTracker) Names(channelID string) []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.activeNamesLocked(channelID)
}

func mapAttachments(atts []*discordgo.MessageAttachment) []model.Attachment {
	if len(atts) == 0 {
		return nil
	}
	out := make([]model.Attachment, 0, len(atts))
	for _, a := range atts {
		if a == nil {
			continue
		}
		out = append(out, model.Attachment{
			ID:          a.ID,
			URL:         a.URL,
			Filename:    a.Filename,
			ContentType: a.ContentType,
			Width:       a.Width,
			Height:      a.Height,
			Size:        a.Size,
		})
	}
	return out
}
