package auth

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/termcord/termcord/internal/cache"
	"github.com/zalando/go-keyring"
)

const cacheKeyAccount = "cache-key"

func ClearCacheKey() error {
	return keyring.Delete(serviceName, cacheKeyAccount)
}

func ResolveCacheKey() ([]byte, error) {
	raw, err := keyring.Get(serviceName, cacheKeyAccount)
	if err != nil {
		key, genErr := cache.GenerateKey()
		if genErr != nil {
			return nil, genErr
		}
		encoded := base64.StdEncoding.EncodeToString(key)
		if setErr := keyring.Set(serviceName, cacheKeyAccount, encoded); setErr != nil {
			return nil, setErr
		}
		return key, nil
	}
	raw = strings.TrimSpace(raw)
	key, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("decode cache key: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("invalid cache key length")
	}
	return key, nil
}
