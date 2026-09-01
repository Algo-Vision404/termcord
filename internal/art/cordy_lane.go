package art

import (
	"strings"

	"github.com/termcord/termcord/internal/ds"
)

// LaneBoxWidth returns a sane inner width for Cordy's lane.
func LaneBoxWidth(terminalWidth int) int { return ds.ClampInner(terminalWidth) }

// RenderCordyLaneBox draws Cordy's full-width walking strip.
func RenderCordyLaneBox(title, lane string, terminalWidth int) string {
	inner := LaneBoxWidth(terminalWidth)
	lane = padRight(lane, inner)
	return strings.Join([]string{
		ds.Top(inner, title),
		ds.Row(inner, " "+strings.TrimRight(lane, " ")),
		ds.Bottom(inner),
	}, "\n")
}

// MiniCordy returns a tiny always-visible sprite for the header.
func MiniCordy(mode MascotMode, frame int, reduceMotion bool) string {
	p := CordyPet{Frame: frame, Dir: 1, State: mapPetToState(mode)}
	return p.Sprite(reduceMotion)
}

func mapPetToState(m MascotMode) PetState {
	switch m {
	case MascotLoading, MascotConnecting:
		return PetWalk
	case MascotWatch, MascotCompose, MascotMention, MascotReact, MascotSpotlight, MascotEmpty:
		return PetSit
	case MascotSent, MascotWelcome:
		return PetPurr
	case MascotFarewell:
		return PetSleep
	default:
		return PetSit
	}
}
