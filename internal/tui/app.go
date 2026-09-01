package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/termcord/termcord/internal/cache"
	"github.com/termcord/termcord/internal/gateway"
	"github.com/termcord/termcord/internal/model"
	"github.com/termcord/termcord/internal/plugins"
	"github.com/termcord/termcord/internal/text"
	"github.com/termcord/termcord/internal/version"
)

type msgReady struct{}
type msgRefreshChannels struct{}
type msgIncoming struct{ message model.Message }
type msgHistory struct {
	messages []model.Message
	channel  string
	prepend  bool
}
type msgSent struct{ err error }
type msgErr struct{ err error }
type msgThreadsLoaded struct {
	threads []model.Channel
	parent  string
	err     error
}
type msgReactionDone struct{ err error }
type msgStatus struct{ text string }

var quickReactions = []string{"👍", "❤️", "😂", "🔥", "✅"}

type App struct {
	client  *gateway.Client
	store   *cache.Store
	opts    Options
	theme   Theme
	plugins []plugins.Plugin

	width  int
	height int

	sidebarOpen bool
	threadsOpen bool

	channels          []model.Channel
	sections          []model.ChannelSection
	channelIndex      int
	activeChannel     string
	activeChannelName string
	activeParentID    string
	activeGuildName   string

	threads     []model.Channel
	threadIndex int

	messages      []model.Message
	messageCursor int
	reactMode     bool
	replyToID     string

	vp    viewport.Model
	input textinput.Model

	status      string
	ready       bool
	quitting    bool
}

func New(client *gateway.Client, store *cache.Store, opts Options, pluginList []plugins.Plugin) *App {
	ti := textinput.New()
	ti.Placeholder = "Message or /help"
	ti.Focus()
	ti.CharLimit = 2000
	ti.Width = 80

	vp := viewport.New(80, 20)
	vp.SetContent("Connecting to Discord…")

	if opts.SidebarWidth <= 0 {
		opts.SidebarWidth = 32
	}
	if opts.HistoryPageSize <= 0 {
		opts.HistoryPageSize = 50
	}

	return &App{
		client:      client,
		store:       store,
		opts:        opts,
		theme:       LoadTheme(opts.Theme),
		plugins:     pluginList,
		sidebarOpen: true,
		input:       ti,
		vp:          vp,
		status:      "connecting…",
	}
}

func (a *App) Init() tea.Cmd {
	return tea.Batch(waitReady(a.client), textinput.Blink)
}

