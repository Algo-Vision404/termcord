package ds

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ApplyBorder styles every line with the border palette.
func ApplyBorder(t Theme, raw string) string {
	lines := strings.Split(raw, "\n")
	for i := range lines {
		lines[i] = t.Border.Render(lines[i])
	}
	return strings.Join(lines, "\n")
}

// ApplyChrome3 styles a 3-line box: border / accent inner / border.
func ApplyChrome3(t Theme, raw string, accent lipgloss.Style) string {
	lines := strings.Split(raw, "\n")
	for i := range lines {
		switch i {
		case 1:
			lines[i] = styleInnerRow(lines[i], t.Border, accent)
		default:
			lines[i] = t.Border.Render(lines[i])
		}
	}
	return strings.Join(lines, "\n")
}

// ApplyFooter styles a multi-line footer box (border top, footer body rows, border bottom).
func ApplyFooter(t Theme, raw string) string {
	lines := strings.Split(raw, "\n")
	for i := range lines {
		if i == 0 || i == len(lines)-1 {
			lines[i] = t.Border.Render(lines[i])
		} else {
			lines[i] = t.Footer.Render(lines[i])
		}
	}
	return strings.Join(lines, "\n")
}

// ApplyCordyLane styles Cordy's 3-line lane (border / cordy accent / border).
func ApplyCordyLane(t Theme, raw string) string {
	lines := strings.Split(raw, "\n")
	for i := range lines {
		switch i {
		case 1:
			lines[i] = t.Cordy.Render(lines[i])
		default:
			lines[i] = t.Border.Render(lines[i])
		}
	}
	return strings.Join(lines, "\n")
}

func styleInnerRow(line string, border, inner lipgloss.Style) string {
	content := strings.TrimSuffix(strings.TrimPrefix(line, "│"), "│")
	content = strings.TrimSpace(content)
	return border.Render("│") + " " + inner.Render(PadRight(content, StandardInner-2)) + border.Render("│")
}
