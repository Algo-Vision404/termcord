package art

import (
	"fmt"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/termcord/termcord/internal/ds"
)

var (
	spinnerBraille = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	spinnerDots    = []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}
	spinnerPulse   = []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█", "▇", "▆", "▅", "▄", "▃", "▂"}
	waveChars      = []string{"∿", "〜", "≈", "~", "≈", "〜", "∿"}
)

func Spinner(frame int) string {
	return spinnerBraille[frame%len(spinnerBraille)]
}

func SpinnerDots(frame int) string {
	return spinnerDots[frame%len(spinnerDots)]
}

func ProgressBar(width int, pct float64, frame int) string {
	if width < 8 {
		width = 8
	}
	filled := int(pct * float64(width))
	if filled > width {
		filled = width
	}
	if filled < width && pct > 0 {
		edge := spinnerPulse[frame%len(spinnerPulse)]
		var b strings.Builder
		b.WriteString("[")
		if filled > 0 {
			b.WriteString(strings.Repeat("█", filled-1))
		}
		b.WriteString(edge)
		b.WriteString(strings.Repeat("░", width-filled))
		b.WriteString("]")
		return b.String()
	}
	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", width-filled) + "]"
}

func PulseDot(frame int) string {
	dots := []string{"○", "◔", "◑", "◕", "●", "◕", "◑", "◔"}
	return dots[frame%len(dots)]
}

func BlinkCursor(frame int) string {
	if frame%16 < 8 {
		return "_"
	}
	return " "
}

func ScanBeam(frame, width int) string {
	if width < 2 {
		width = 2
	}
	period := (width - 1) * 2
	if period < 1 {
		period = 1
	}
	pos := frame % period
	if pos >= width {
		pos = period - pos
	}
	if pos < 0 {
		pos = 0
	}
	if pos >= width {
		pos = width - 1
	}
	line := make([]rune, width)
	for i := range line {
		line[i] = '─'
	}
	line[pos] = '◆'
	return string(line)
}

func Marquee(text string, width, frame int) string {
	if len(text) <= width {
		return padRight(text, width)
	}
	offset := frame % len(text)
 doubled := text + "   " + text
	if offset >= len(doubled) {
		offset = 0
	}
	chunk := doubled[offset:]
	if len(chunk) < width {
		chunk += doubled
	}
	return chunk[:width]
}

func LogoFrame(frame int) []string {
	cursor := BlinkCursor(frame)
	return []string{
		fmt.Sprintf("  ❯ termcord%s", cursor),
		"  discord in your terminal",
	}
}

func connectingSteps() []string {
	return []string{
		"finding discord servers",
		"opening secure connection",
		"signing you in",
		"syncing your channels",
		"almost ready",
	}
}

// AnimConnecting is the animated startup splash with Cordy.
func AnimConnecting(frame int, version string, opts SceneOpts) string {
	step := connectingSteps()[frame/12%len(connectingSteps())]
	pct := float64(frame%120) / 120.0
	if pct > 0.92 {
		pct = 0.92
	}
	wave := waveChars[frame%len(waveChars)]

	panel := sceneGatewayPanel(frame, step, pct)
	panel = append([]string{"      ╭────────────────────────────────────────╮"}, panel...)
	panel = append(panel, "      ╰────────────────────────────────────────╯")

	scene := MascotPanel(opts, frame, panel)

	lines := []string{
		"",
		strings.Join(LogoFrame(frame), "\n"),
		fmt.Sprintf("  %s %s · v%s", wave, ds.Tagline, version),
		"",
	}
	lines = append(lines, scene...)
	lines = append(lines,
		"",
		"      ╭─ signal ───────────────────────────────╮",
		fmt.Sprintf("      │  %s direct · encrypted · local cache │", SpinnerDots(frame)),
		"      ╰────────────────────────────────────────╯",
		"",
		"      tip: cordy listens below · /help · ctrl+c quit",
		"",
	)
	return strings.Join(lines, "\n")
}

// AnimLoading is the animated channel history loader.
func AnimLoading(frame int, channelLabel string, opts SceneOpts) string {
	name := channelLabel
	if name == "" {
		name = "channel"
	}
	if runewidth.StringWidth(name) > 24 {
		name = runewidth.Truncate(name, 21, "...")
	}
	pct := float64(frame%40) / 40.0
	bar := ProgressBar(32, pct, frame)
	spin := Spinner(frame)

	opts.Bubble = name
	panel := []string{
		"  ╭─ loading messages ─────────────────────────────────╮",
		fmt.Sprintf("  │  %s  %s", spin, padRight(name, 42)),
		fmt.Sprintf("  │  %s                                   │", bar),
		"  │  fetching from discord · saving to local cache     │",
		"  ╰────────────────────────────────────────────────────╯",
		"",
		"  " + Marquee("messages · replies · reactions · embeds", 52, frame),
		"",
	}
	scene := MascotPanel(opts, frame, panel)
	return strings.Join(scene, "\n")
}

