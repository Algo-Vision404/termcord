package plugins

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/termcord/termcord/internal/security"
)

type Plugin struct {
	Name        string
	Description string
	Command     string
}

type file struct {
	Plugins []Plugin `toml:"plugin"`
}

func LoadDir(dir string) ([]Plugin, error) {
	if dir == "" {
		dir = defaultDir()
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []Plugin
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".toml") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		plugins, err := loadFile(path)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}
		out = append(out, plugins...)
	}
	return out, nil
}

func loadFile(path string) ([]Plugin, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var f file
	if err := toml.Unmarshal(data, &f); err != nil {
		return nil, err
	}
	return f.Plugins, nil
}

func defaultDir() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "termcord", "plugins")
	default:
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".config", "termcord", "plugins")
	}
}

func DefaultDir() string {
	return defaultDir()
}

type RunContext struct {
	ChannelID   string
	ChannelName string
	Username    string
	Message     string
	Args        string
}

func (p Plugin) Run(ctx RunContext) (string, error) {
	if strings.TrimSpace(p.Command) == "" {
		return "", fmt.Errorf("plugin %q has empty command", p.Name)
	}
	cmdStr := expand(p.Command, ctx)
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", cmdStr)
	} else {
		cmd = exec.Command("sh", "-c", cmdStr)
	}
	cmd.Env = security.PluginEnv(
		"TERMCORD_CHANNEL_ID="+ctx.ChannelID,
		"TERMCORD_CHANNEL="+ctx.ChannelName,
		"TERMCORD_USER="+ctx.Username,
		"TERMCORD_MESSAGE="+ctx.Message,
		"TERMCORD_ARGS="+ctx.Args,
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("plugin %q: %s", p.Name, msg)
	}
	return strings.TrimSpace(stdout.String()), nil
}

func expand(s string, ctx RunContext) string {
	r := strings.NewReplacer(
		"$CHANNEL_ID", ctx.ChannelID,
		"$CHANNEL", ctx.ChannelName,
		"$USER", ctx.Username,
		"$MESSAGE", ctx.Message,
		"$ARGS", ctx.Args,
	)
	return r.Replace(s)
}

func ByName(plugins []Plugin, name string) (Plugin, bool) {
	for _, p := range plugins {
		if strings.EqualFold(p.Name, name) {
			return p, true
		}
	}
	return Plugin{}, false
}
