package ds

// Brand and layout tokens shared by TUI and CLI.
const (
	BrandName  = "TERMCORD"
	Tagline    = "terminal-native discord · direct gateway · local cache"
	LogoBadge  = "termcord"
	// ShortcutHint is the single footer hint in the TUI (see /help for full list).
	ShortcutHint = "Ctrl+G jump · Enter send · Ctrl+P/N channels · /help · Ctrl+B sidebar · Ctrl+C quit"
	ComposeBox = 60 // legacy default; TUI uses PanelInner(terminalWidth)

	StandardInner = 58 // fallback when terminal size is unknown
	MinInner      = 40
	MaxInner      = 78
	MinCLIBox     = 44
)

// ClampInner returns a safe inner width for terminal chrome.
func ClampInner(terminalWidth int) int {
	inner := terminalWidth - 2
	if inner < MinInner {
		inner = MinInner
	}
	if inner > MaxInner {
		inner = MaxInner
	}
	return inner
}
