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
	"github.com/termcord/termcord/internal/art"
	"github.com/termcord/termcord/internal/cache"
	"github.com/termcord/termcord/internal/ds"
	"github.com/termcord/termcord/internal/gateway"
	"github.com/termcord/termcord/internal/model"
	"github.com/termcord/termcord/internal/plugins"
	"github.com/termcord/termcord/internal/termimg"
	"github.com/termcord/termcord/internal/text"
	"github.com/termcord/termcord/internal/ux"
	"github.com/termcord/termcord/internal/version"
)

type msgReady struct{}
type msgRefreshChannels struct{}
type msgIncoming struct{ message model.Message }
type msgHistory struct {
	messages []model.Message
	channel  string
	prepend  bool
	cached   bool
}
type msgSearchDone struct {
	query   string
	results []model.Message
	err     error
}
type msgSelectChannel struct{ channelID string }
type msgSent struct{ err error }
type msgErr struct{ err error }
type msgThreadsLoaded struct {
	threads []model.Channel
	parent  string
	err     error
}
type msgReactionDone struct{ err error }
type msgStatus struct{ text string }
type msgTypingUpdate struct {
	channelID string
	users     []string
}
type msgAnimTick struct{}

const animIntervalBusy = 150 * time.Millisecond
const quitAnimFrames = 8

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

	status         string
	ready          bool
	quitting       bool
	frame          int
	loadingHistory bool

	spotlightOpen    bool
	spotlight        textinput.Model
	spotlightIndex   int
	spotlightResults []model.Channel

	focusMode   bool
	showWelcome bool
	mentionHop  int

	typingUsers    []string
	imageCache     map[string]string
	imageProto     termimg.Protocol
	lastTypingSent time.Time
	tabIndex       int
	tabMatches     []string
	lastInput      string
	mascotFlash    int
	cordy          art.CordyPet
	cordyLaneRow   int
}

func New(client *gateway.Client, store *cache.Store, opts Options, pluginList []plugins.Plugin) *App {
	ti := textinput.New()
	ti.Placeholder = "Message…  /help for commands"
	ti.Prompt = "❯ "
	theme, themeNote := LoadTheme(opts.Theme)
	ti.PromptStyle = theme.Banner.Copy().Bold(true)
	ti.CharLimit = 2000
	ti.Width = 40
	ti.Focus()

	vp := viewport.New(80, 20)

	if opts.SidebarWidth <= 0 {
		opts.SidebarWidth = 32
	}
	if opts.HistoryPageSize <= 0 {
		opts.HistoryPageSize = 50
	}

	app := &App{
		client:      client,
		store:       store,
		opts:        opts,
		theme:       theme,
		plugins:     pluginList,
		sidebarOpen: false,
		input:       ti,
		vp:          vp,
		spotlight:   newSpotlightInput(theme),
		status:      "Connecting to Discord…",
		imageCache:  make(map[string]string),
		imageProto:  termimg.Resolve(opts.ImageProtocol),
	}
	if themeNote != "" {
		app.status = themeNote
	}
	app.updateInputPlaceholder()
	app.initCordyLane()
	return app
}

func tickAnim() tea.Cmd {
	return tea.Tick(animIntervalBusy, func(time.Time) tea.Msg { return msgAnimTick{} })
}

func (a *App) needsMotionTick() bool {
	if a.quitting || !a.ready || a.loadingHistory {
		return true
	}
	if a.opts.ShowMascot && !a.opts.ReduceMotion {
		return true
	}
	if !a.opts.ReduceMotion && a.needsDecorRefresh() {
		return true
	}
	return false
}

func (a *App) needsDecorRefresh() bool {
	if a.quitting || !a.ready || a.loadingHistory {
		return true
	}
	if a.showWelcome && a.activeChannel != "" && len(a.messages) == 0 {
		return true
	}
	if len(a.messages) == 0 && a.activeChannel != "" {
		return !a.opts.ReduceMotion
	}
	return false
}

