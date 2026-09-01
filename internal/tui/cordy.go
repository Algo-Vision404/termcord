package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/termcord/termcord/internal/art"
	"github.com/termcord/termcord/internal/ds"
)

const cordyLaneRows = 3

func (a *App) initCordyLane() {
	a.cordy = art.NewCordyPet()
	a.cordyLaneRow = 6
}

func (a *App) tickCordy() {
	if !a.showCordy() || a.quitting {
		return
	}
	a.cordy.Tick(a.cordyTrackWidth(), a.cordyHints())
}

func (a *App) showCordy() bool {
	return a.opts.ShowMascot
}

func (a *App) cordyHints() art.PetHints {
	return art.PetHints{
		Ready:        a.ready,
		Compose:      strings.TrimSpace(a.input.Value()) != "",
		Mention:      a.hasMentions(),
		Typing:       len(a.typingUsers) > 0,
		JustSent:     a.mascotFlash > 20,
		ReduceMotion: a.opts.ReduceMotion,
	}
}

func (a *App) hasMentions() bool {
	for _, ch := range a.channels {
		if ch.Mention {
			return true
		}
	}
	return false
}

func (a *App) cordyTrackWidth() int {
	w := a.laneWidth()
	if w < 24 {
		w = 24
	}
	return w
}

func (a *App) laneWidth() int {
	w := a.width - 2
	if w <= 0 {
		w = art.LaneBoxWidth(80)
	}
	return art.LaneBoxWidth(w)
}

func (a *App) notifyCordyActivity() {
	if !a.showCordy() {
		return
	}
	a.cordy.NotifyActivity()
}

func (a *App) scritchCordy() {
	if !a.showCordy() {
		a.status = "cordy is hidden — set show_mascot = true in config, then restart"
		return
	}
	if !a.ready {
		a.status = "cordy is still waking up…"
		return
	}
	a.cordy.Scritch()
	if a.cordy.Quip != "" {
		a.status = a.cordy.Quip
	} else {
		a.status = "cordy purrs happily ♡"
	}
}

func (a *App) cordyMouseHit(x, y int) bool {
	if !a.showCordy() || !a.ready || a.spotlightOpen {
		return false
	}
	if y != a.cordyLaneRow+1 {
		return false
	}
	sprite := a.cordy.Sprite(a.opts.ReduceMotion)
	sw := lipgloss.Width(sprite)
	if sw < 5 {
		sw = 7
	}
	left := 2 + a.cordy.X
	right := left + sw + 1
	return x >= left && x <= right
}

func (a *App) handleCordyMouse(m tea.MouseMsg) (tea.Model, tea.Cmd) {
	if m.Type != tea.MouseLeft {
		return a, nil
	}
	if a.cordyMouseHit(m.X, m.Y) {
		a.scritchCordy()
	}
	return a, nil
}

func (a *App) cordyMini() string {
	if !a.showCordy() {
		return ""
	}
	frame := a.frame
	if a.opts.ReduceMotion {
		frame = 0
	}
	return art.MiniCordy(a.mascotMode(), frame, a.opts.ReduceMotion)
}

func (a *App) renderCordyLane() string {
	if !a.showCordy() || a.quitting {
		return ""
	}

	termW := a.width
	if termW <= 0 {
		termW = 80
	}

	var lane string
	title := "cordy"
	if !a.ready {
		lane = padLane("(o.o)~  …connecting…", a.cordyTrackWidth())
		title = "cordy · waking up"
	} else {
		lane = a.cordy.RenderTrack(a.cordyTrackWidth(), a.opts.ReduceMotion)
		title = "cordy · ctrl+y scritch · click the cat"
		if a.cordy.State == art.PetSleep {
			title = "cordy · zzZ napping · ctrl+y to wake"
		}
	}

	raw := art.RenderCordyLaneBox(title, lane, termW)
	return ds.ApplyCordyLane(a.theme, raw)
}

func padLane(s string, w int) string {
	if len(s) >= w {
		return s[:w]
	}
	return s + strings.Repeat(" ", w-len(s))
}

func (a *App) cordyChromeRows() int {
	if !a.showCordy() || a.quitting {
		return 0
	}
	return cordyLaneRows
}

func (a *App) cordyStatusSuffix() string {
	if !a.showCordy() {
		return ""
	}
	mini := a.cordyMini()
	if mini == "" {
		return ""
	}
	return fmt.Sprintf(" · %s cordy", mini)
}
