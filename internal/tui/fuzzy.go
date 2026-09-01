package tui

import (
	"strings"
	"unicode"

	"github.com/termcord/termcord/internal/model"
)

func (a *App) channelLabel(ch model.Channel) string {
	return a.client.DisplayLabel(ch)
}

func scoreMatch(query, target string) int {
	q := strings.ToLower(strings.TrimSpace(query))
	t := strings.ToLower(target)
	if q == "" {
		return 1
	}
	if t == q {
		return 1000
	}
	if strings.HasPrefix(t, q) {
		return 500 + len(q)*10
	}
	if i := strings.Index(t, q); i >= 0 {
		return 200 - i + len(q)*5
	}
	if subsequenceScore(q, t) {
		return 50 + len(q)*2
	}
	return 0
}

func subsequenceScore(query, target string) bool {
	qi := 0
	for _, r := range target {
		if qi >= len(query) {
			break
		}
		if rune(query[qi]) == r {
			qi++
		}
	}
	return qi == len(query)
}

func (a *App) filterChannels(query string, limit int) []model.Channel {
	type item struct {
		ch    model.Channel
		score int
	}
	var scored []item
	for _, ch := range a.channels {
		label := a.channelLabel(ch)
		s := scoreMatch(query, label)
		if s == 0 && query != "" {
			s = scoreMatch(query, a.client.DisplayName(ch))
		}
		if s > 0 {
			scored = append(scored, item{ch: ch, score: s})
		}
	}
	for i := 0; i < len(scored); i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].score > scored[i].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}
	if limit <= 0 {
		limit = 12
	}
	out := make([]model.Channel, 0, limit)
	for i := 0; i < len(scored) && i < limit; i++ {
		out = append(out, scored[i].ch)
	}
	return out
}

func normalizeSpotlightQuery(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "#")
	s = strings.TrimPrefix(s, "@")
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return ' '
		}
		return r
	}, s)
}
