package auth

import (
	"testing"
)

func TestNormalizeTokenExtractsEmbedded(t *testing.T) {
	raw := `Authorization: "mfa.` + stringsRepeat("a", 24) + `.` + stringsRepeat("b", 6) + `.` + stringsRepeat("c", 27) + `"`
	got, err := NormalizeToken(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !looksLikeDiscordToken(got) {
		t.Fatalf("unexpected %q", got)
	}
}

func TestValidateTokenShapeRejectsLong(t *testing.T) {
	long := stringsRepeat("a", 200) + "." + stringsRepeat("b", 10) + "." + stringsRepeat("c", 30)
	err := ValidateTokenShape(long)
	if err == nil {
		t.Fatal("expected error for long token")
	}
}

func stringsRepeat(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}
