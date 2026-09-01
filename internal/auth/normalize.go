package auth

import (
	"fmt"
	"regexp"
	"strings"
)

var discordTokenRe = regexp.MustCompile(`(?:mfa\.)?[A-Za-z0-9_-]{20,}\.[A-Za-z0-9_-]{4,}\.[A-Za-z0-9_-]{20,}`)

func NormalizeToken(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", fmt.Errorf("empty token")
	}

	s = strings.Trim(s, `"'`)
	s = strings.TrimPrefix(s, "Bearer ")
	s = strings.TrimPrefix(s, "bearer ")
	s = strings.TrimPrefix(s, "Bot ")

	if idx := strings.Index(strings.ToLower(s), "authorization:"); idx >= 0 {
		s = strings.TrimSpace(s[idx+len("authorization:"):])
		s = strings.Trim(s, `"'`)
	}

	if idx := strings.IndexAny(s, "\r\n"); idx >= 0 {
		s = strings.TrimSpace(s[:idx])
	}

	if !looksLikeDiscordToken(s) {
		if match := discordTokenRe.FindString(s); match != "" {
			s = match
		}
	}

	if err := ValidateTokenShape(s); err != nil {
		return "", err
	}
	return s, nil
}

func looksLikeDiscordToken(s string) bool {
	return discordTokenRe.MatchString(s) && len(s) <= 120
}

func ValidateTokenShape(s string) error {
	if s == "" {
		return fmt.Errorf("empty token")
	}
	if strings.ContainsAny(s, " \t") {
		return fmt.Errorf("token contains spaces — paste only the token, not headers or JSON")
	}
	if len(s) > 120 {
		return fmt.Errorf("token is %d characters (expected ~60–90) — you likely pasted extra text; copy only the token value", len(s))
	}
	parts := strings.Split(s, ".")
	if len(parts) < 3 {
		return fmt.Errorf("token format invalid (expected 3 parts separated by dots, got %d)", len(parts))
	}
	if !discordTokenRe.MatchString(s) {
		return fmt.Errorf("token format invalid — should look like XXXXX.XXXXX.XXXXX (optionally starting with mfa.)")
	}
	return nil
}

func TokenSummary(s string) string {
	parts := len(strings.Split(s, "."))
	return fmt.Sprintf("%d chars, %d segments", len(s), parts)
}
