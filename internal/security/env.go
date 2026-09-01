package security

import (
	"os"
	"strings"
)

// secretEnvPrefixes lists variables that must not reach plugin subprocesses.
var secretEnvPrefixes = []string{
	"TERMCORD_TOKEN",
	"TERMCORD_CONFIG",
	"DISCORD_TOKEN",
}

// PluginEnv returns a copy of the environment safe for plugin commands.
// Discord tokens and other secrets are stripped; plugin-specific vars are appended.
func PluginEnv(extra ...string) []string {
	out := make([]string, 0, len(os.Environ())+len(extra))
	for _, kv := range os.Environ() {
		key, _, ok := strings.Cut(kv, "=")
		if !ok || isSecretEnv(key) {
			continue
		}
		out = append(out, kv)
	}
	return append(out, extra...)
}

func isSecretEnv(key string) bool {
	upper := strings.ToUpper(strings.TrimSpace(key))
	for _, prefix := range secretEnvPrefixes {
		if upper == prefix || strings.HasPrefix(upper, prefix+"_") {
			return true
		}
	}
	return false
}
