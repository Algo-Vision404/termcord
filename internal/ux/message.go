package ux

import (
	"errors"
	"fmt"
	"runtime"
	"strings"

	"github.com/termcord/termcord/internal/gateway"
)

// Friendly turns errors into short, actionable status text.
func Friendly(err error) string {
	if err == nil {
		return ""
	}
	msg := strings.ToLower(err.Error())

	switch {
	case strings.Contains(msg, "403") || strings.Contains(msg, "missing permissions"):
		return "you don't have permission to do that here"
	case strings.Contains(msg, "401") || strings.Contains(msg, "unauthorized"):
		return "session expired — run termcord logout, then termcord login"
	case strings.Contains(msg, "404") || strings.Contains(msg, "unknown channel"):
		return "that channel or message no longer exists"
	case strings.Contains(msg, "429") || strings.Contains(msg, "rate limit"):
		return "rate limited — wait a few seconds and try again"
	case strings.Contains(msg, "500") || strings.Contains(msg, "502") || strings.Contains(msg, "503"):
		return "discord is having trouble — try again shortly"
	case strings.Contains(msg, "couldn't connect within"):
		return err.Error()
	case strings.Contains(msg, "gateway closed"):
		return err.Error()
	case strings.Contains(msg, "authentication failed"):
		return err.Error()
	case strings.Contains(msg, "websocket"):
		if friendly := gateway.FriendlyGatewayError(err); friendly != "" {
			return friendly
		}
		return err.Error()
	case strings.Contains(msg, "no older messages"):
		return "you've reached the beginning of history"
	case strings.Contains(msg, "thread not found"):
		return err.Error()
	}

	// Keep short raw errors; trim noisy prefixes.
	clean := err.Error()
	if len(clean) > 90 {
		clean = clean[:87] + "..."
	}
	return clean
}

// FriendlySend wraps send/history failures.
func FriendlySend(err error) string {
	if err == nil {
		return ""
	}
	return "couldn't send: " + Friendly(err)
}

// TokenSetupHint explains how to authenticate on this machine.
func TokenSetupHint() string {
	lines := []string{
		"Get started:",
		"  1. termcord init",
		"  2. termcord login",
		"  3. termcord",
	}
	if runtime.GOOS == "windows" {
		lines = append(lines,
			"",
			"Or set a token for this window only:",
			"  set TERMCORD_TOKEN=your_token_here",
		)
	} else {
		lines = append(lines,
			"",
			"Or export TERMCORD_TOKEN for this shell.",
		)
	}
	return strings.Join(lines, "\n")
}

// TokenMissingHint is appended when ResolveToken fails.
func TokenMissingHint() string {
	return "No saved token. Run: termcord init  then  termcord login"
}

// LoginIntro is shown before the hidden token prompt.
func LoginIntro() string {
	return strings.Join([]string{
		"Sign in with your Discord user token.",
		"How to find it: Discord in browser → F12 → Network → any request → Authorization header.",
		"Paste only the token (about 60–90 characters). Never share it.",
	}, "\n")
}

// KeyringSavedMessage confirms where the token was stored.
func KeyringSavedMessage() string {
	switch runtime.GOOS {
	case "windows":
		return "token saved to Windows Credential Manager"
	case "darwin":
		return "token saved to macOS Keychain"
	default:
		return "token saved to your OS secret store"
	}
}

// DoctorFixHint tells users how to re-authenticate.
func DoctorFixHint() string {
	if runtime.GOOS == "windows" {
		return "fix: termcord logout  then  termcord login YOUR_TOKEN"
	}
	return "fix: termcord logout && termcord login YOUR_TOKEN"
}

// PluginDirHint explains where to put plugins.
func PluginDirHint(dir string) string {
	return fmt.Sprintf("no plugins loaded — add .toml files to:\n  %s", dir)
}

// CachedHistoryNotice explains offline/cached content.
func CachedHistoryNotice(reason string) string {
	if reason == "" {
		return "showing cached messages (live fetch failed)"
	}
	return "showing cached messages — " + Friendly(errors.New(reason))
}
