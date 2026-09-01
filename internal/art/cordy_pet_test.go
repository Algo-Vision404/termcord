package art

import (
	"testing"

	"github.com/mattn/go-runewidth"
)

func TestCordyPetWalkBounds(t *testing.T) {
	p := NewCordyPet()
	p.State = PetWalk
	for i := 0; i < 200; i++ {
		p.Tick(40, PetHints{Ready: true})
		if p.X < 0 || p.X > 40 {
			t.Fatalf("x out of bounds: %d", p.X)
		}
	}
}

func TestCordyScritch(t *testing.T) {
	p := NewCordyPet()
	p.State = PetSleep
	p.Scritch()
	if p.State != PetScratch {
		t.Fatalf("want scratch, got %v", p.State)
	}
	if p.Quip == "" {
		t.Fatal("expected quip")
	}
}

func TestRenderTrack(t *testing.T) {
	p := NewCordyPet()
	p.X = 5
	p.QuipTicks = 0
	lane := p.RenderTrack(40, true)
	if runewidth.StringWidth(lane) != 40 {
		t.Fatalf("lane width %d want 40", runewidth.StringWidth(lane))
	}
}
