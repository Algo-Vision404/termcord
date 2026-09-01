package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/termcord/termcord/internal/art"
	"github.com/termcord/termcord/internal/auth"
	"github.com/termcord/termcord/internal/cache"
	"github.com/termcord/termcord/internal/config"
	"github.com/termcord/termcord/internal/gateway"
	"github.com/termcord/termcord/internal/model"
	"github.com/termcord/termcord/internal/plugins"
	"github.com/termcord/termcord/internal/tui"
	"github.com/termcord/termcord/internal/ux"
	"github.com/termcord/termcord/internal/version"
	"golang.org/x/term"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "login":
			runLogin(os.Args[2:])
			return
		case "logout":
			runLogout(os.Args[2:])
			return
		case "init":
			runInit(os.Args[2:])
			return
		case "doctor":
			runDoctor(os.Args[2:])
			return
		case "version", "-v", "--version":
			fmt.Println(art.CLILogo(version.Version))
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
	quiet := fs.Bool("quiet", false, "skip startup banner")
	_ = fs.Parse(args)

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	printSecurityWarnings(cfg, *quiet)

	if !*quiet && term.IsTerminal(int(os.Stderr.Fd())) {
		art.PrintBootBanner(os.Stderr, version.Version, !cfg.UI.ReduceMotion)
	}

	token, err := auth.ResolveToken(cfg.General.Token)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, ux.TokenSetupHint())
		os.Exit(1)
	}

	if !*quiet && term.IsTerminal(int(os.Stderr.Fd())) {
		art.PrintBootStep(os.Stderr, "Opening local cache...")
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
	typingCh := make(chan tui.TypingEvent, 8)
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
		OnTyping: func(channelID string, users []string) {
			select {
			case typingCh <- tui.TypingEvent{ChannelID: channelID, Users: users}:
			default:
			}
		},
		OnError: func(err error) {
			select {
			case statusCh <- ux.Friendly(err):
			default:
			}
		},
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if !*quiet && term.IsTerminal(int(os.Stderr.Fd())) {
		art.PrintBootStep(os.Stderr, "Connecting to Discord...")
	}

	if err := client.Open(); err != nil {
		fmt.Fprintln(os.Stderr, "couldn't connect:", gateway.FriendlyGatewayError(err))
		fmt.Fprintln(os.Stderr, "try: termcord doctor")
		os.Exit(1)
	}
	defer client.Close()
	defer close(incoming)
	defer close(statusCh)
	defer close(typingCh)

	if !*quiet && term.IsTerminal(int(os.Stderr.Fd())) {
		art.PrintBootStep(os.Stderr, "Syncing channels...")
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		_ = client.Close()
		os.Exit(0)
	}()

	if err := tui.Run(client, store, tui.OptionsFromConfig(cfg), pluginList, incoming, statusCh, typingCh); err != nil {
		fmt.Fprintln(os.Stderr, "termcord exited:", err)
		os.Exit(1)
	}
}

func runLogin(args []string) {
	fs := flag.NewFlagSet("login", flag.ExitOnError)
	tokenFile := fs.String("file", "", "read token from a file")
	_ = fs.Parse(args)

	token := strings.TrimSpace(strings.Join(fs.Args(), " "))
	if *tokenFile != "" {
		data, err := os.ReadFile(*tokenFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, "read token file:", err)
			os.Exit(1)
		}
		token = strings.TrimSpace(string(data))
	}
	if token == "" {
		token = readTokenInteractive()
	}
	if token == "" {
		fmt.Fprintln(os.Stderr, "no token provided")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  termcord login                    (interactive, recommended)")
		fmt.Fprintln(os.Stderr, "  termcord login --file token.txt")
		fmt.Fprintln(os.Stderr, "  set TERMCORD_TOKEN=YOUR_TOKEN   (cmd.exe, current window only)")
		os.Exit(1)
	}
	if *tokenFile == "" && len(fs.Args()) > 0 {
		fmt.Fprintln(os.Stderr, "warning: passing a token on the command line may expose it in process listings")
		fmt.Fprintln(os.Stderr, "         prefer interactive login or --file (then delete the file)")
	}
	if err := auth.StoreToken(token); err != nil {
		fmt.Fprintln(os.Stderr, "couldn't save token:", err)
		os.Exit(1)
	}
	fmt.Println(art.CLILogo(version.Version))
	fmt.Println()
	fmt.Println(art.CLIBox("login", []string{
		ux.KeyringSavedMessage(),
		"",
		"next: termcord doctor",
		"      termcord",
	}))
}

func readTokenInteractive() string {
	fmt.Fprintln(os.Stderr, ux.LoginIntro())
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Paste your token below (input is hidden).")

	if term.IsTerminal(int(os.Stdin.Fd())) {
		fmt.Fprint(os.Stderr, "Token: ")
		bytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr)
		if err != nil {
			fmt.Fprintln(os.Stderr, "read token:", err)
			os.Exit(1)
		}
		return strings.TrimSpace(string(bytes))
	}

	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		fmt.Fprintln(os.Stderr, "read token:", err)
		os.Exit(1)
	}
	return strings.TrimSpace(line)
}

