package art

import (
	"math/rand"
	"time"

	"github.com/mattn/go-runewidth"
)

// PetState is Cordy's ambient on-screen behavior.
type PetState int

const (
	PetWalk PetState = iota
	PetSit
	PetSleep
	PetScratch
	PetPurr
	PetWake
	PetTrot
)

// PetHints drives contextual movement and mood.
type PetHints struct {
	Ready         bool
	Compose       bool
	Mention       bool
	Typing        bool
	JustSent      bool
	ReduceMotion  bool
}

// CordyPet is the small walking mascot on the track lane.
type CordyPet struct {
	X            int
	Dir          int
	State        PetState
	StateTicks   int
	Quip         string
	QuipTicks    int
	TargetX      int
	LastActivity time.Time
	Frame        int
	rng          *rand.Rand
}

const petSleepAfter = 70 * time.Second

func NewCordyPet() CordyPet {
	return CordyPet{
		X:            6,
		Dir:          1,
		State:        PetSit,
		LastActivity: time.Now(),
		rng:          rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// NotifyActivity wakes Cordy from sleep and refreshes idle timer.
func (p *CordyPet) NotifyActivity() {
	now := time.Now()
	if p.State == PetSleep {
		p.State = PetWake
		p.StateTicks = 0
		p.Quip = pickWakeQuip(p.rng)
		p.QuipTicks = 50
	}
	p.LastActivity = now
}

// Scritch pets Cordy — scratch animation then purr.
func (p *CordyPet) Scritch() {
	p.NotifyActivity()
	p.State = PetScratch
	p.StateTicks = 0
	p.Quip = pickScritchQuip(p.rng)
	p.QuipTicks = 90
}

// Tick advances movement and animation. trackW is inner lane width in cells.
func (p *CordyPet) Tick(trackW int, hints PetHints) {
	if trackW < 16 {
		trackW = 16
	}
	p.Frame++
	if p.QuipTicks > 0 {
		p.QuipTicks--
		if p.QuipTicks == 0 {
			p.Quip = ""
		}
	}

	if !hints.Ready {
		p.State = PetSit
		return
	}

	if p.State == PetScratch || p.State == PetPurr || p.State == PetWake {
		p.StateTicks++
		p.advanceSpecialState()
		return
	}

	if hints.JustSent {
		p.State = PetPurr
		p.StateTicks = 0
		p.Quip = pickSentQuip(p.rng)
		p.QuipTicks = 60
		p.TargetX = trackW - runewidth.StringWidth(p.Sprite(hints.ReduceMotion)) - 4
	}

	now := time.Now()
	idle := now.Sub(p.LastActivity)
	if idle > petSleepAfter && p.State != PetSleep {
		p.State = PetSleep
		p.StateTicks = 0
		p.Quip = "zzZ… cordy naps while you read"
		p.QuipTicks = 40
	}

	switch {
	case hints.Mention && p.State != PetSleep:
		p.State = PetTrot
		p.TargetX = 3
		p.Quip = "psst — @mentions!"
		p.QuipTicks = 35
	case hints.Typing && p.State != PetSleep:
		p.State = PetSit
		if p.QuipTicks == 0 {
			p.Quip = "…someone's typing"
			p.QuipTicks = 25
		}
	case hints.Compose:
		p.TargetX = trackW - runewidth.StringWidth(p.Sprite(hints.ReduceMotion)) - 3
		if p.State == PetSleep {
			p.NotifyActivity()
		} else if p.State != PetTrot {
			p.State = PetTrot
		}
	}

	p.move(trackW, hints.ReduceMotion)
}

func (p *CordyPet) advanceSpecialState() {
	switch p.State {
	case PetScratch:
		if p.StateTicks >= 18 {
			p.State = PetPurr
			p.StateTicks = 0
			if p.QuipTicks < 40 {
				p.Quip = pickPurrQuip(p.rng)
				p.QuipTicks = 70
			}
		}
	case PetPurr:
		if p.StateTicks >= 45 {
			p.State = PetSit
			p.StateTicks = 0
		}
	case PetWake:
		if p.StateTicks >= 20 {
			p.State = PetWalk
			p.StateTicks = 0
		}
	}
}

func (p *CordyPet) move(trackW int, reduceMotion bool) {
	spriteW := runewidth.StringWidth(p.Sprite(reduceMotion))
	maxX := trackW - spriteW - 1
	if maxX < 1 {
		maxX = 1
	}

	stepEvery := 3
	if p.State == PetTrot {
		stepEvery = 2
	}
	if p.State == PetSleep || p.State == PetSit {
		stepEvery = 0
	}

	if p.State == PetWalk || p.State == PetTrot {
		if p.Frame%40 == 0 && p.State == PetWalk && p.rng.Intn(3) == 0 {
			p.Dir = -p.Dir
		}
		if p.TargetX > 0 {
			if p.X < p.TargetX {
				p.Dir = 1
			} else if p.X > p.TargetX {
				p.Dir = -1
			} else {
				p.State = PetSit
				p.TargetX = 0
			}
		}
		if stepEvery > 0 && !reduceMotion && p.Frame%stepEvery == 0 {
			p.X += p.Dir
		}
	}

	if p.X <= 1 {
		p.X = 1
		p.Dir = 1
	}
	if p.X >= maxX {
		p.X = maxX
		p.Dir = -1
	}
}

// Sprite returns the current compact Cordy glyph sequence.
func (p *CordyPet) Sprite(reduceMotion bool) string {
	idx := 0
	if !reduceMotion {
		idx = p.Frame
	}
	switch p.State {
	case PetSleep:
		return sleepSprites[idx%len(sleepSprites)]
	case PetScratch:
		return scratchSprites[idx%len(scratchSprites)]
	case PetPurr:
		return purrSprites[idx%len(purrSprites)]
	case PetWake:
		return wakeSprites[idx%len(wakeSprites)]
	case PetTrot:
		if p.Dir > 0 {
			return trotRight[idx%len(trotRight)]
		}
		return trotLeft[idx%len(trotLeft)]
	case PetWalk:
		if p.Dir > 0 {
			return walkRight[idx%len(walkRight)]
		}
		return walkLeft[idx%len(walkLeft)]
	default:
		return sitSprites[idx%len(sitSprites)]
	}
}

// RenderTrack draws the dotted lane with Cordy placed at p.X.
func (p *CordyPet) RenderTrack(innerWidth int, reduceMotion bool) string {
	if innerWidth < 20 {
		innerWidth = 20
	}
	sprite := p.Sprite(reduceMotion)
	sw := runewidth.StringWidth(sprite)
	if sw <= 0 {
		sw = 7
	}

	dots := make([]rune, innerWidth)
	for i := range dots {
		if i%5 == 0 {
			dots[i] = '·'
		} else {
			dots[i] = ' '
		}
	}
	pos := p.X
	if pos < 0 {
		pos = 0
	}
	if pos+sw > innerWidth {
		pos = innerWidth - sw
	}
	for i, r := range sprite {
		if pos+i < len(dots) {
			dots[pos+i] = r
		}
	}
	lane := string(dots)

	if p.QuipTicks > 0 && p.Quip != "" {
		q := " · ♡ " + p.Quip
		qw := runewidth.StringWidth(q)
		if qw < innerWidth/2 {
			start := innerWidth - qw
			if start < pos+sw+1 {
				start = pos + sw + 1
			}
			if start+qw <= innerWidth {
				runes := []rune(lane)
				for i, r := range q {
					if start+i < len(runes) {
						runes[start+i] = r
					}
				}
				lane = string(runes)
			}
		}
	}
	return padRight(lane, innerWidth)
}

var (
	walkRight   = []string{"/\\_/\\>", "/\\_/\\·", "·/\\_/\\"}
	walkLeft    = []string{"</\\_/\\", "·/\\_/\\", "/\\_/\\·"}
	trotRight   = []string{"/\\_/\\>>", "/\\_/\\> ", " /\\_/\\>"}
	trotLeft    = []string{"<</\\_/\\", " </\\_/\\", "<\\_/\\ "}
	sitSprites  = []string{"(^.^)", "(u.u)", "(^.^)"}
	sleepSprites = []string{"(zZz)", "-.-zZ", " zZz~"}
	scratchSprites = []string{"(>.<)~", "(>.<)~~", "(^v^)♡"}
	purrSprites = []string{"(^v^)♡", "(^o^)~", "(◕‿◕)"}
	wakeSprites = []string{"(O.O)!", "(>.<)~", "(^.^)"}
)

func pickScritchQuip(r *rand.Rand) string {
	return scritchQuips[r.Intn(len(scritchQuips))]
}

func pickPurrQuip(r *rand.Rand) string {
	return purrQuips[r.Intn(len(purrQuips))]
}

func pickWakeQuip(r *rand.Rand) string {
	return wakeQuips[r.Intn(len(wakeQuips))]
}

func pickSentQuip(r *rand.Rand) string {
	return sentQuips[r.Intn(len(sentQuips))]
}

var scritchQuips = []string{
	"mrrp~ that tickles",
	"*purrrrr*",
	"best scritches ♡",
	"right there… yes~",
	"cordy melts (⌒‿⌒)",
	"♡ soft paws happy cat",
}

var purrQuips = []string{
	"*purrr*",
	"(^v^) ♡",
	"mrrrow~",
	"so cozy…",
	"♡ ♡ ♡",
}

var wakeQuips = []string{
	"mrr? oh hi~",
	"*stretch* (^.^)",
	"cordy wakes up!",
	"yawn~ ready to chat",
}

var sentQuips = []string{
	"nice send~",
	"purrfect message!",
	"cordy approves ♡",
	"ship it~ (^.^)",
}
