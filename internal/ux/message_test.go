package ux

import "testing"

func TestFriendly403(t *testing.T) {
	got := Friendly(err403())
	if got != "you don't have permission to do that here" {
		t.Fatalf("got %q", got)
	}
}

type stubErr string

func (e stubErr) Error() string { return string(e) }

func err403() error { return stubErr("HTTP 403 Forbidden") }
