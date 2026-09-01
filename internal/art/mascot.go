package art

import (
	"fmt"
	"strings"

	"github.com/mattn/go-runewidth"
)

// MascotMode is Cordy's current mood — drives animation frames and speech.
type MascotMode int

const (
	MascotIdle MascotMode = iota
	MascotConnecting
	MascotLoading
	MascotWelcome
	MascotEmpty
	MascotWatch
	MascotCompose
	MascotMention
	MascotReact
	MascotSpotlight
	MascotSent
	MascotFarewell
)

const catColWidth = 14

// SceneOpts configures animated scenes and Cordy visibility.
type SceneOpts struct {
	Mode         MascotMode
	ReduceMotion bool
	ShowMascot   bool
	Bubble       string // optional override or channel hint
}

// Cat returns multi-line ASCII art for Cordy.
func Cat(mode MascotMode, frame int, reduceMotion bool) []string {
	frames := cordyFrames[mode]
	if len(frames) == 0 {
		frames = cordyFrames[MascotIdle]
	}
	idx := 0
	if len(frames) > 0 && !reduceMotion {
		idx = frame % len(frames)
	}
	out := make([]string, len(frames[idx]))
	copy(out, frames[idx])
	return out
}

// CatBadge is a compact face for the header strip.
func CatBadge(mode MascotMode, frame int, reduceMotion bool) string {
	faces := badgeFaces[mode]
	if len(faces) == 0 {
		faces = badgeFaces[MascotIdle]
	}
	idx := 0
	if len(faces) > 0 && !reduceMotion {
		idx = frame % len(faces)
	}
	return faces[idx]
}

// MascotBubble returns Cordy's speech for a scene. extra is a channel label for loading, etc.
func MascotBubble(mode MascotMode, extra string) string {
	if extra != "" && mode != MascotLoading {
		return extra
	}
	switch mode {
	case MascotConnecting:
		return "hold on… linking discord"
	case MascotLoading:
		if extra != "" {
			name := extra
			if runewidth.StringWidth(name) > 20 {
				name = runewidth.Truncate(name, 17, "...")
			}
			return "fetching " + name + "…"
		}
		return "digging up messages…"
	case MascotWelcome:
		return "welcome back — ctrl+g to jump"
	case MascotEmpty:
		return "quiet here — say hi?"
	case MascotWatch:
		return "someone's typing…"
	case MascotCompose:
		return "nice draft — enter to send"
	case MascotMention:
		return "psst — you've got @mentions"
	case MascotReact:
		return "pick a reaction 1–5"
	case MascotSpotlight:
		return "where to next?"
	case MascotSent:
		return "sent! purrfect."
	case MascotFarewell:
		return "see you in the terminal~"
	default:
		return "listening…"
	}
}

// MascotPanel lays Cordy beside a text panel.
func MascotPanel(opts SceneOpts, frame int, panel []string) []string {
	if !opts.ShowMascot {
		return panel
	}
	cat := Cat(opts.Mode, frame, opts.ReduceMotion)
	if len(cat) == 0 {
		return panel
	}
	bubble := MascotBubble(opts.Mode, opts.Bubble)
	catBlock := append(cat, speechBubble(bubble, 28)...)
	return joinBeside(catBlock, panel, catColWidth)
}

func speechBubble(text string, width int) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	if width < 10 {
		width = 10
	}
	inner := width - 2
	if runewidth.StringWidth(text) > inner {
		text = runewidth.Truncate(text, inner, "…")
	}
	return []string{
		"╭" + strings.Repeat("─", inner) + "╮",
		"│ " + padRight(text, inner) + "│",
		"╰" + strings.Repeat("─", inner) + "╯",
	}
}

func joinBeside(left, right []string, leftWidth int) []string {
	height := max(len(left), len(right))
	out := make([]string, 0, height)
	for i := 0; i < height; i++ {
		l, r := "", ""
		if i < len(left) {
			l = left[i]
		}
		if i < len(right) {
			r = right[i]
		}
		out = append(out, padRight(l, leftWidth)+"  "+r)
	}
	return out
}

// SidebarCordy decorates the loading sidebar.
func SidebarCordy(frame int, reduceMotion bool) string {
	cat := Cat(MascotConnecting, frame, reduceMotion)
	if len(cat) < 2 {
		return "  sniffing\n  channels…"
	}
	return strings.Join([]string{
		"  " + cat[0],
		"  " + cat[1],
		"  sniffing",
		"  channels…",
	}, "\n")
}

// SpotlightHint is Cordy above the spotlight palette.
func SpotlightHint(frame int, reduceMotion bool) string {
	cat := Cat(MascotSpotlight, frame, reduceMotion)
	if len(cat) == 0 {
		return "  ⌕ fuzzy jump anywhere"
	}
	return strings.Join(cat, "\n") + "\n  " + MascotBubble(MascotSpotlight, "")
}

// ComposePrompt returns an animated compose-row prefix.
func ComposePrompt(frame int, mode MascotMode, reduceMotion bool) string {
	if mode == MascotCompose && !reduceMotion {
		paws := []string{"❯", "≈", "❯", "≈"}
		return paws[frame%len(paws)]
	}
	return AnimInputPromptClassic(frame)
}

