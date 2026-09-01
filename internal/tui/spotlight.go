package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/termcord/termcord/internal/art"
	"github.com/termcord/termcord/internal/ds"
)

func (a *App) openSpotlight() {
	a.spotlightOpen = true
	a.spotlightIndex = 0
	a.spotlight.SetValue("")
	a.spotlight.Focus()
	a.input.Blur()
	a.refreshSpotlightResults()
	a.status = "spotlight — type to filter · enter open · esc close"
}

func (a *App) closeSpotlight() {
	a.spotlightOpen = false
	a.spotlight.Blur()
	a.input.Focus()
	a.status = fmt.Sprintf("connected as @%s%s", a.client.Username(), a.activityDigest())
}

func (a *App) refreshSpotlightResults() {
	q := normalizeSpotlightQuery(a.spotlight.Value())
	a.spotlightResults = a.filterChannels(q, 14)
	if a.spotlightIndex >= len(a.spotlightResults) {
		a.spotlightIndex = 0
	}
}

func (a *App) spotlightUp() {
	if len(a.spotlightResults) == 0 {
		return
	}
	a.spotlightIndex--
	if a.spotlightIndex < 0 {
		a.spotlightIndex = len(a.spotlightResults) - 1
	}
}

func (a *App) spotlightDown() {
	if len(a.spotlightResults) == 0 {
		return
	}
	a.spotlightIndex = (a.spotlightIndex + 1) % len(a.spotlightResults)
}

func (a *App) renderSpotlightOverlay() string {
	width := a.width
	if width < 50 {
		width = 50
	}
	if width > 72 {
		width = 72
	}

	var body strings.Builder
	inner := width - 2
	body.WriteString(a.theme.Border.Render(ds.Top(inner, "spotlight — jump anywhere")))
	body.WriteString("\n")
	body.WriteString(a.theme.Border.Render("│ "))
	body.WriteString(a.spotlight.View())
	for lipgloss.Width(a.spotlight.View()) < width-4 {
		body.WriteString(" ")
	}
	body.WriteString(a.theme.Border.Render("│"))
	body.WriteString("\n")
	body.WriteString(a.theme.Dim.Render(strings.Repeat("─", width)))
	body.WriteString("\n")
	if a.opts.ShowMascot {
		body.WriteString(a.theme.Dim.Render(art.SpotlightHint(a.frame, a.opts.ReduceMotion)))
		body.WriteString("\n")
	}

	if len(a.spotlightResults) == 0 {
		body.WriteString(a.theme.Dim.Render("  no matches — try guild name or #channel"))
		body.WriteString("\n")
	} else {
		for i, ch := range a.spotlightResults {
			label := a.channelLabel(ch)
			line := "  " + label
			if ch.Mention {
				line = "@ " + line
			} else if ch.Unread > 0 {
				line = fmt.Sprintf("  (%d) %s", ch.Unread, label)
			}
			if i == a.spotlightIndex {
				line = a.theme.Active.Render("▸ " + strings.TrimSpace(line))
			}
			body.WriteString(line)
			body.WriteString("\n")
		}
	}
	body.WriteString(a.theme.Border.Render(ds.Bottom(inner)))

	box := lipgloss.NewStyle().Width(width).Render(body.String())
	return lipgloss.Place(a.width, a.height-6, lipgloss.Center, lipgloss.Center, box)
}

func (a *App) updateSpotlight(m tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.String() {
	case "esc":
		a.closeSpotlight()
		return a, nil
	case "enter":
		if len(a.spotlightResults) > 0 {
			ch := a.spotlightResults[a.spotlightIndex]
			a.closeSpotlight()
			a.showWelcome = false
			return a, a.selectChannel(ch.ID)
		}
		return a, nil
	case "up", "ctrl+p":
		a.spotlightUp()
		return a, nil
	case "down", "ctrl+n":
		a.spotlightDown()
		return a, nil
	}
	var cmd tea.Cmd
	a.spotlight, cmd = a.spotlight.Update(m)
	a.refreshSpotlightResults()
	return a, cmd
}

func newSpotlightInput(theme Theme) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "guild, #channel, @friend…"
	ti.Prompt = "⌕ "
	ti.PromptStyle = theme.Banner.Copy().Bold(true)
	ti.CharLimit = 80
	ti.Width = 60
	return ti
}
