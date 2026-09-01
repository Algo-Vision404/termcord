package gateway

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseCloseCode extracts a Discord gateway close code from an error string.
func ParseCloseCode(err error) (int, string, bool) {
	if err == nil {
		return 0, "", false
	}
	msg := err.Error()
	const prefix = "websocket: close "
	if !strings.HasPrefix(msg, prefix) {
		return 0, "", false
	}
	rest := strings.TrimPrefix(msg, prefix)
	colon := strings.Index(rest, ":")
	if colon < 0 {
		return 0, "", false
	}
	code, convErr := strconv.Atoi(strings.TrimSpace(rest[:colon]))
	if convErr != nil {
		return 0, "", false
	}
	text := strings.TrimSpace(rest[colon+1:])
	return code, text, true
}

// IsReconnectable reports whether Open() should retry after this gateway error.
func IsReconnectable(err error) bool {
	code, _, ok := ParseCloseCode(err)
	if !ok {
		// Network errors, timeouts, dial failures — worth retrying.
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "timeout") || strings.Contains(msg, "connection refused") ||
			strings.Contains(msg, "temporary failure") || strings.Contains(msg, "no such host") {
			return true
		}
		return !strings.Contains(msg, "identify:")
	}
	switch code {
	case 4004, 4010, 4011, 4012, 4013, 4014:
		return false
	default:
		return true
	}
}

// FriendlyGatewayError turns gateway failures into actionable copy.
func FriendlyGatewayError(err error) string {
	if err == nil {
		return ""
	}
	if code, text, ok := ParseCloseCode(err); ok {
		switch code {
		case 4004:
			return "authentication failed — run termcord logout, then termcord login with a fresh token"
		case 4003:
			return "session not authenticated — retrying usually fixes this"
		case 4014:
			return "connection rejected by discord (4014) — user-token clients may be blocked in your region or account"
		case 4013:
			return "invalid gateway intents (4013)"
		default:
			if text != "" && text != "Unknown error" {
				return fmt.Sprintf("gateway closed (%d): %s", code, text)
			}
			return fmt.Sprintf("gateway closed with code %d", code)
		}
	}
	msg := err.Error()
	if len(msg) > 120 {
		msg = msg[:117] + "..."
	}
	return msg
}
