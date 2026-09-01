package art

import (
	"fmt"
	"strings"
)

// WelcomeConnected is shown once after the first successful connection.
func WelcomeConnected(username string, channels, guilds int, opts SceneOpts) string {
	user := username
	if len(user) > 20 {
		user = user[:17] + "..."
	}
	panel := []string{
		"  ╔══════════════════════════════════════════════════════════╗",
		fmt.Sprintf("  ║  welcome, @%-45s║", user),
		"  ╠══════════════════════════════════════════════════════════╣",
		"  ║                                                          ║",
		"  ║   your discord in the terminal — direct, local, yours      ║",
		"  ║                                                          ║",
		"  ║   ctrl+g  spotlight   fuzzy jump anywhere                ║",
		"  ║   ctrl+m  mention hop next channel with @you               ║",
		"  ║   ctrl+f  focus mode  hide sidebar clutter               ║",
		"  ║   tab     complete @user and #channel                    ║",
		"  ║   /help   commands & keys                                ║",
		"  ║                                                          ║",
		fmt.Sprintf("  ║   %s synced                              ║", padRight(fmt.Sprintf("%d channels · %d servers", channels, guilds), 42)),
		"  ║                                                          ║",
		"  ║   cordy is listening — type below or ctrl+g to jump       ║",
		"  ║                                                          ║",
		"  ╚══════════════════════════════════════════════════════════╝",
		"",
	}
	opts.Mode = MascotWelcome
	return strings.Join(MascotPanel(opts, 0, panel), "\n")
}
