package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/termcord/termcord/internal/model"
)

func (a *App) handleTab() (tea.Model, tea.Cmd) {
	val := a.input.Value()
	cursor := a.inputCursor()
	prefix := val[:cursor]
	lastSpace := strings.LastIndexAny(prefix, " \n\t")
	wordStart := 0
	if lastSpace >= 0 {
		wordStart = lastSpace + 1
	}
	word := prefix[wordStart:]
	if word == "" {
		return a, nil
	}

	var matches []string
	var replace func(string) string

	switch {
	case strings.HasPrefix(word, "#"):
		matches = a.matchChannels(strings.TrimPrefix(word, "#"))
		replace = func(name string) string {
			return val[:wordStart] + "#" + name + val[cursor:]
		}
	case strings.HasPrefix(word, "@"):
		matches = a.matchUsers(strings.TrimPrefix(word, "@"))
		replace = func(name string) string {
			return val[:wordStart] + "@" + name + val[cursor:]
		}
	default:
		return a, nil
	}

	if len(matches) == 0 {
		a.status = "no completion for " + word
		return a, nil
	}

	a.tabMatches = matches
	if a.tabIndex >= len(matches) {
		a.tabIndex = 0
	}
	name := matches[a.tabIndex]
	a.tabIndex++
	a.input.SetValue(replace(name))
	a.status = "tab completion · " + name
	return a, nil
}

func (a *App) matchChannels(query string) []string {
	query = strings.ToLower(query)
	var out []string
	seen := map[string]bool{}
	for _, ch := range a.channels {
		if ch.Kind != model.ChannelGuildText && ch.Kind != model.ChannelThread {
			continue
		}
		if seen[ch.Name] {
			continue
		}
		if query == "" || strings.HasPrefix(strings.ToLower(ch.Name), query) || strings.Contains(strings.ToLower(ch.Name), query) {
			out = append(out, ch.Name)
			seen[ch.Name] = true
		}
	}
	return out
}

func (a *App) matchUsers(query string) []string {
	query = strings.ToLower(query)
	seen := map[string]bool{}
	var out []string
	add := func(name string) {
		if name == "" || seen[name] {
			return
		}
		lower := strings.ToLower(name)
		if query == "" || strings.HasPrefix(lower, query) || strings.Contains(lower, query) {
			seen[name] = true
			out = append(out, name)
		}
	}
	for _, ch := range a.channels {
		if ch.Kind == model.ChannelDM || ch.Kind == model.ChannelGroupDM {
			add(ch.Name)
		}
	}
	for _, msg := range a.messages {
		add(msg.Author)
	}
	return out
}

func resetTabCycle(a *App) {
	a.tabIndex = 0
	a.tabMatches = nil
}

func onInputChanged(a *App, ti textinput.Model) {
	if ti.Value() != a.lastInput {
		a.lastInput = ti.Value()
		resetTabCycle(a)
	}
}

func (a *App) inputCursor() int {
	pos := len(a.input.Value())
	if p := a.input.Position(); p >= 0 && p <= pos {
		pos = p
	}
	return pos
}
