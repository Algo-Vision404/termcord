package ds

import (
	"fmt"
	"strings"
)

// Rule draws a full-width horizontal separator.
func Rule(t Theme, inner int) string {
	if inner < 4 {
		inner = 4
	}
	return t.Border.Render(strings.Repeat("─", inner))
}

// RenderAppBar is the merged status + channel context strip (4 lines).
func RenderAppBar(t Theme, inner int, version, status, badge, channel string) string {
	brand := t.Header.Render(fmt.Sprintf("termcord v%s", version))
	stat := strings.TrimSpace(status)
	if badge != "" {
		stat = badge + "  " + stat
	}
	line1 := brand + "  " + t.Status.Render(stat)

	ch := strings.TrimSpace(channel)
	if ch == "" {
		ch = "no channel selected"
	}
	line2 := t.ChannelBar.Render(PadRight(" "+ch, inner))

	sep := Rule(t, inner)
	return strings.Join([]string{
		sep,
		PadRight(line1, inner),
		line2,
		sep,
	}, "\n")
}

// RenderComposeRow draws the input row with a rule above it (2 lines).
func RenderComposeRow(t Theme, inner int, inputLine string) string {
	return Rule(t, inner) + "\n" + PadRight(inputLine, inner)
}

// RenderStatusFooter is a single dim shortcut hint line.
func RenderStatusFooter(t Theme, inner int) string {
	return t.Footer.Render(PadRight(" "+ShortcutHint, inner))
}

// FramePanel wraps content in a titled box (CLI/help only — not used in chat stream).
func FramePanel(t Theme, inner int, title string, content string) string {
	lines := strings.Split(content, "\n")
	out := make([]string, 0, len(lines)+2)
	out = append(out, t.Border.Render(Top(inner, title)))
	for _, line := range lines {
		out = append(out, t.Border.Render("│")+PadRight(line, inner)+t.Border.Render("│"))
	}
	out = append(out, t.Border.Render(Bottom(inner)))
	return strings.Join(out, "\n")
}

// Legacy helpers kept for CLI/tests — prefer RenderAppBar in the TUI.

func RenderAppHeader(t Theme, inner int, version, status, badge string) string {
	return RenderAppBar(t, inner, version, status, badge, "")
}

func RenderChannelStrip(t Theme, inner int, label string) string {
	return RenderAppBar(t, inner, "", "", "", label)
}

func RenderCompose(t Theme, inner int, _title, inputLine string) string {
	return RenderComposeRow(t, inner, inputLine)
}

func RenderShortcutFooter(t Theme, inner int) string {
	return RenderStatusFooter(t, inner)
}

// ShortcutsFooter is the unstyled shortcut box (used by tests and CLI).
func ShortcutsFooter() string {
	return Build(StandardInner, "shortcuts", []string{
		" " + ShortcutHint,
	})
}