func waitReady(client *gateway.Client) tea.Cmd {
	return func() tea.Msg {
		for i := 0; i < 300; i++ {
			if len(client.AllSelectableChannels()) > 0 {
				return msgReady{}
			}
			time.Sleep(100 * time.Millisecond)
		}
		return msgReady{}
	}
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = m.Width
		a.height = m.Height
		a.layout()
		return a, nil

	case msgReady:
		a.applyConnected()
		if len(a.channels) > 0 && a.activeChannel == "" {
			a.channelIndex = 0
			return a, a.selectChannel(a.channels[0].ID)
		}
		return a, scheduleSlowRefresh()

	case msgRefreshChannels:
		a.syncChannels()
		return a, scheduleSlowRefresh()

	case msgIncoming:
		_ = a.store.UpsertMessage(context.Background(), m.message)
		if m.message.ChannelID == a.activeChannel {
			a.messages = upsertMessage(a.messages, m.message)
			a.ensureMessageCursor()
			a.renderMessages()
			a.vp.GotoBottom()
		} else {
			a.markUnread(m.message)
		}
		return a, nil

	case msgThreadsLoaded:
		if m.err != nil {
			a.status = "threads: " + m.err.Error()
			return a, nil
		}
		if m.parent == a.threadParentID() {
			a.threads = m.threads
			if len(a.threads) == 0 {
				a.status = "no active threads"
			} else {
				a.status = fmt.Sprintf("%d threads in #%s", len(a.threads), a.parentChannelName())
			}
		}
		return a, nil

	case msgReactionDone:
		a.reactMode = false
		if m.err != nil {
			a.status = "reaction failed: " + m.err.Error()
		} else {
			a.status = "reaction updated"
		}
		return a, nil

	case msgHistory:
		if m.prepend {
			a.messages = append(m.messages, a.messages...)
		} else {
			a.messages = m.messages
		}
		a.ensureMessageCursor()
		a.renderMessages()
		if !m.prepend && m.channel == a.activeChannel {
			a.vp.GotoBottom()
		}
		return a, nil

	case msgSent:
		if m.err != nil {
			a.status = "send failed: " + m.err.Error()
			return a, nil
		}
		a.input.SetValue("")
		a.replyToID = ""
		return a, nil

	case msgErr:
		a.status = m.err.Error()
		return a, nil

	case msgStatus:
		a.status = m.text
		return a, nil

	case tea.KeyMsg:
		if a.reactMode {
			switch m.String() {
			case "esc":
				a.reactMode = false
				a.status = "reaction cancelled"
				return a, nil
			case "1", "2", "3", "4", "5":
				idx := int(m.String()[0] - '1')
				if idx >= 0 && idx < len(quickReactions) {
					return a, a.toggleReaction(quickReactions[idx])
				}
			}
		}

		if a.input.Value() == "" {
			switch m.String() {
			case "pgup", "pgdown", "up", "down":
				var cmd tea.Cmd
				a.vp, cmd = a.vp.Update(msg)
				return a, cmd
			}
		}

		switch {
		case m.Type == tea.KeyCtrlC, m.String() == "ctrl+q":
			a.quitting = true
			return a, tea.Quit
		case m.String() == "ctrl+b":
			a.sidebarOpen = !a.sidebarOpen
			a.layout()
			return a, nil
		case m.String() == "ctrl+t":
			a.threadsOpen = !a.threadsOpen
			if a.threadsOpen {
				return a, a.loadThreads()
			}
			return a, nil
		case m.String() == "ctrl+r":
			if len(a.messages) == 0 {
				a.status = "no message to react to"
				return a, nil
			}
			a.ensureMessageCursor()
			a.reactMode = true
			a.status = "react: 1=👍 2=❤️ 3=😂 4=🔥 5=✅ · esc cancel"
			return a, nil
		case m.String() == "ctrl+up":
			a.moveMessageCursor(-1)
			return a, nil
		case m.String() == "ctrl+down":
			a.moveMessageCursor(1)
			return a, nil
		case m.String() == "ctrl+p":
			return a, a.prevChannel()
		case m.String() == "ctrl+n":
			return a, a.nextChannel()
		case m.String() == "ctrl+l":
			a.vp.GotoBottom()
			return a, nil
		case m.Type == tea.KeyEnter:
			if a.reactMode {
				return a, nil
			}
			line := strings.TrimSpace(a.input.Value())
			if line == "" {
				return a, nil
			}
			if strings.HasPrefix(line, "/") {
				return a, a.runCommand(line)
			}
			return a, a.sendMessage(line)
		}
	}

	var cmd tea.Cmd
	a.input, cmd = a.input.Update(msg)
	return a, cmd
}

func scheduleSlowRefresh() tea.Cmd {
	return tea.Tick(10*time.Second, func(time.Time) tea.Msg {
		return msgRefreshChannels{}
	})
}

func (a *App) applyConnected() {
	a.ready = true
	a.syncChannels()
	a.status = fmt.Sprintf("connected as @%s", a.client.Username())
}

func (a *App) syncChannels() {
	prev := a.activeChannel
	unread := map[string]model.Channel{}
	for _, sec := range a.sections {
		for _, ch := range sec.Channels {
			unread[ch.ID] = ch
		}
	}

	a.sections = a.client.ChannelSections()
	for si := range a.sections {
		for ci := range a.sections[si].Channels {
			id := a.sections[si].Channels[ci].ID
			if old, ok := unread[id]; ok {
				a.sections[si].Channels[ci].Unread = old.Unread
				a.sections[si].Channels[ci].Mention = old.Mention
			}
		}
	}
	a.channels = a.flatFromSections()

	if prev != "" {
		for i, ch := range a.channels {
			if ch.ID == prev {
				a.channelIndex = i
				break
			}
		}
	}
}

func (a *App) flatFromSections() []model.Channel {
	var out []model.Channel
	for _, sec := range a.sections {
		out = append(out, sec.Channels...)
	}
	return out
}

func (a *App) View() string {
	if a.quitting {
		return a.theme.Dim.Render("Goodbye.\n")
	}
	title := a.theme.Header.Render("termcord "+version.Version) + " " + a.theme.Status.Render(a.status)
	bar := a.theme.ChannelBar.Render(a.channelBarTitle())
	main := a.renderMain()
	footer := a.theme.Footer.Render("ctrl+b sidebar · ctrl+t threads · ctrl+r react · /help · ctrl+c quit")
	return lipgloss.JoinVertical(lipgloss.Left, title, bar, main, a.input.View(), footer)
}