// AnimEmpty is the animated empty-channel state.
func AnimEmpty(frame int, opts SceneOpts) string {
	pulse := spinnerPulse[frame%len(spinnerPulse)]
	wave := waveChars[frame%len(waveChars)]
	panel := []string{
		"  ╭────────────────────────────────────────────────────╮",
		"  │                                                    │",
		fmt.Sprintf("  │   %s  quiet channel — be the first to say hi %s     │", PulseDot(frame), wave),
		"  │                                                    │",
		"  │   ┌──────────────┐    ┌─────────────────────────┐  │",
		fmt.Sprintf("  │   │ %s compose   │    │ /help  command palette  │  │", pulse),
		"  │   │ below        │    │ tab @user #channel      │  │",
		"  │   └──────────────┘    └─────────────────────────┘  │",
		"  │                                                    │",
		"  ╰────────────────────────────────────────────────────╯",
		"",
		"  " + Marquee("type below to chat · ctrl+g jump · ctrl+p/n switch", 52, frame),
		"",
	}
	return strings.Join(MascotPanel(opts, frame, panel), "\n")
}

// AnimFarewell is the animated quit screen.
func AnimFarewell(frame int, opts SceneOpts) string {
	panel := []string{
		"  ╭──────────────────────────────────────╮",
		fmt.Sprintf("  │  %s closing session…                │", Spinner(frame)),
		"  │  thanks for using termcord           │",
		"  │  cordy will nap until you return     │",
		"  ╰──────────────────────────────────────╯",
		"",
	}
	if frame > 8 {
		panel = append(panel, "  "+ProgressBar(28, float64(min(frame, 20))/20.0, frame), "")
	}
	if frame > 18 {
		panel = append(panel, "  "+Marquee("offline · cache saved · gateway closed", 40, frame), "")
	}
	opts.Mode = MascotFarewell
	return strings.Join(MascotPanel(opts, frame, panel), "\n")
}

// AnimHeader renders the live TUI header strip.
func AnimHeader(frame int, version, status string, busy bool, opts SceneOpts) string {
	spin := opts.badge(frame)
	if busy && !opts.ShowMascot {
		spin = PulseDot(frame)
	}
	title := fmt.Sprintf("termcord  v%s", version)
	row1 := boxTop(title)
	row2 := boxRow(fmt.Sprintf(" %s  %s", spin, status))
	row3 := boxBottom()
	return strings.Join([]string{row1, row2, row3}, "\n")
}

// AnimChannelBar renders the active channel strip.
func AnimChannelBar(frame int, title string, loading bool, opts SceneOpts) string {
	inner := strings.TrimSpace(title)
	if len(inner) > 44 {
		inner = inner[:41] + "..."
	}
	if loading {
		inner += " " + Spinner(frame)
	}
	if opts.ShowMascot && opts.Mode == MascotWatch {
		inner = CatBadge(MascotWatch, frame, opts.ReduceMotion) + "  " + inner
	}
	return strings.Join([]string{
		boxTop("channel"),
		boxRow(" " + inner),
		boxBottom(),
	}, "\n")
}

// AnimFooter renders the animated keybinding footer.
func AnimFooter(frame int) string {
	tail := cordyTail(frame)
	row := boxRow(" tab @user #channel  ctrl+y scritch  ctrl+g spotlight  ctrl+m @mentions" + tail)
	if pos := 2 + (frame % 40); pos < innerWidth-1 {
		runes := []rune(row)
		if runes[pos] == ' ' {
			runes[pos] = '·'
		}
		row = string(runes)
	}
	return strings.Join([]string{
		boxTop("keys · cordy"),
		row,
		boxRow(" ctrl+b sidebar  ctrl+t threads  ctrl+r react  ctrl+p/n channels"),
		boxRow(" ctrl+↑/↓ msg    pgup/pgdn scroll  ctrl+l bottom  ctrl+c quit"),
		boxBottom(),
	}, "\n")
}

func cordyTail(frame int) string {
	tails := []string{"  ~", " ~", "  ~", "~ "}
	return tails[frame%len(tails)]
}

// AnimInputPrompt returns the compose row prefix animation.
func AnimInputPrompt(frame int, mode MascotMode, reduceMotion bool) string {
	return ComposePrompt(frame, mode, reduceMotion)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
