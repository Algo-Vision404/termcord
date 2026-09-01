package auth

import (
	"fmt"
	"os"
	"strings"

	"github.com/zalando/go-keyring"
)

const (
	serviceName = "termcord"
	accountName = "discord-token"
	envVar      = "TERMCORD_TOKEN"
)

func ResolveToken(configToken string) (string, error) {
	var raw string
	if v := strings.TrimSpace(os.Getenv(envVar)); v != "" {
		raw = v
	} else if v := strings.TrimSpace(configToken); v != "" {
		raw = v
	} else {
		token, err := keyring.Get(serviceName, accountName)
		if err != nil {
			return "", fmt.Errorf("no token found — run termcord init, then termcord login (or set %s)", envVar)
		}
		raw = token
	}
	return NormalizeToken(raw)
}

func StoreToken(token string) error {
	normalized, err := NormalizeToken(token)
	if err != nil {
		return err
	}
	return keyring.Set(serviceName, accountName, normalized)
}

func ClearToken() error {
	return keyring.Delete(serviceName, accountName)
}

// TokenSource reports where ResolveToken would read the token from.
func TokenSource(configToken string) string {
	if v := strings.TrimSpace(os.Getenv(envVar)); v != "" {
		return "environment"
	}
	if v := strings.TrimSpace(configToken); v != "" {
		return "config"
	}
	return "keyring"
}