func (a *App) channelBarTitle() string {
	if a.activeChannel == "" {
		return "  pick a channel"
	}
	if a.activeParentID != "" {
		return fmt.Sprintf("  ↳ %s", a.activeChannelName)
	}
	if a.activeGuildName != "" {
		return fmt.Sprintf("  %s  ›  #%s", a.activeGuildName, a.activeChannelName)
	}
	if ch, ok := a.client.FindChannel(a.activeChannel); ok && ch.Kind == model.ChannelDM {
		return "  @" + a.activeChannelName
	}
	return "  #" + a.activeChannelName
}

func (a *App) renderMain() string {
	chat := a.vp.View()
	if !a.sidebarOpen {
		return chat
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, a.renderSidebar(), chat)
}

func (a *App) renderSidebar() string {
	var b strings.Builder
	for si, sec := range a.sections {
		if si > 0 {
			b.WriteString("\n")
		}
		title := sec.Title
		if len(title) > a.opts.SidebarWidth-2 {
			title = title[:a.opts.SidebarWidth-5] + "..."
		}
		b.WriteString(a.theme.SidebarTitle.Render(title))
		b.WriteString("\n")
		for _, ch := range sec.Channels {
			label := formatChannelLabel(ch)
			if ch.Mention {
				label = a.theme.Mention.Render("@ " + label)
			} else if ch.Unread > 0 {
				label = a.theme.Unread.Render(fmt.Sprintf("%s (%d)", label, ch.Unread))
			}
			if ch.ID == a.activeChannel || (a.activeParentID != "" && ch.ID == a.activeParentID) {
				label = a.theme.Active.Render("> " + label)
			} else {
				label = "  " + label
			}
			b.WriteString(label)
			b.WriteString("\n")
		}
	}
	if a.threadsOpen {
		b.WriteString("\n")
		b.WriteString(a.theme.SidebarTitle.Render("Threads · " + a.parentChannelName()))
		b.WriteString("\n")
		if len(a.threads) == 0 {
			b.WriteString(a.theme.Dim.Render("  (none)"))
			b.WriteString("\n")
		}
		for _, th := range a.threads {
			label := "↳ " + th.Name
			if th.ID == a.activeChannel {
				label = a.theme.Active.Render("> " + label)
			} else {
				label = "  " + label
			}
			b.WriteString(label)
			b.WriteString("\n")
		}
	}
	return lipgloss.NewStyle().Width(a.opts.SidebarWidth).Render(b.String())
}

func formatChannelLabel(ch model.Channel) string {
	switch ch.Kind {
	case model.ChannelDM:
		return "@" + ch.Name
	default:
		return "#" + ch.Name
	}
}

func (a *App) renderMessages() {
	if len(a.messages) == 0 {
		a.vp.SetContent(a.theme.Dim.Render("No messages yet. Type to send or /help for commands."))
		return
	}
	width := a.vp.Width - 4
	if width < 20 {
		width = 20
	}
	var lines []string
	for i, msg := range a.messages {
		line := a.formatMessageLine(msg, width)
		if i == a.messageCursor {
			line = a.theme.Selected.Render("▸ " + line)
		}
		lines = append(lines, line)
	}
	a.vp.SetContent(strings.Join(lines, "\n"))
}

func (a *App) formatMessageLine(msg model.Message, width int) string {
	var head string
	if a.opts.Timestamps {
		head = fmt.Sprintf("%s ", a.theme.Dim.Render(msg.Timestamp.Local().Format("15:04")))
	}
	author := a.theme.User.Render(msg.Author)
	body := text.RenderDiscord(msg.Content, width)
	body = text.HighlightLinks(body, func(s string) string { return a.theme.Link.Render(s) })
	if a.opts.ShowEmbeds && len(msg.Embeds) > 0 {
		embeds := make([]text.Embed, 0, len(msg.Embeds))
		for _, e := range msg.Embeds {
			fields := make([]text.EmbedField, 0, len(e.Fields))
			for _, f := range e.Fields {
				fields = append(fields, text.EmbedField{Name: f.Name, Value: f.Value})
			}
			embeds = append(embeds, text.Embed{
				Title: e.Title, Description: e.Description, URL: e.URL, Author: e.Author, Fields: fields,
			})
		}
		extra := text.FormatEmbeds(embeds)
		if extra != "" {
			body = strings.TrimSpace(body + "\n" + a.theme.Embed.Render(extra))
		}
	}
	if msg.ReplyToAuthor != "" {
		preview := text.RenderDiscord(msg.ReplyToContent, 60)
		if preview == "" {
			preview = "…"
		}
		body = a.theme.Reply.Render(fmt.Sprintf("↩ %s: %s", msg.ReplyToAuthor, preview)) + "\n" + body
	}
	line := head + author + ": " + body
	if msg.Edited {
		line += a.theme.Dim.Render(" (edited)")
	}
	if msg.MentionsMe {
		line = a.theme.Mention.Render(line)
	}
	if len(msg.Reactions) > 0 {
		line += "\n  " + a.renderReactions(msg.Reactions)
	}
	return line
}