func (a *App) Init() tea.Cmd {
	return tea.Batch(waitReady(a.client), textinput.Blink, tickAnim())
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
		if len(a.channels) == 0 {
			a.status = "Connected, but no channels yet — try termcord doctor"
		}
		a.refreshDecor()
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
		a.notifyCordyActivity()
		if m.message.AuthorID == a.client.UserID() {
			a.mascotFlash = 24
		}
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
			a.status = ux.Friendly(m.err)
			return a, nil
		}
		if m.parent == a.threadParentID() {
			a.threads = m.threads
			if len(a.threads) == 0 {
				a.status = "no active threads in this channel"
			} else {
				a.status = fmt.Sprintf("%d threads in #%s", len(a.threads), a.parentChannelName())
			}
		}
		return a, nil

	case msgReactionDone:
		a.reactMode = false
		if m.err != nil {
			a.status = ux.Friendly(m.err)
		} else {
			a.status = "reaction updated"
		}
		return a, nil

	case msgHistory:
		a.loadingHistory = false
		if m.cached {
			a.status = ux.CachedHistoryNotice("")
		}
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
			a.status = ux.FriendlySend(m.err)
			return a, nil
		}
		a.input.SetValue("")
		a.replyToID = ""
		a.updateInputPlaceholder()
		return a, nil

	case msgErr:
		a.status = ux.Friendly(m.err)
		return a, nil

	case msgSearchDone:
		if m.err != nil {
			a.status = ux.Friendly(m.err)
			return a, nil
		}
		if len(m.results) == 0 {
			a.status = fmt.Sprintf("no matches for \"%s\" in local cache", m.query)
			return a, nil
		}
		a.injectSystem(fmt.Sprintf("search results for \"%s\" (%d) — /clear to return to live view", m.query, len(m.results)))
		a.messages = m.results
		a.ensureMessageCursor()
		a.renderMessages()
		a.vp.GotoBottom()
		a.status = fmt.Sprintf("showing %d search results", len(m.results))
		return a, nil

	case msgSelectChannel:
		a.input.SetValue("")
		return a, a.selectChannel(m.channelID)

	case msgStatus:
		a.status = m.text
		return a, nil

	case msgTypingUpdate:
		if m.channelID == a.activeChannel {
			a.typingUsers = m.users
		}
		return a, nil

	case msgAnimTick:
		a.frame++
		if a.mascotFlash > 0 {
			a.mascotFlash--
		}
		a.tickCordy()
		if a.quitting {
			if a.frame >= quitAnimFrames {
				return a, tea.Quit
			}
			return a, tickAnim()
		}
		if a.needsDecorRefresh() {
			a.refreshDecor()
		}
		a.input.Prompt = "❯ "
		if a.needsMotionTick() {
			return a, tickAnim()
		}
		return a, nil

	case tea.MouseMsg:
		return a.handleCordyMouse(m)

	case tea.KeyMsg:
		a.notifyCordyActivity()
		if a.spotlightOpen {
			return a.updateSpotlight(m)
		}
		if a.showWelcome && m.Type != tea.KeyCtrlC {
			a.showWelcome = false
			a.refreshDecor()
		}
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
				if a.activeChannel == "" && len(a.channels) > 0 {
					switch m.String() {
					case "up":
						return a, a.prevChannel()
					case "down":
						return a, a.nextChannel()
					}
				}
				var cmd tea.Cmd
				a.vp, cmd = a.vp.Update(msg)
				return a, cmd
			}
		}

		switch {
		case m.Type == tea.KeyCtrlC:
			if a.quitting {
				return a, tea.Quit
			}
			a.quitting = true
			a.frame = 0
			a.status = "Closing… (Ctrl+C again to quit now)"
			return a, tickAnim()
		case m.String() == "ctrl+q":
			a.quitting = true
			a.frame = 0
			a.status = "closing…"
			return a, tickAnim()
		case m.String() == "esc":
			if a.replyToID != "" {
				a.replyToID = ""
				a.updateInputPlaceholder()
				a.status = "reply cancelled"
				return a, nil
			}
		case m.String() == "ctrl+g":
			a.openSpotlight()
			return a, textinput.Blink
		case m.String() == "ctrl+f":
			a.focusMode = !a.focusMode
			if a.focusMode {
				a.sidebarOpen = false
				a.status = "Focus mode — sidebar hidden (Ctrl+F to restore chrome)"
			} else {
				a.status = fmt.Sprintf("Connected as @%s%s", a.client.Username(), a.activityDigest())
			}
			a.layout()
			return a, nil
		case m.String() == "ctrl+m":
			return a, a.hopMention()
		case m.String() == "ctrl+b":
			if a.focusMode {
				a.status = "exit focus mode first (ctrl+f)"
				return a, nil
			}
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
			a.status = "React: press 1–5 for 👍 ❤️ 😂 🔥 ✅ · Esc to cancel"
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
		case m.String() == "ctrl+y":
			a.scritchCordy()
			return a, nil
		case m.String() == "tab":
			if !a.reactMode {
				return a.handleTab()
			}
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
	prev := a.input.Value()
	a.input, cmd = a.input.Update(msg)
	onInputChanged(a, a.input)
	if a.input.Value() != prev && strings.TrimSpace(a.input.Value()) != "" {
		if typingCmd := a.sendTypingIfNeeded(); typingCmd != nil {
			return a, tea.Batch(cmd, typingCmd)
		}
	}
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
	a.showWelcome = true
	a.status = fmt.Sprintf("Connected as @%s%s", a.client.Username(), a.activityDigest())
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
		if ch, ok := a.client.FindChannel(prev); ok {
			a.activeChannelName = a.client.DisplayName(ch)
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
	opts := a.sceneOpts()
	if a.quitting {
		return a.theme.Dim.Render(art.RenderFarewell(a.frame, opts))
	}
	bar := a.renderAppBar()
	if a.spotlightOpen {
		parts := []string{bar}
		if a.showCordy() {
			if cordy := a.renderCordyLane(); cordy != "" {
				parts = append(parts, cordy)
			}
		}
		parts = append(parts, a.renderSpotlightOverlay())
		return lipgloss.JoinVertical(lipgloss.Left, parts...)
	}
	cordy := a.renderCordyLane()
	main := a.renderMain()
	inner := a.chromeInner()
	input := ds.RenderComposeRow(a.theme, inner, a.input.View())
	parts := []string{bar}
	if a.showCordy() {
		if cordy == "" {
			cordy = a.renderCordyLane()
		}
		if cordy != "" {
			parts = append(parts, cordy)
		}
	}
	parts = append(parts, main, input)
	if !a.focusMode {
		parts = append(parts, a.renderFooter())
	}
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (a *App) renderAppBar() string {
	status := a.status
	if status == "" {
		status = fmt.Sprintf("Connected as @%s%s", a.client.Username(), a.activityDigest())
	}
	if a.focusMode && !strings.Contains(strings.ToLower(status), "focus") {
		status += " · focus mode"
	}
	badge := ""
	if !a.opts.ReduceMotion {
		busy := !a.ready || a.loadingHistory
		if busy {
			badge = art.PulseDot(a.frame)
		} else if a.opts.ShowMascot {
			badge = a.sceneOpts().Badge(a.frame)
		}
	}
	label := a.channelBarTitle()
	if a.loadingHistory && !a.opts.ReduceMotion {
		label += " " + art.Spinner(a.frame)
	}
	return ds.RenderAppBar(a.theme, a.chromeInner(), version.Version, status, badge, label)
}

func (a *App) renderHeader() string {
	return a.renderAppBar()
}

func (a *App) renderChannelBar() string {
	return a.renderAppBar()
}

func (a *App) renderFooter() string {
	return ds.RenderStatusFooter(a.theme, a.chromeInner())
}

func (a *App) chromeInner() int {
	return ds.PanelInner(a.width)
}

func (a *App) chatInner() int {
	return ds.ChatInner(a.width, a.opts.SidebarWidth, a.sidebarOpen)
}

func (a *App) shouldAnimate() bool {
	return a.needsMotionTick()
}

func (a *App) loadingChannelLabel() string {
	name := a.activeChannelName
	if name == "" {
		return "channel"
	}
	if ch, ok := a.client.FindChannel(a.activeChannel); ok {
		if ch.Kind == model.ChannelDM || ch.Kind == model.ChannelGroupDM {
			return "@" + name
		}
	}
	return "#" + name
}

func (a *App) updateInputPlaceholder() {
	if a.replyToID != "" {
		a.input.Placeholder = "Reply…  Esc to cancel"
		return
	}
	a.input.Placeholder = "Message…  /help for commands"
}

func (a *App) composeTitle() string {
	if a.spotlightOpen {
		return "jump"
	}
	if a.replyToID != "" {
		return "reply"
	}
	return "message"
}

func (a *App) sceneOpts() art.SceneOpts {
	return art.SceneOpts{
		Mode:         a.mascotMode(),
		ReduceMotion: a.opts.ReduceMotion,
		ShowMascot:   a.opts.ShowMascot,
		Bubble:       a.loadingChannelLabel(),
	}
}

func (a *App) mascotMode() art.MascotMode {
	if a.quitting {
		return art.MascotFarewell
	}
	if !a.ready {
		return art.MascotConnecting
	}
	if a.loadingHistory {
		return art.MascotLoading
	}
	if a.spotlightOpen {
		return art.MascotSpotlight
	}
	if a.reactMode {
		return art.MascotReact
	}
	if a.mascotFlash > 0 {
		return art.MascotSent
	}
	if len(a.typingUsers) > 0 {
		return art.MascotWatch
	}
	if strings.TrimSpace(a.input.Value()) != "" {
		return art.MascotCompose
	}
	for _, ch := range a.channels {
		if ch.Mention {
			return art.MascotMention
		}
	}
	if a.showWelcome && a.activeChannel != "" && len(a.messages) == 0 {
		return art.MascotWelcome
	}
	if len(a.messages) == 0 && a.activeChannel != "" {
		return art.MascotEmpty
	}
	return art.MascotIdle
}

func (a *App) activityDigest() string {
	unread, mentions := 0, 0
	for _, ch := range a.channels {
		unread += ch.Unread
		if ch.Mention {
			mentions++
		}
	}
	if unread == 0 && mentions == 0 {
		return ""
	}
	var parts []string
	if unread > 0 {
		parts = append(parts, fmt.Sprintf("%d unread", unread))
	}
	if mentions > 0 {
		parts = append(parts, fmt.Sprintf("%d @", mentions))
	}
	return " · " + strings.Join(parts, " · ")
}

func (a *App) countGuilds() int {
	n := 0
	for _, sec := range a.sections {
		if sec.GuildID != "" {
			n++
		}
	}
	return n
}

func (a *App) hopMention() tea.Cmd {
	var targets []model.Channel
	for _, ch := range a.channels {
		if ch.Mention {
			targets = append(targets, ch)
		}
	}
	if len(targets) == 0 {
		a.status = "no @mentions right now"
		return nil
	}
	a.mentionHop = (a.mentionHop + 1) % len(targets)
	ch := targets[a.mentionHop]
	a.status = fmt.Sprintf("mention hop → %s", a.client.DisplayLabel(ch))
	return a.selectChannel(ch.ID)
}

func (a *App) refreshDecor() {
	if a.quitting {
		return
	}
	opts := a.sceneOpts()
	frame := a.frame
	if a.opts.ReduceMotion {
		frame = 0
	}
	if !a.ready {
		a.vp.SetContent(a.theme.Dim.Render(art.RenderConnecting(frame, version.Version, opts)))
		return
	}
	if a.loadingHistory {
		opts.Bubble = a.loadingChannelLabel()
		a.vp.SetContent(a.theme.Dim.Render(art.RenderLoading(frame, a.loadingChannelLabel(), opts)))
		return
	}
	if a.showWelcome && a.activeChannel != "" && len(a.messages) == 0 {
		a.vp.SetContent(a.theme.Dim.Render(art.RenderWelcome(
			a.client.Username(),
			len(a.channels),
			a.countGuilds(),
			opts,
		)))
		return
	}
	if len(a.messages) == 0 && a.activeChannel != "" {
		a.vp.SetContent(a.theme.Dim.Render(art.RenderEmpty(frame, opts)))
		return
	}
	if len(a.messages) > 0 {
		a.renderMessages()
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (a *App) channelBarTitle() string {
	base := a.channelBarBase()
	if len(a.typingUsers) == 0 {
		return base
	}
	return base + " · " + a.theme.Dim.Render(formatTyping(a.typingUsers))
}

func (a *App) channelBarBase() string {
	if a.activeChannel == "" {
		return "Select a channel"
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

func formatTyping(names []string) string {
	switch len(names) {
	case 0:
		return ""
	case 1:
		return names[0] + " is typing…"
	case 2:
		return names[0] + " and " + names[1] + " are typing…"
	default:
		return fmt.Sprintf("%s and %d others are typing…", names[0], len(names)-1)
	}
}

func (a *App) sendTypingIfNeeded() tea.Cmd {
	if a.activeChannel == "" {
		return nil
	}
	if time.Since(a.lastTypingSent) < 8*time.Second {
		return nil
	}
	a.lastTypingSent = time.Now()
	channelID := a.activeChannel
	client := a.client
	return func() tea.Msg {
		_ = client.SendTyping(channelID)
		return nil
	}
}

func (a *App) renderMain() string {
	chat := a.vp.View()
	if !a.sidebarOpen {
		return chat
	}
	sep := a.theme.Border.Render("│")
	return lipgloss.JoinHorizontal(lipgloss.Top, a.renderSidebar(), sep, chat)
}

func (a *App) renderSidebar() string {
	if len(a.sections) == 0 {
		return lipgloss.NewStyle().Width(a.opts.SidebarWidth).Render(a.theme.Dim.Render("Loading channels…"))
	}
	var b strings.Builder
	for si, sec := range a.sections {
		if si > 0 {
			b.WriteString("\n")
		}
		title := sec.Title
		if len(title) > a.opts.SidebarWidth-2 {
			title = title[:a.opts.SidebarWidth-5] + "..."
		}
		b.WriteString(a.theme.SidebarTitle.Render(" " + title))
		b.WriteString("\n")
		for _, ch := range sec.Channels {
			label := a.client.DisplayLabel(ch)
			if ch.Mention {
				label = a.theme.Mention.Render("@ " + label)
			} else if ch.Unread > 0 {
				label = a.theme.Unread.Render(fmt.Sprintf("%s (%d)", label, ch.Unread))
			}
			if ch.ID == a.activeChannel || (a.activeParentID != "" && ch.ID == a.activeParentID) {
				marker := ">"
				if a.shouldAnimate() && !a.opts.ReduceMotion {
					marker = activeMarker(a.frame)
				}
				label = a.theme.Active.Render(marker + " " + label)
			} else {
				label = "  " + label
			}
			b.WriteString(label)
			b.WriteString("\n")
		}
	}
	if a.threadsOpen {
		b.WriteString("\n")
		b.WriteString(a.theme.SidebarTitle.Render(" Threads · " + a.parentChannelName()))
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

func activeMarker(frame int) string {
	markers := []string{">", "▸", "›", "▸"}
	return markers[frame%len(markers)]
}

func (a *App) renderMessages() {
	if len(a.messages) == 0 {
		opts := a.sceneOpts()
		frame := a.frame
		if a.opts.ReduceMotion {
			frame = 0
		}
		a.vp.SetContent(a.theme.Dim.Render(art.RenderEmpty(frame, opts)))
		return
	}
	width := a.vp.Width
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
		if i < len(a.messages)-1 {
			lines = append(lines, "")
		}
	}
	a.vp.SetContent(strings.Join(lines, "\n"))
}

func (a *App) formatMessageLine(msg model.Message, width int) string {
	var head string
	if a.opts.Timestamps {
		head = fmt.Sprintf("%s ", a.theme.Dim.Render(msg.Timestamp.Local().Format("15:04")))
	}
	author := a.theme.User.Render(msg.Author)
	body := text.RenderDiscordWithResolvers(msg.Content, width, text.MentionResolvers{
		User:    a.client.ResolveUserName,
		Channel: a.client.ResolveChannelMention,
	})
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
	if len(msg.Attachments) > 0 {
		att := termimg.FormatAttachments(msg.Attachments, a.imageProto, a.client.AuthToken(), a.imageCache)
		if att != "" {
			line += "\n" + att
		}
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
	chatInner := a.chatInner()
	chatH := a.height - a.chromeHeight()
	if chatH < 5 {
		chatH = 5
	}
	a.vp.Width = chatInner
	a.vp.Height = chatH

	promptW := lipgloss.Width(a.input.Prompt)
	fieldW := a.chromeInner() - promptW - 2
	if fieldW < 12 {
		fieldW = 12
	}
	a.input.Width = fieldW
	a.spotlight.Width = min(a.width-8, 64)
	a.cordyLaneRow = 6
	a.renderMessages()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (a *App) chromeHeight() int {
	// app bar(4) + compose(2) + footer(1)
	h := 7
	if a.focusMode {
		h = 6
	}
	return h + a.cordyChromeRows()
}

func (a *App) selectChannel(channelID string) tea.Cmd {
	for i, ch := range a.channels {
		if ch.ID == channelID {
			a.channelIndex = i
			a.activeChannel = ch.ID
			a.activeChannelName = a.client.DisplayName(ch)
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
			a.activeChannelName = a.client.DisplayName(th)
			a.activeParentID = th.ParentID
			break
		}
	}

	a.loadingHistory = true
	a.messages = nil
	a.messageCursor = -1
	a.reactMode = false
	a.replyToID = ""
	a.typingUsers = a.client.TypingUsers(channelID)
	a.refreshDecor()
	limit := a.opts.HistoryPageSize
	client := a.client
	store := a.store
	return func() tea.Msg {
		ctx := context.Background()
		history, err := client.FetchHistory(channelID, limit)
		if err != nil {
			cached, cacheErr := store.ListMessages(ctx, channelID, limit, "")
			if cacheErr == nil && len(cached) > 0 {
				return msgHistory{messages: cached, channel: channelID, cached: true}
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
		a.injectSystem(HelpText())
		a.input.SetValue("")
		return nil
	case "cordy", "pet":
		a.scritchCordy()
		a.input.SetValue("")
		return nil
	case "quit":
		a.quitting = true
		a.frame = 0
		return tickAnim()
	case "clear":
		a.messages = nil
		a.renderMessages()
		a.input.SetValue("")
		a.status = "screen cleared — cached history kept"
		return nil
	case "join":
		if len(args) == 0 {
			a.status = "usage: /join #channel-name"
			return nil
		}
		target := strings.TrimPrefix(args[0], "#")
		for _, ch := range a.channels {
			if strings.EqualFold(a.client.DisplayName(ch), target) || strings.EqualFold(ch.Name, target) {
				a.input.SetValue("")
				return a.selectChannel(ch.ID)
			}
		}
		a.status = fmt.Sprintf("channel not found: %s — try /join #exact-name or ctrl+p/n", target)
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
		a.updateInputPlaceholder()
		a.status = fmt.Sprintf("replying to %s — type below, esc to cancel", a.messages[a.messageCursor].Author)
		a.input.SetValue("")
		return nil
	case "search":
		if len(args) == 0 {
			a.status = "usage: /search your query"
			return nil
		}
		query := strings.Join(args, " ")
		store := a.store
		a.input.SetValue("")
		return func() tea.Msg {
			results, err := store.Search(context.Background(), query, 20)
			if err != nil {
				return msgSearchDone{query: query, err: err}
			}
			return msgSearchDone{query: query, results: results}
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
		return a.openThreadByName(strings.Join(args, " "))
	case "parent":
		parent := a.activeParentID
		if parent == "" {
			parent = a.client.ParentChannelID(a.activeChannel)
		}
		if parent == "" {
			a.status = "not in a thread — open a server channel first"
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
			a.injectSystem(ux.PluginDirHint(a.opts.PluginDir))
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
		a.status = fmt.Sprintf("unknown command: /%s — type /help", name)
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
			return msgStatus{text: "plugin " + p.Name + " finished (no output)"}
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
			return msgStatus{text: ux.Friendly(fmt.Errorf("no older messages"))}
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
		a.status = "threads only work in server text channels — pick a #channel first"
		return nil
	}
	client := a.client
	return func() tea.Msg {
		threads, err := client.FetchActiveThreads(parentID)
		return msgThreadsLoaded{threads: threads, parent: parentID, err: err}
	}
}

func (a *App) openThreadByName(name string) tea.Cmd {
	parentID := a.threadParentID()
	if parentID == "" {
		a.status = "open a server #channel first, then /thread name"
		return nil
	}
	name = strings.TrimPrefix(name, "↳ ")
	client := a.client
	return func() tea.Msg {
		threads, err := client.FetchActiveThreads(parentID)
		if err != nil {
			return msgErr{err: err}
		}
		for _, th := range threads {
			if strings.EqualFold(th.Name, name) {
				return msgSelectChannel{channelID: th.ID}
			}
		}
		return msgErr{err: fmt.Errorf("thread not found: %s — try /threads to list active threads", name)}
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

func Run(client *gateway.Client, store *cache.Store, opts Options, pluginList []plugins.Plugin, incoming <-chan model.Message, status <-chan string, typing <-chan TypingEvent) error {
	app := New(client, store, opts, pluginList)
	p := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseCellMotion())
	client.SetOnChannelsUpdated(func() {
		p.Send(msgRefreshChannels{})
	})
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
	go func() {
		for ev := range typing {
			p.Send(msgTypingUpdate{channelID: ev.ChannelID, users: ev.Users})
		}
	}()
	_, err := p.Run()
	return err
}
