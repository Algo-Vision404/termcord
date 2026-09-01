package security

import (
	"os"
	"strings"
	"testing"
)

func TestPluginEnvStripsSecrets(t *testing.T) {
	t.Setenv("TERMCORD_TOKEN", "secret-token")
	t.Setenv("TERMCORD_CONFIG", "C:\\secret\\config.toml")
	t.Setenv("TERMCORD_CHANNEL_ID", "123")
	t.Setenv("PATH", os.Getenv("PATH"))

	env := PluginEnv("TERMCORD_ARGS=hi")
	joined := strings.Join(env, "\n")
	if strings.Contains(joined, "secret-token") {
		t.Fatalf("token leaked into plugin env: %q", joined)
	}
	if strings.Contains(joined, "config.toml") && strings.Contains(joined, "TERMCORD_CONFIG") {
		t.Fatalf("config path leaked into plugin env: %q", joined)
	}
	if !strings.Contains(joined, "TERMCORD_ARGS=hi") {
		t.Fatal("expected extra env vars to be appended")
	}
}