func (a *App) renderReactions(reactions []model.Reaction) string {
	parts := make([]string, 0, len(reactions))
	for _, r := range reactions {
		label := fmt.Sprintf("%s %d", r.Emoji, r.Count)
		if r.Me {
			label = a.theme.ReactionMe.Render(label)
		} else {
			label = a.theme.Reaction.Render(label)
		}
		parts = append(parts, label)
	}
	return strings.Join(parts, " ")
}

func (a *App) layout() {
	sidebar := 0
	if a.sidebarOpen {
		sidebar = a.opts.SidebarWidth + 1
	}
	chatW := a.width - sidebar - 2
	if chatW < 20 {
		chatW = 20
	}
	chatH := a.height - 7
	if chatH < 5 {
		chatH = 5
	}
	a.vp.Width = chatW
	a.vp.Height = chatH
	a.input.Width = a.width - 2
	a.renderMessages()
}

func (a *App) selectChannel(channelID string) tea.Cmd {
	for i, ch := range a.channels {
		if ch.ID == channelID {
			a.channelIndex = i
			a.activeChannel = ch.ID
			a.activeChannelName = ch.Name
			a.activeGuildName = ch.GuildName
			a.activeParentID = ""
			break
		}
	}
	for si := range a.sections {
		for ci := range a.sections[si].Channels {
			if a.sections[si].Channels[ci].ID == channelID {
				a.sections[si].Channels[ci].Unread = 0
				a.sections[si].Channels[ci].Mention = false
			}
		}
	}
	a.channels = a.flatFromSections()
	for _, th := range a.threads {
		if th.ID == channelID {
			a.activeChannel = th.ID
			a.activeChannelName = th.Name
			a.activeParentID = th.ParentID
			break
		}
	}

	a.messages = nil
	a.messageCursor = -1
	a.reactMode = false
	a.replyToID = ""
	a.vp.SetContent(a.theme.Dim.Render("Loading history…"))
	limit := a.opts.HistoryPageSize
	client := a.client
	store := a.store
	return func() tea.Msg {
		ctx := context.Background()
		history, err := client.FetchHistory(channelID, limit)
		if err != nil {
			cached, cacheErr := store.ListMessages(ctx, channelID, limit, "")
			if cacheErr == nil && len(cached) > 0 {
				return msgHistory{messages: cached, channel: channelID}
			}
			return msgErr{err: err}
		}
		for _, msg := range history {
			_ = store.UpsertMessage(ctx, msg)
		}
		return msgHistory{messages: history, channel: channelID}
	}
}

func (a *App) sendMessage(content string) tea.Cmd {
	channelID := a.activeChannel
	replyTo := a.replyToID
	client := a.client
	store := a.store
	return func() tea.Msg {
		var msg *model.Message
		var err error
		if replyTo != "" {
			msg, err = client.SendReply(channelID, content, replyTo)
		} else {
			msg, err = client.SendMessage(channelID, content)
		}
		if err != nil {
			return msgSent{err: err}
		}
		_ = store.UpsertMessage(context.Background(), *msg)
		return msgIncoming{message: *msg}
	}
}

