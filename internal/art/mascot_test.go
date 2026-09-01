package art

import "testing"

func TestCatFramesNonEmpty(t *testing.T) {
	for mode := MascotIdle; mode <= MascotFarewell; mode++ {
		cat := Cat(mode, 0, false)
		if len(cat) < 3 {
			t.Fatalf("mode %d: expected cat art, got %v", mode, cat)
		}
	}
}

func TestMascotPanelPreservesPanel(t *testing.T) {
	panel := []string{"hello", "world"}
	out := MascotPanel(SceneOpts{ShowMascot: false}, 0, panel)
	if len(out) != len(panel) {
		t.Fatalf("want %d lines, got %d", len(panel), len(out))
	}
}

func TestCatBadge(t *testing.T) {
	if CatBadge(MascotIdle, 0, true) == "" {
		t.Fatal("expected badge")
	}
}
