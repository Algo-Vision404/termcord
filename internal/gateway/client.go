package gateway

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/termcord/termcord/internal/model"
)

type Client struct {
	session *discordgo.Session
	userID  string
	username string

	mu       sync.RWMutex
	closing  bool
	guilds   []model.Guild
	channels map[string][]model.Channel
	threads  map[string][]model.Channel
	dms      []model.Channel

	onReady   func()
	onMessage func(model.Message)
	onError   func(error)
	onStatus  func(string)
}

type Options struct {
	Token     string
	Logger    *slog.Logger
	OnReady   func()
	OnMessage func(model.Message)
	OnError   func(error)
	OnStatus  func(string)
}

func New(opts Options) (*Client, error) {
	token := strings.TrimSpace(opts.Token)
	if token == "" {
		return nil, fmt.Errorf("empty discord token")
	}
	if strings.HasPrefix(token, "Bot ") {
		return nil, fmt.Errorf("bot tokens are not supported; termcord requires a user token")
	}

	session, err := discordgo.New(token)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	session.LogLevel = discordgo.LogError

	c := &Client{
		session:  session,
		channels: make(map[string][]model.Channel),
		threads:  make(map[string][]model.Channel),
		onReady:  opts.OnReady,
		onMessage: opts.OnMessage,
		onError:  opts.OnError,
		onStatus: opts.OnStatus,
	}

	session.AddHandler(c.handleReady)
	session.AddHandler(c.handleMessageCreate)
	session.AddHandler(c.handleMessageUpdate)
	session.AddHandler(c.handleThreadCreate)
	session.AddHandler(c.handleReactionAdd)
	session.AddHandler(c.handleReactionRemove)
	session.AddHandler(c.handleDisconnect)
	session.AddHandler(c.handleConnect)

	return c, nil
}

func (c *Client) Open() error {
	return c.session.Open()
}

func (c *Client) Close() error {
	c.mu.Lock()
	c.closing = true
	c.mu.Unlock()
	return c.session.Close()
}

func (c *Client) isClosing() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.closing
}

func (c *Client) handleConnect(_ *discordgo.Session, _ *discordgo.Connect) {
	if c.onStatus != nil && !c.isClosing() {
		c.onStatus("gateway connected")
	}
}

func (c *Client) handleDisconnect(_ *discordgo.Session, _ *discordgo.Disconnect) {
	if c.isClosing() {
		return
	}
	if c.onStatus != nil {
		c.onStatus("disconnected — reconnecting…")
	}
}

func (c *Client) UserID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.userID
}

func (c *Client) Username() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.username
}

func (c *Client) Guilds() []model.Guild {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]model.Guild, len(c.guilds))
	copy(out, c.guilds)
	return out
}

func (c *Client) Channels(guildID string) []model.Channel {
	c.mu.RLock()
	defer c.mu.RUnlock()
	chs := c.channels[guildID]
	out := make([]model.Channel, len(chs))
	copy(out, chs)
	return out
}

func (c *Client) DMs() []model.Channel {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]model.Channel, len(c.dms))
	copy(out, c.dms)
	return out
}

func (c *Client) AllSelectableChannels() []model.Channel {
	return c.FlatChannels()
}

func (c *Client) SendMessage(channelID, content string) (*model.Message, error) {
	msg, err := c.session.ChannelMessageSend(channelID, content)
	if err != nil {
		return nil, err
	}
	out := c.mapMessage(msg)
	return &out, nil
}

func (c *Client) SendReply(channelID, content, replyToID string) (*model.Message, error) {
	msg, err := c.session.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content: content,
		Reference: &discordgo.MessageReference{
			MessageID: replyToID,
			ChannelID: channelID,
		},
	})
	if err != nil {
		return nil, err
	}
	out := c.mapMessage(msg)
	return &out, nil
}

func (c *Client) FetchHistory(channelID string, limit int) ([]model.Message, error) {
	if limit <= 0 {
		limit = 50
	}
	msgs, err := c.session.ChannelMessages(channelID, limit, "", "", "")
	if err != nil {
		return nil, err
	}
	out := make([]model.Message, 0, len(msgs))
	for i := len(msgs) - 1; i >= 0; i-- {
		out = append(out, c.mapMessage(msgs[i]))
	}
	return out, nil
}

func (c *Client) FetchHistoryBefore(channelID, beforeID string, limit int) ([]model.Message, error) {
	if limit <= 0 {
		limit = 50
	}
	msgs, err := c.session.ChannelMessages(channelID, limit, beforeID, "", "")
	if err != nil {
		return nil, err
	}
	out := make([]model.Message, 0, len(msgs))
	for i := len(msgs) - 1; i >= 0; i-- {
		out = append(out, c.mapMessage(msgs[i]))
	}
	return out, nil
}

func (c *Client) handleReady(_ *discordgo.Session, r *discordgo.Ready) {
	c.mu.Lock()
	c.userID = r.User.ID
	c.username = r.User.Username
	c.guilds = c.guilds[:0]
	c.channels = make(map[string][]model.Channel)
	c.dms = c.dms[:0]

	for _, ch := range r.PrivateChannels {
		name := channelDisplayName(ch)
		c.dms = append(c.dms, model.Channel{
			ID:   ch.ID,
			Name: name,
			Kind: dmKind(ch.Type),
		})
	}

	for _, g := range r.Guilds {
		c.guilds = append(c.guilds, model.Guild{ID: g.ID, Name: g.Name})
	}
	c.mu.Unlock()

	go c.loadGuildChannels(r.Guilds)
}

