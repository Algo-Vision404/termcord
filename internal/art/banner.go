package art

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/termcord/termcord/internal/ds"
)

// LogoCompact is a single-line badge for tight headers.
const LogoCompact = ds.LogoBadge

// SplashConnecting shown while the gateway comes online.
func SplashConnecting(version string) string {
	return strings.Join([]string{
		"",
		"      ╭────────────────────────────────────────╮",
		"      │  ◈  connecting to discord gateway…     │",
		"      │      handshake · identify · ready      │",
		"      ╰────────────────────────────────────────╯",
		"",
		"      ┌─ status ───────────────────────────────┐",
		fmt.Sprintf("      │  version %-32s│", version),
		"      │  waiting for READY event…              │",
		"      └────────────────────────────────────────┘",
		"",
		"      tip: /help once connected · ctrl+c to quit",
		"",
	}, "\n")
}

// SplashEmpty shown when a channel has no messages.
func SplashEmpty() string {
	return strings.Join([]string{
		"",
		"  ╭────────────────────────────────────────────────────╮",
		"  │                                                    │",
		"  │   no messages here yet                             │",
		"  │                                                    │",
		"  │   ┌──────────────┐    ┌─────────────────────────┐  │",
		"  │   │ type below   │    │ /help  command list     │  │",
		"  │   │ to send      │    │ ctrl+b toggle sidebar   │  │",
		"  │   └──────────────┘    └─────────────────────────┘  │",
		"  │                                                    │",
		"  ╰────────────────────────────────────────────────────╯",
		"",
	}, "\n")
}

// SplashLoading shown while history is fetched.
func SplashLoading(channel string) string {
	name := channel
	if name == "" {
		name = "channel"
	}
	if len(name) > 28 {
		name = name[:25] + "..."
	}
	return strings.Join([]string{
		"",
		"  ╭─ loading history ──────────────────────────────────╮",
		fmt.Sprintf("  │  #%s", padRight(name, 49)+"│"),
		"  │  fetching messages from discord…                   │",
		"  ╰────────────────────────────────────────────────────╯",
		"",
	}, "\n")
}

// Farewell shown on quit.
func Farewell() string {
	return strings.Join([]string{
		"",
		"  ╭──────────────────────────────────────╮",
		"  │  thanks for using termcord           │",
		"  │  see you in the terminal ◆           │",
		"  ╰──────────────────────────────────────╯",
		"",
	}, "\n")
}

// HelpBlock is the in-app /help body with ASCII framing.
func HelpBlock() string { return ds.HelpBlock() }

// CLILogo returns the logo for plain CLI output (no colors).
func CLILogo(version string) string { return ds.CLILogo(version) }

// CLIBox wraps lines in a simple box for doctor/init output.
func CLIBox(title string, lines []string) string { return ds.BuildDynamic(title, lines) }

// Divider returns a horizontal rule sized to width.
func Divider(width int) string {
	if width < 10 {
		width = 10
	}
	inner := width - 2
	if inner < 8 {
		inner = 8
	}
	return "╭" + strings.Repeat("─", inner) + "╮"
}

// SectionBox frames a sidebar section title and items.
func SectionBox(title string, width int) string { return ds.SectionTop(title, width) }

// SectionBoxBottom closes a sidebar section.
func SectionBoxBottom(width int) string { return ds.SectionBottom(width) }

// RenderLogo applies theme colors to the compact wordmark.
func RenderLogo(accent, border, dim lipgloss.Style, version string) string {
	_ = border
	mark := accent.Render("  ❯ termcord")
	tag := dim.Render("  discord in your terminal")
	meta := dim.Render(fmt.Sprintf("  v%s · %s", version, ds.Tagline))
	return strings.Join([]string{mark, tag, meta}, "\n")
}

// RenderCompactHeader is a compact TUI header: logo row + status bar.
func RenderCompactHeader(accent, border, status lipgloss.Style, version, statusText string) string {
	row1 := border.Render("╭─") + " " + accent.Render("❯ termcord") + " " +
		dimStyle(border, "v"+version) + " " +
		border.Render(strings.Repeat("─", 12)) + border.Render("╮")
	row2 := border.Render("│") + " " + status.Render(padRight(statusText, 52)) + border.Render("│")
	row3 := border.Render("╰" + strings.Repeat("─", ds.StandardInner) + "╯")
	return strings.Join([]string{row1, row2, row3}, "\n")
}

func dimStyle(base lipgloss.Style, s string) string {
	return base.Copy().Foreground(lipgloss.Color("241")).Render(s)
}
