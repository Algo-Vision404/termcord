package gateway

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/termcord/termcord/internal/model"
)

const (
	userCapabilities   = 16381
	clientBuildNumber  = 340000
	gatewayOpenTimeout = 25 * time.Second
	maxGatewayAttempts = 3
)

type Client struct {
	session *discordgo.Session
	userWS  *userGateway
	userID  string
	username string

	mu       sync.RWMutex
	closing  bool
	readyCh  chan struct{}
	guilds   []model.Guild
	channels map[string][]model.Channel
	threads  map[string][]model.Channel
	dms      []model.Channel
	users    map[string]string

	onReady   func()
	onMessage func(model.Message)
	onError   func(error)
	onStatus  func(string)
	onTyping  func(string, []string)
	onChannelsUpdated func()

	typing *typingTracker
	authToken string
}

type Options struct {
	Token     string
	Logger    *slog.Logger
	OnReady   func()
	OnMessage func(model.Message)
	OnError   func(error)
	OnStatus  func(string)
	OnTyping  func(channelID string, users []string)
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
	configureUserREST(session)

	c := &Client{
		session:  session,
		readyCh:  make(chan struct{}),
		channels: make(map[string][]model.Channel),
		threads:  make(map[string][]model.Channel),
		users:    make(map[string]string),
		onReady:  opts.OnReady,
		onMessage: opts.OnMessage,
		onError:  opts.OnError,
		onStatus: opts.OnStatus,
		onTyping: opts.OnTyping,
		authToken: token,
	}
	c.typing = newTypingTracker(func(channelID string, users []string) {
		if c.onTyping != nil {
			c.onTyping(channelID, users)
		}
	})

	session.AddHandler(c.handleReady)

	c.userWS = newUserGateway(token, session, c)

	return c, nil
}

func configureUserREST(s *discordgo.Session) {
	s.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) discord/1.0.9204 Chrome/134.0.0.0 Safari/537.36"
}

func (c *Client) Open() error {
	var lastErr error
	for attempt := 0; attempt < maxGatewayAttempts; attempt++ {
		if attempt > 0 {
			if c.onStatus != nil {
				c.onStatus(fmt.Sprintf("retrying gateway (%d/%d)…", attempt+1, maxGatewayAttempts))
			}
			time.Sleep(time.Duration(attempt) * time.Second)
			c.resetGateway()
		} else if c.onStatus != nil {
			c.onStatus("connecting gateway…")
		}
		lastErr = c.openOnce()
		if lastErr == nil {
			if c.onStatus != nil {
				c.onStatus("gateway connected")
			}
			return nil
		}
		if !IsReconnectable(lastErr) {
			return lastErr
		}
	}
	return lastErr
}

func (c *Client) resetGateway() {
	c.userWS.close()
	c.readyCh = make(chan struct{})
	c.userWS = newUserGateway(c.authToken, c.session, c)
}

func (c *Client) openOnce() error {
	if err := c.userWS.open(); err != nil {
		return err
	}
	select {
	case <-c.readyCh:
		return nil
	case err := <-c.userWS.failCh:
		return err
	case <-time.After(gatewayOpenTimeout):
		return fmt.Errorf("couldn't connect within %s — check network or run termcord doctor", gatewayOpenTimeout)
	}
}

func (c *Client) Session() *discordgo.Session {
	return c.session
}

func (c *Client) Close() error {
	c.mu.Lock()
	c.closing = true
	c.mu.Unlock()
	if c.userWS != nil {
		c.userWS.close()
	}
	return nil
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
	select {
	case <-c.readyCh:
	default:
		close(c.readyCh)
	}
	c.mu.Unlock()

	go c.resolveDMNames()
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
	msg := c.mapMessage(m.Message)
	c.enrichChannelFromMessage(msg.ChannelID, msg.AuthorID, msg.Author)
	c.onMessage(msg)
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
		c.rememberUser(authorID, author)
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
	out := model.Message{
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
		Attachments:     mapAttachments(m.Attachments),
		Reactions:       mapReactions(m.Reactions),
	}
	c.noteAuthor(out)
	return out
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

func (c *Client) SendTyping(channelID string) error {
	if channelID == "" {
		return nil
	}
	return c.session.ChannelTyping(channelID)
}

func (c *Client) TypingUsers(channelID string) []string {
	if c.typing == nil {
		return nil
	}
	return c.typing.Names(channelID)
}

func (c *Client) AuthToken() string {
	return c.authToken
}

func (c *Client) noteAuthor(msg model.Message) {
	if c.typing != nil {
		c.typing.rememberName(msg.AuthorID, msg.Author)
	}
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
	if ch == nil {
		return "channel"
	}
	if ch.Name != "" && !looksLikeSnowflake(ch.Name) {
		return ch.Name
	}
	if len(ch.Recipients) == 1 && ch.Recipients[0] != nil {
		return ch.Recipients[0].Username
	}
	if len(ch.Recipients) > 1 {
		names := make([]string, 0, len(ch.Recipients))
		for _, u := range ch.Recipients {
			if u != nil && u.Username != "" {
				names = append(names, u.Username)
			}
		}
		if len(names) > 0 {
			label := strings.Join(names, ", ")
			if len(label) > 32 {
				return label[:29] + "..."
			}
			return label
		}
		return "Group chat"
	}
	return "Direct Message"
}
