package main

import (
	"bufio"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/termcord/termcord/internal/auth"
	"github.com/termcord/termcord/internal/cache"
	"github.com/termcord/termcord/internal/config"
	"github.com/termcord/termcord/internal/gateway"
	"github.com/termcord/termcord/internal/model"
	"github.com/termcord/termcord/internal/plugins"
	"github.com/termcord/termcord/internal/tui"
	"github.com/termcord/termcord/internal/version"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "login":
			runLogin(os.Args[2:])
			return
		case "logout":
			runLogout()
			return
		case "init":
			runInit(os.Args[2:])
			return
		case "doctor":
			runDoctor(os.Args[2:])
			return
		case "version", "-v", "--version":
			fmt.Println("termcord", version.Version)
			return
		case "help", "-h", "--help":
			printUsage()
			return
		}
	}
	runTUI(os.Args[1:])
}

func runTUI(args []string) {
	fs := flag.NewFlagSet("termcord", flag.ExitOnError)
	cfgPath := fs.String("config", "", "path to config.toml")
	_ = fs.Parse(args)

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	token, err := auth.ResolveToken(cfg.General.Token)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, "\nSetup:")
		fmt.Fprintln(os.Stderr, "  termcord init")
		fmt.Fprintln(os.Stderr, "  termcord login")
		os.Exit(1)
	}

	store, err := openCache(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "cache:", err)
		os.Exit(1)
	}
	defer store.Close()

	var pluginList []plugins.Plugin
	if cfg.Plugins.Enabled {
		pluginList, err = plugins.LoadDir(cfg.PluginDir())
		if err != nil {
			fmt.Fprintln(os.Stderr, "plugins:", err)
			os.Exit(1)
		}
	}

	incoming := make(chan model.Message, 128)
	statusCh := make(chan string, 8)
	client, err := gateway.New(gateway.Options{
		Token: token,
		OnMessage: func(msg model.Message) {
			select {
			case incoming <- msg:
			default:
			}
		},
		OnStatus: func(s string) {
			select {
			case statusCh <- s:
			default:
			}
		},
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := client.Open(); err != nil {
		fmt.Fprintln(os.Stderr, "gateway:", err)
		os.Exit(1)
	}
	defer client.Close()
	defer close(incoming)
	defer close(statusCh)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		_ = client.Close()
		os.Exit(0)
	}()

	if err := tui.Run(client, store, tui.OptionsFromConfig(cfg), pluginList, incoming, statusCh); err != nil {
		slog.Error("tui exited", "err", err)
		os.Exit(1)
	}
}

func runLogin(args []string) {
	fs := flag.NewFlagSet("login", flag.ExitOnError)
	_ = fs.Parse(args)

	token := strings.TrimSpace(fs.Arg(0))
	if token == "" {
		fmt.Fprint(os.Stderr, "Discord user token: ")
		line, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "read token:", err)
			os.Exit(1)
		}
		token = strings.TrimSpace(line)
	}
	if token == "" {
		fmt.Fprintln(os.Stderr, "usage: termcord login [token]")
		os.Exit(1)
	}
	if err := auth.StoreToken(token); err != nil {
		fmt.Fprintln(os.Stderr, "store token:", err)
		os.Exit(1)
	}
	fmt.Println("token saved to OS keyring")
}

func runLogout() {
	if err := auth.ClearToken(); err != nil {
		fmt.Fprintln(os.Stderr, "logout:", err)
		os.Exit(1)
	}
	fmt.Println("token removed from keyring")
}

func runInit(args []string) {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	path := fs.String("config", "", "config path")
	_ = fs.Parse(args)
	out, err := config.Init(*path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("created", out)
}

func runDoctor(args []string) {
	fs := flag.NewFlagSet("doctor", flag.ExitOnError)
	path := fs.String("config", "", "config path")
	_ = fs.Parse(args)

	cfg, err := config.Load(*path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "config:", err)
		os.Exit(1)
	}
	fmt.Println("termcord doctor")
	fmt.Println("  version:", version.Version)
	fmt.Println("  config:", configPath(*path))
	fmt.Println("  cache:", cfg.CachePath())

	token, err := auth.ResolveToken(cfg.General.Token)
	if err != nil {
		fmt.Println("  token: missing — run termcord login")
		os.Exit(1)
	}
	fmt.Printf("  token: present (%d chars)\n", len(token))

	client, err := gateway.New(gateway.Options{Token: token})
	if err != nil {
		fmt.Fprintln(os.Stderr, "gateway:", err)
		os.Exit(1)
	}
	if err := client.Open(); err != nil {
		fmt.Fprintln(os.Stderr, "connect:", err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Printf("  discord: connected as @%s (%s)\n", client.Username(), client.UserID())
	fmt.Printf("  channels: %d\n", len(client.AllSelectableChannels()))
}

func openCache(cfg config.Config) (*cache.Store, error) {
	opts := cache.Options{Path: cfg.CachePath(), Encrypt: cfg.Cache.Encrypt}
	if cfg.Cache.Encrypt {
		key, err := auth.ResolveCacheKey()
		if err != nil {
			return nil, err
		}
		opts.Key = key
	}
	return cache.OpenWith(opts)
}

func configPath(override string) string {
	if override != "" {
		return override
	}
	return config.DefaultConfigPath()
}

func printUsage() {
	fmt.Printf(`termcord %s — terminal-native Discord client

Usage:
  termcord                 Launch interactive TUI
  termcord login [token]   Save token to OS keyring
  termcord logout          Remove saved token
  termcord init            Create default config.toml
  termcord doctor          Verify token and connectivity
  termcord version         Print version
  termcord help            Show this help

Environment:
  TERMCORD_TOKEN           Discord user token
  TERMCORD_CONFIG          Path to config.toml

WARNING: User-token clients violate Discord's Terms of Service.
Use at your own risk. Messages stay local; no third-party relay.
`, version.Version)
}