func runLogout(args []string) {
	fs := flag.NewFlagSet("logout", flag.ExitOnError)
	purgeCache := fs.Bool("purge-cache", false, "delete local message cache and encryption key")
	cfgPath := fs.String("config", "", "config path")
	_ = fs.Parse(args)

	if err := auth.ClearToken(); err != nil {
		fmt.Fprintln(os.Stderr, "logout:", err)
		os.Exit(1)
	}

	var lines []string
	lines = append(lines, "token removed from keyring")

	if *purgeCache {
		cfg, err := config.Load(*cfgPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "config:", err)
			os.Exit(1)
		}
		path := cfg.CachePath()
		for _, p := range []string{path, path + "-wal", path + "-shm"} {
			if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
				fmt.Fprintln(os.Stderr, "purge cache:", err)
				os.Exit(1)
			}
		}
		if err := auth.ClearCacheKey(); err != nil {
			fmt.Fprintln(os.Stderr, "purge cache key:", err)
			os.Exit(1)
		}
		lines = append(lines, "local cache deleted")
	}

	fmt.Println(art.CLILogo(version.Version))
	fmt.Println()
	fmt.Println(art.CLIBox("logout", lines))
}

func runInit(args []string) {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	path := fs.String("config", "", "config path")
	_ = fs.Parse(args)
	out, created, err := config.Init(*path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(art.CLILogo(version.Version))
	fmt.Println()
	if created {
		fmt.Println(art.CLIBox("init", []string{
			"created " + out,
			"",
			"next steps:",
			"  1. termcord login",
			"  2. termcord doctor",
			"  3. termcord",
		}))
		return
	}
	fmt.Println(art.CLIBox("init", []string{"config already exists at " + out}))
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
	fmt.Println(art.CLILogo(version.Version))
	fmt.Println()
	var lines []string
	lines = append(lines, "version: "+version.Version)
	lines = append(lines, "config: "+configPath(*path))
	lines = append(lines, "cache: "+cfg.CachePath())
	if cfg.UI.ShowMascot {
		lines = append(lines, "cordy: enabled (set show_mascot = false to hide)")
	} else {
		lines = append(lines, "cordy: off (default) — set show_mascot = true to enable")
	}
	if cfg.UI.ReduceMotion {
		lines = append(lines, "motion: calm mode (reduce_motion = true)")
	} else {
		lines = append(lines, "motion: full animations (reduce_motion = false)")
	}

	token, err := auth.ResolveToken(cfg.General.Token)
	if err != nil {
		lines = append(lines, "token: "+err.Error())
		fmt.Println(art.CLIBox("doctor", lines))
		os.Exit(1)
	}
	lines = append(lines, "token: present ("+auth.TokenSummary(token)+")")
	lines = append(lines, "token source: "+auth.TokenSource(cfg.General.Token))
	for _, w := range config.SecurityWarnings(cfg) {
		lines = append(lines, "security: "+w)
	}

	client, err := gateway.New(gateway.Options{Token: token})
	if err != nil {
		fmt.Fprintln(os.Stderr, "gateway:", err)
		os.Exit(1)
	}

	user, err := client.Session().User("@me")
	if err != nil {
		lines = append(lines, "rest: failed — token invalid or expired")
		lines = append(lines, ux.DoctorFixHint())
		fmt.Println(art.CLIBox("doctor", lines))
		os.Exit(1)
	}
	lines = append(lines, "rest: ok (@"+user.Username+")")
	fmt.Println(art.CLIBox("doctor · checks", lines))

	if err := art.RunWithSpinner("gateway: connecting", func() error {
		return client.Open()
	}); err != nil {
		lines = append(lines, "gateway: "+err.Error())
		fmt.Println(art.CLIBox("doctor · failed", lines))
		os.Exit(1)
	}
	defer client.Close()

	final := []string{
		"gateway: connected",
		fmt.Sprintf("discord: @%s (%s)", client.Username(), client.UserID()),
		fmt.Sprintf("channels: %d ready", len(client.AllSelectableChannels())),
		"",
		"you're good to go — run termcord",
	}
	fmt.Println(art.CLIBox("doctor · ok", final))
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

func printSecurityWarnings(cfg config.Config, quiet bool) {
	if quiet || !term.IsTerminal(int(os.Stderr.Fd())) {
		return
	}
	for _, w := range config.SecurityWarnings(cfg) {
		fmt.Fprintln(os.Stderr, "security:", w)
	}
}

func printUsage() {
	fmt.Println(art.CLILogo(version.Version))
	fmt.Println()
	fmt.Printf(`Usage:
  termcord                 Launch interactive TUI
  termcord login [token]   Save token to OS keyring (prefer interactive)
  termcord logout          Remove saved token
  termcord logout --purge-cache  Also delete local message cache
  termcord init            Create default config.toml
  termcord doctor          Verify token and connectivity
  termcord version         Print version
  termcord help            Show this help

Environment:
  TERMCORD_TOKEN           Discord user token
  TERMCORD_CONFIG          Path to config.toml

WARNING: User-token clients violate Discord's Terms of Service.
Use at your own risk. Messages stay local; no third-party relay.
`)
}