func (c *Client) loadGuildChannels(guilds []*discordgo.Guild) {
	for _, g := range guilds {
		channels, err := c.session.GuildChannels(g.ID)
		if err != nil {
			if c.onError != nil {
				c.onError(fmt.Errorf("guild %s channels: %w", g.Name, err))
			}
			continue
		}
		var entries []model.Channel
		for _, ch := range channels {
			if !isTextLike(ch.Type) {
				continue
			}
			kind := guildChannelKind(ch.Type)
			if kind == model.ChannelThread {
				c.mu.Lock()
				c.threads[ch.ParentID] = upsertChannel(c.threads[ch.ParentID], model.Channel{
					ID:       ch.ID,
					Name:     ch.Name,
					GuildID:  ch.GuildID,
					Kind:     kind,
					ParentID: ch.ParentID,
				})
				c.mu.Unlock()
				continue
			}
			entries = append(entries, model.Channel{
				ID:        ch.ID,
				Name:      ch.Name,
				GuildID:   ch.GuildID,
				GuildName: g.Name,
				Kind:      guildChannelKind(ch.Type),
				ParentID:  ch.ParentID,
			})
		}
		c.mu.Lock()
		c.channels[g.ID] = entries
		c.mu.Unlock()
	}

	if c.onReady != nil {
		c.onReady()
	}
}

func (c *Client) handleMessageCreate(_ *discordgo.Session, m *discordgo.MessageCreate) {
	if c.onMessage == nil {
		return
	}
	c.onMessage(c.mapMessage(m.Message))
}

func (c *Client) handleMessageUpdate(_ *discordgo.Session, m *discordgo.MessageUpdate) {
	if c.onMessage == nil {
		return
	}
	msg := c.mapMessage(m.Message)
	msg.Edited = true
	c.onMessage(msg)
}

func (c *Client) mapMessage(m *discordgo.Message) model.Message {
	author := "unknown"
	authorID := ""
	if m.Author != nil {
		author = m.Author.Username
		authorID = m.Author.ID
		if m.Member != nil && m.Member.Nick != "" {
			author = m.Member.Nick
		}
	}
	replyTo := ""
	replyAuthor := ""
	replyContent := ""
	if m.MessageReference != nil {
		replyTo = m.MessageReference.MessageID
	}
	if m.ReferencedMessage != nil {
		if m.ReferencedMessage.Author != nil {
			replyAuthor = m.ReferencedMessage.Author.Username
		}
		replyContent = m.ReferencedMessage.Content
		if len(m.ReferencedMessage.Embeds) > 0 && replyContent == "" {
			replyContent = "[embed]"
		}
	}
	mentionsMe := m.MentionEveryone
	if !mentionsMe {
		for _, u := range m.Mentions {
			if u != nil && u.ID == c.UserID() {
				mentionsMe = true
				break
			}
		}
	}
	return model.Message{
		ID:              m.ID,
		ChannelID:       m.ChannelID,
		Author:          author,
		AuthorID:        authorID,
		Content:         m.Content,
		Timestamp:       m.Timestamp,
		ReplyToID:       replyTo,
		ReplyToAuthor:   replyAuthor,
		ReplyToContent:  replyContent,
		MentionEveryone: m.MentionEveryone,
		MentionsMe:      mentionsMe,
		Embeds:          mapEmbeds(m.Embeds),
		Reactions:       mapReactions(m.Reactions),
	}
}

func mapEmbeds(embeds []*discordgo.MessageEmbed) []model.Embed {
	if len(embeds) == 0 {
		return nil
	}
	out := make([]model.Embed, 0, len(embeds))
	for _, e := range embeds {
		if e == nil {
			continue
		}
		entry := model.Embed{
			Title:       e.Title,
			Description: e.Description,
			URL:         e.URL,
		}
		if e.Author != nil {
			entry.Author = e.Author.Name
		}
		for _, f := range e.Fields {
			entry.Fields = append(entry.Fields, model.EmbedField{Name: f.Name, Value: f.Value})
		}
		out = append(out, entry)
	}
	return out
}

func isTextLike(t discordgo.ChannelType) bool {
	switch t {
	case discordgo.ChannelTypeGuildText,
		discordgo.ChannelTypeGuildNews,
		discordgo.ChannelTypeGuildPublicThread,
		discordgo.ChannelTypeGuildPrivateThread,
		discordgo.ChannelTypeGuildNewsThread:
		return true
	default:
		return false
	}
}

func guildChannelKind(t discordgo.ChannelType) model.ChannelKind {
	switch t {
	case discordgo.ChannelTypeGuildPublicThread, discordgo.ChannelTypeGuildPrivateThread, discordgo.ChannelTypeGuildNewsThread:
		return model.ChannelThread
	default:
		return model.ChannelGuildText
	}
}

func dmKind(t discordgo.ChannelType) model.ChannelKind {
	if t == discordgo.ChannelTypeGroupDM {
		return model.ChannelGroupDM
	}
	return model.ChannelDM
}

func channelDisplayName(ch *discordgo.Channel) string {
	if ch.Name != "" {
		return ch.Name
	}
	if len(ch.Recipients) == 1 && ch.Recipients[0] != nil {
		return ch.Recipients[0].Username
	}
	if len(ch.Recipients) > 1 {
		return "group-dm"
	}
	return ch.ID
}
