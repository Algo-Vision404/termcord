package ds

import (
	"strings"

	"github.com/mattn/go-runewidth"
)

// PadRight pads s to display width n using spaces.
func PadRight(s string, n int) string {
	w := runewidth.StringWidth(s)
	if w >= n {
		return runewidth.Truncate(s, n, "")
	}
	return s + strings.Repeat(" ", n-w)
}

// Top draws the top rule of a box: ╭─ title ───╮
func Top(inner int, title string) string {
	if inner < 8 {
		inner = 8
	}
	label := "─ " + title
	labelW := runewidth.StringWidth(label)
	if labelW > inner {
		label = runewidth.Truncate(label, inner, "")
		labelW = runewidth.StringWidth(label)
	}
	return "╭" + label + strings.Repeat("─", inner-labelW) + "╮"
}

// Row draws a padded box row: │ content │
func Row(inner int, text string) string {
	if inner < 4 {
		inner = 4
	}
	return "│" + PadRight(text, inner) + "│"
}

// Bottom draws the bottom rule: ╰──────╯
func Bottom(inner int) string {
	if inner < 4 {
		inner = 4
	}
	return "╰" + strings.Repeat("─", inner) + "╯"
}

// Build assembles a titled box from body rows (each row is inner-width content).
func Build(inner int, title string, rows []string) string {
	lines := []string{Top(inner, title)}
	for _, r := range rows {
		lines = append(lines, Row(inner, r))
	}
	lines = append(lines, Bottom(inner))
	return strings.Join(lines, "\n")
}

// BuildDynamic sizes the box to fit the title and body lines.
func BuildDynamic(title string, rows []string) string {
	width := runewidth.StringWidth(title) + 4
	for _, line := range rows {
		if w := runewidth.StringWidth(line) + 4; w > width {
			width = w
		}
	}
	if width < MinCLIBox {
		width = MinCLIBox
	}
	inner := width - 2
	out := []string{Top(inner, title)}
	for _, line := range rows {
		out = append(out, Row(inner, " "+line))
	}
	out = append(out, Bottom(inner))
	return strings.Join(out, "\n")
}

// SectionTop is a sidebar section header at variable width.
func SectionTop(title string, width int) string {
	if width < 12 {
		width = 12
	}
	inner := width - 2
	label := title
	if runewidth.StringWidth(label) > inner-2 {
		label = runewidth.Truncate(label, inner-5, "...")
	}
	pad := inner - runewidth.StringWidth(label) - 2
	if pad < 0 {
		pad = 0
	}
	return "╭─ " + label + strings.Repeat("─", pad) + "╮"
}

// SectionBottom closes a sidebar section.
func SectionBottom(width int) string {
	if width < 12 {
		width = 12
	}
	return "╰" + strings.Repeat("─", width-2) + "╯"
}