func AnimInputPromptClassic(frame int) string {
	if frame%20 < 10 {
		return "❯"
	}
	return "›"
}

func (o SceneOpts) Badge(frame int) string {
	return o.badge(frame)
}

func (o SceneOpts) badge(frame int) string {
	if !o.ShowMascot {
		return "●"
	}
	return CatBadge(o.Mode, frame, o.ReduceMotion)
}

func (o SceneOpts) spin(frame int) string {
	if o.Mode == MascotConnecting || o.Mode == MascotLoading {
		return Spinner(frame)
	}
	return PulseDot(frame)
}

var badgeFaces = map[MascotMode][]string{
	MascotIdle:       {"(^_^)", "(^.^)", "(^_^)"},
	MascotConnecting: {"(o.o)", "(O.O)", "(o.o)"},
	MascotLoading:    {"(-.-)", "(o.o)", "(-.-)"},
	MascotWelcome:    {"(^o^)", "(^o^)", "(^o^)"},
	MascotEmpty:      {"(^.^)", "(u.u)", "(^.^)"},
	MascotWatch:      {"(o.o)", "(O.O)", "(o.o)"},
	MascotCompose:    {"(>.<)", "(^.^)", "(>.<)"},
	MascotMention:    {"(@.@)", "(O.O)", "(@.@)"},
	MascotReact:      {"(^.^)", "(◕◡◕)", "(^.^)"},
	MascotSpotlight:  {"(o.o)", "(O.O)", "(o.o)"},
	MascotSent:       {"(^o^)", "(^o^)", "(^o^)"},
	MascotFarewell:   {"(T.T)", "(^.^)", "(T.T)"},
}

var cordyFrames = map[MascotMode][][]string{
	MascotIdle: {
		{"   /\\_/\\", "  ( ^.^ )", "  (\")_(\")", "   |   |"},
		{"   /\\_/\\", "  ( ^.^ )", "  (u   u)", "   |   |"},
	},
	MascotConnecting: {
		{"   /\\_/\\", "  ( o.o )", "   > ^ <", "  /|   |\\"},
		{"   /\\_/\\", "  ( o.o )", "   > w <", "  / \\ | \\"},
		{"   /\\_/\\", "  ( O.O )", "   > ^ <", "   |   |"},
	},
	MascotLoading: {
		{"   /\\_/\\", "  ( -.- )", "   /\\_/\\", "  d|   |b"},
		{"   /\\_/\\", "  ( o.o )", "   > v <", "  d|   |b"},
		{"   /\\_/\\", "  ( -.- )", "   \\___/", "    | |"},
	},
	MascotWelcome: {
		{"   /\\_/\\", "  ( ^o^ )", "   \\   /", "    hi~"},
		{"   /\\_/\\", "  ( ^o^ )", "  ~(   )~", "  welcome!"},
	},
	MascotEmpty: {
		{"   /\\_/\\", "  ( ^.^ )", "  (\")_(\")", "   zZ zZ"},
		{"   /\\_/\\", "  ( u.u )", "  (\")_(\")", "   |   |"},
	},
	MascotWatch: {
		{"   /\\_/\\", "  ( o.o )·", "   ≥·≤", "   |   |"},
		{"   /\\_/\\", "  ( o.o )··", "   ≥··≤", "   |   |"},
	},
	MascotCompose: {
		{"   /\\_/\\", "  ( ^.^ )", "   ≥⌨≤", "   |   |"},
		{"   /\\_/\\", "  ( >.< )", "   ≥▓≤", "   |   |"},
	},
	MascotMention: {
		{"   /\\_/\\", "  ( O.O ) !", "   > @ <", "   |   |"},
		{"   /\\_/\\", "  ( @.@ ) !!", "   > @ <", "   \\   /"},
	},
	MascotReact: {
		{"   /\\_/\\", "  ( ^.^ )", "   (◕)", "   |   |"},
		{"   /\\_/\\", "  ( ^.^ )", "   👍?", "   |   |"},
	},
	MascotSpotlight: {
		{"   /\\_/\\", "  ( o.o )", "   O---O", "   |   |"},
		{"   /\\_/\\", "  ( O.O )", "   O~--O", "   |   |"},
	},
	MascotSent: {
		{"   /\\_/\\", "  ( ^o^ )", "   ✓ sent", "   |   |"},
	},
	MascotFarewell: {
		{"   /\\_/\\", "  ( T.T )", "   > ~ <", "   bye~"},
		{"   /\\_/\\", "  ( ^.^ )", "   > ~ <", "   |   |"},
	},
}

// InnerWidthForPet returns the standard inner track width.
func InnerWidthForPet() int { return innerWidth }

func sceneGatewayPanel(frame int, step string, pct float64) []string {
	bar := ProgressBar(32, pct, frame)
	spin := Spinner(frame)
	return []string{
		fmt.Sprintf("      │  %s  gateway link                          │", spin),
		fmt.Sprintf("      │     %s                              │", padRight(step, 36)),
		fmt.Sprintf("      │     %s  %3.0f%%                       │", bar, pct*100),
	}
}
