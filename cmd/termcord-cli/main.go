package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/termcord/termcord/internal/auth"
	"github.com/termcord/termcord/internal/config"
	"github.com/termcord/termcord/internal/gateway"
	"github.com/termcord/termcord/internal/model"
	"github.com/termcord/termcord/internal/version"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "send":
		runSend(os.Args[2:])
	case "whoami":
		runWhoami()
	case "channels":
		runChannels(os.Args[2:])
	case "history":
		runHistory(os.Args[2:])
	case "version":
		fmt.Println("termcord-cli", version.Version)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(2)
	}
}

func runSend(args []string) {
	fs := flag.NewFlagSet("send", flag.ExitOnError)
	channel := fs.String("channel", "", "channel ID")
	reply := fs.String("reply", "", "message ID to reply to")
	configPath := fs.String("config", "", "config path")
	_ = fs.Parse(args)

	content := strings.TrimSpace(strings.Join(fs.Args(), " "))
	if *channel == "" || content == "" {
		fmt.Fprintln(os.Stderr, "usage: termcord-cli send -channel ID [-reply MSG_ID] message")
		os.Exit(2)
	}

	client := openClient(*configPath)
	defer client.Close()

	var (
		msg *model.Message
		err error
	)
	if *reply != "" {
		msg, err = client.SendReply(*channel, content, *reply)
	} else {
		msg, err = client.SendMessage(*channel, content)
	}
	if err != nil {
		exitErr(err)
	}
	out, _ := json.MarshalIndent(msg, "", "  ")
	fmt.Println(string(out))
}

func runWhoami() {
	client := openClient("")
	defer client.Close()
	out, _ := json.MarshalIndent(map[string]string{
		"id":       client.UserID(),
		"username": client.Username(),
		"version":  version.Version,
	}, "", "  ")
	fmt.Println(string(out))
}

func runChannels(args []string) {
	fs := flag.NewFlagSet("channels", flag.ExitOnError)
	configPath := fs.String("config", "", "config path")
	jsonOut := fs.Bool("json", false, "JSON output")
	_ = fs.Parse(args)

	client := openClient(*configPath)
	defer client.Close()

	sections := client.ChannelSections()
	if *jsonOut {
		out, _ := json.MarshalIndent(sections, "", "  ")
		fmt.Println(string(out))
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for _, sec := range sections {
		fmt.Fprintf(w, "[%s]\n", sec.Title)
		for _, ch := range sec.Channels {
			fmt.Fprintf(w, "  %s\t%s\n", ch.ID, ch.Name)
		}
	}
	_ = w.Flush()
}

func runHistory(args []string) {
	fs := flag.NewFlagSet("history", flag.ExitOnError)
	channel := fs.String("channel", "", "channel ID")
	limit := fs.Int("limit", 20, "message count")
	configPath := fs.String("config", "", "config path")
	_ = fs.Parse(args)

	if *channel == "" {
		fmt.Fprintln(os.Stderr, "usage: termcord-cli history -channel ID")
		os.Exit(2)
	}

	client := openClient(*configPath)
	defer client.Close()

	msgs, err := client.FetchHistory(*channel, *limit)
	if err != nil {
		exitErr(err)
	}
	out, _ := json.MarshalIndent(msgs, "", "  ")
	fmt.Println(string(out))
}

func openClient(configPath string) *gateway.Client {
	cfg, err := config.Load(configPath)
	if err != nil {
		exitErr(err)
	}
	token, err := auth.ResolveToken(cfg.General.Token)
	if err != nil {
		exitErr(err)
	}
	client, err := gateway.New(gateway.Options{Token: token})
	if err != nil {
		exitErr(err)
	}
	if err := client.Open(); err != nil {
		exitErr(err)
	}
	return client
}

func exitErr(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func printUsage() {
	fmt.Printf(`termcord-cli %s — headless termcord

Usage:
  termcord-cli send -channel ID message
  termcord-cli send -channel ID -reply MSG_ID message
  termcord-cli whoami
  termcord-cli channels [-json]
  termcord-cli history -channel ID [-limit N]
  termcord-cli version
  termcord-cli help
`, version.Version)
}
