package gateway

import "testing"

func TestParseCloseCode(t *testing.T) {
	code, text, ok := ParseCloseCode(fmtErr("websocket: close 4004: Authentication failed"))
	if !ok || code != 4004 || text != "Authentication failed" {
		t.Fatalf("parse failed: ok=%v code=%d text=%q", ok, code, text)
	}
}

func TestIsReconnectable(t *testing.T) {
	if IsReconnectable(fmtErr("websocket: close 4004: Authentication failed")) {
		t.Fatal("4004 should not reconnect")
	}
	if !IsReconnectable(fmtErr("websocket: close 4000: Unknown error")) {
		t.Fatal("4000 should reconnect")
	}
}

func fmtErr(s string) error { return &testErr{s} }

type testErr struct{ s string }

func (e *testErr) Error() string { return e.s }
