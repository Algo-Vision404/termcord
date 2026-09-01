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
	if v := strings.TrimSpace(os.Getenv(envVar)); v != "" {
		return v, nil
	}
	if v := strings.TrimSpace(configToken); v != "" {
		return v, nil
	}
	token, err := keyring.Get(serviceName, accountName)
	if err != nil {
		return "", fmt.Errorf("no token found: set %s, config general.token, or run termcord login", envVar)
	}
	return token, nil
}

func StoreToken(token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return fmt.Errorf("empty token")
	}
	return keyring.Set(serviceName, accountName, token)
}

func ClearToken() error {
	return keyring.Delete(serviceName, accountName)
}