func (a *App) runCommand(input string) tea.Cmd {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil
	}
	name := strings.TrimPrefix(parts[0], "/")
	args := parts[1:]

	switch name {
	case "help":
		a.injectSystem(helpText)
		a.input.SetValue("")
		return nil
	case "quit":
		a.quitting = true
		return tea.Quit
	case "clear":
		a.messages = nil
		a.renderMessages()
		a.input.SetValue("")
		return nil
	case "join":
		if len(args) == 0 {
			a.status = "usage: /join #channel-name"
			return nil
		}
		target := strings.TrimPrefix(args[0], "#")
		for _, ch := range a.channels {
			if strings.EqualFold(ch.Name, target) {
				a.input.SetValue("")
				return a.selectChannel(ch.ID)
			}
		}
		a.status = "channel not found: " + target
		return nil
	case "history":
		a.input.SetValue("")
		return a.loadOlderHistory()
	case "reply":
		if len(a.messages) == 0 {
			a.status = "no message to reply to"
			return nil
		}
		a.ensureMessageCursor()
		a.replyToID = a.messages[a.messageCursor].ID
		a.status = "replying — type message and press enter"
		a.input.SetValue("")
		return nil
	case "search":
		if len(args) == 0 {
			a.status = "usage: /search query"
			return nil
		}
		query := strings.Join(args, " ")
		store := a.store
		a.input.SetValue("")
		return func() tea.Msg {
			results, err := store.Search(context.Background(), query, 20)
			if err != nil {
				return msgErr{err: err}
			}
			return msgHistory{messages: results, channel: a.activeChannel}
		}
	case "threads":
		a.input.SetValue("")
		a.threadsOpen = true
		return a.loadThreads()
	case "thread":
		if len(args) == 0 {
			a.status = "usage: /thread thread-name"
			return nil
		}
		a.input.SetValue("")
		return a.joinThreadByName(strings.Join(args, " "))
	case "parent":
		parent := a.activeParentID
		if parent == "" {
			parent = a.client.ParentChannelID(a.activeChannel)
		}
		if parent == "" {
			a.status = "not inside a thread"
			return nil
		}
		a.input.SetValue("")
		return a.selectChannel(parent)
	case "react":
		if len(args) == 0 {
			a.status = "usage: /react emoji"
			return nil
		}
		a.input.SetValue("")
		return a.toggleReaction(args[0])
	case "plugins":
		a.input.SetValue("")
		if len(a.plugins) == 0 {
			a.injectSystem("no plugins loaded — add .toml files to the plugins directory")
			return nil
		}
		var lines []string
		for _, p := range a.plugins {
			lines = append(lines, fmt.Sprintf("/%s — %s", p.Name, p.Description))
		}
		a.injectSystem(strings.Join(lines, "\n"))
		return nil
	default:
		if a.opts.PluginsEnabled {
			if p, ok := plugins.ByName(a.plugins, name); ok {
				a.input.SetValue("")
				return a.runPlugin(p, strings.Join(args, " "))
			}
		}
		a.status = "unknown command — try /help"
		return nil
	}
}

func (a *App) runPlugin(p plugins.Plugin, args string) tea.Cmd {
	channelID := a.activeChannel
	channelName := a.activeChannelName
	username := a.client.Username()
	return func() tea.Msg {
		out, err := p.Run(plugins.RunContext{
			ChannelID:   channelID,
			ChannelName: channelName,
			Username:    username,
			Message:     args,
			Args:        args,
		})
		if err != nil {
			return msgErr{err: err}
		}
		if out == "" {
			return msgStatus{text: "plugin " + p.Name + " finished"}
		}
		return msgIncoming{message: model.Message{
			Author:    "plugin:" + p.Name,
			Content:   out,
			Timestamp: time.Now(),
			ChannelID: a.activeChannel,
		}}
	}
}

func (a *App) injectSystem(content string) {
	msg := model.Message{
		Author:    "termcord",
		Content:   content,
		Timestamp: time.Now(),
		ChannelID: a.activeChannel,
	}
	a.messages = append(a.messages, msg)
	a.renderMessages()
	a.vp.GotoBottom()
}

func (a *App) loadOlderHistory() tea.Cmd {
	if a.activeChannel == "" {
		return nil
	}
	beforeID := ""
	if len(a.messages) > 0 {
		beforeID = a.messages[0].ID
	}
	limit := a.opts.HistoryPageSize
	channelID := a.activeChannel
	client := a.client
	store := a.store
	return func() tea.Msg {
		ctx := context.Background()
		older, err := client.FetchHistoryBefore(channelID, beforeID, limit)
		if err != nil {
			return msgErr{err: err}
		}
		if len(older) == 0 {
			return msgErr{err: fmt.Errorf("no older messages")}
		}
		for _, msg := range older {
			_ = store.UpsertMessage(ctx, msg)
		}
		return msgHistory{messages: older, channel: channelID, prepend: true}
	}
}

func (a *App) prevChannel() tea.Cmd {
	if len(a.channels) == 0 {
		return nil
	}
	a.channelIndex--
	if a.channelIndex < 0 {
		a.channelIndex = len(a.channels) - 1
	}
	return a.selectChannel(a.channels[a.channelIndex].ID)
}

func (a *App) nextChannel() tea.Cmd {
	if len(a.channels) == 0 {
		return nil
	}
	a.channelIndex = (a.channelIndex + 1) % len(a.channels)
	return a.selectChannel(a.channels[a.channelIndex].ID)
}

func (a *App) markUnread(msg model.Message) {
	for si := range a.sections {
		for ci := range a.sections[si].Channels {
			if a.sections[si].Channels[ci].ID != msg.ChannelID {
				continue
			}
			a.sections[si].Channels[ci].Unread++
			if msg.MentionsMe {
				a.sections[si].Channels[ci].Mention = true
			}
			a.channels = a.flatFromSections()
			if a.opts.NotifyOnMention && msg.MentionsMe {
				a.status = fmt.Sprintf("mention in #%s from %s", a.sections[si].Channels[ci].Name, msg.Author)
				fmt.Print("\a")
			}
			return
		}
	}
}

func (a *App) loadThreads() tea.Cmd {
	parentID := a.threadParentID()
	if parentID == "" {
		a.status = "threads require a guild text channel"
		return nil
	}
	client := a.client
	return func() tea.Msg {
		threads, err := client.FetchActiveThreads(parentID)
		return msgThreadsLoaded{threads: threads, parent: parentID, err: err}
	}
}

func (a *App) joinThreadByName(name string) tea.Cmd {
	name = strings.TrimPrefix(name, "↳ ")
	for _, th := range a.threads {
		if strings.EqualFold(th.Name, name) {
			return a.selectChannel(th.ID)
		}
	}
	a.status = "thread not found: " + name
	return nil
}

func (a *App) toggleReaction(emoji string) tea.Cmd {
	if len(a.messages) == 0 {
		return nil
	}
	a.ensureMessageCursor()
	msg := a.messages[a.messageCursor]
	client := a.client
	channelID := a.activeChannel
	return func() tea.Msg {
		err := client.ToggleReaction(channelID, msg.ID, emoji)
		return msgReactionDone{err: err}
	}
}

func (a *App) threadParentID() string {
	if a.activeParentID != "" {
		return a.activeParentID
	}
	for _, ch := range a.channels {
		if ch.ID == a.activeChannel && ch.Kind == model.ChannelGuildText {
			return ch.ID
		}
	}
	return ""
}

func (a *App) parentChannelName() string {
	parentID := a.threadParentID()
	if parentID == "" {
		return a.activeChannelName
	}
	for _, ch := range a.channels {
		if ch.ID == parentID {
			return ch.Name
		}
	}
	return a.activeChannelName
}

func (a *App) ensureMessageCursor() {
	if len(a.messages) == 0 {
		a.messageCursor = -1
		return
	}
	if a.messageCursor < 0 || a.messageCursor >= len(a.messages) {
		a.messageCursor = len(a.messages) - 1
	}
}

func (a *App) moveMessageCursor(delta int) {
	if len(a.messages) == 0 {
		return
	}
	a.ensureMessageCursor()
	a.messageCursor += delta
	if a.messageCursor < 0 {
		a.messageCursor = 0
	}
	if a.messageCursor >= len(a.messages) {
		a.messageCursor = len(a.messages) - 1
	}
	a.renderMessages()
	a.status = fmt.Sprintf("message %d/%d", a.messageCursor+1, len(a.messages))
}

func upsertMessage(messages []model.Message, incoming model.Message) []model.Message {
	for i, msg := range messages {
		if msg.ID == incoming.ID {
			messages[i] = incoming
			return messages
		}
	}
	return append(messages, incoming)
}

func Run(client *gateway.Client, store *cache.Store, opts Options, pluginList []plugins.Plugin, incoming <-chan model.Message, status <-chan string) error {
	app := New(client, store, opts, pluginList)
	p := tea.NewProgram(app, tea.WithAltScreen())
	go func() {
		for msg := range incoming {
			p.Send(msgIncoming{message: msg})
		}
	}()
	go func() {
		for s := range status {
			p.Send(msgStatus{text: s})
		}
	}()
	_, err := p.Run()
	return err
}
