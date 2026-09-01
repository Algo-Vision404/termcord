package art

import (
	"fmt"

	"github.com/termcord/termcord/internal/ds"
)

// RenderConnecting picks calm or animated connecting splash.
func RenderConnecting(frame int, version string, opts SceneOpts) string {
	if opts.ReduceMotion {
		return CalmConnecting(version)
	}
	return AnimConnecting(frame, version, opts)
}

// RenderLoading picks calm or animated history loader.
func RenderLoading(frame int, channel string, opts SceneOpts) string {
	if opts.ReduceMotion {
		return CalmLoading(channel)
	}
	return AnimLoading(frame, channel, opts)
}

// RenderEmpty picks calm or animated empty channel view.
func RenderEmpty(frame int, opts SceneOpts) string {
	if opts.ReduceMotion {
		return CalmEmpty()
	}
	return AnimEmpty(frame, opts)
}

// RenderWelcome picks calm or animated welcome deck.
func RenderWelcome(username string, channels, guilds int, opts SceneOpts) string {
	if opts.ReduceMotion {
		return CalmWelcome(username, channels, guilds)
	}
	return WelcomeConnected(username, channels, guilds, opts)
}

// RenderFarewell picks calm or animated quit screen.
func RenderFarewell(frame int, opts SceneOpts) string {
	if opts.ReduceMotion {
		return CalmFarewell()
	}
	return AnimFarewell(frame, opts)
}

// CalmConnecting is a simple, static connecting screen.
func CalmConnecting(version string) string {
	return calmLines(
		"Connecting to Discord…",
		"",
		fmt.Sprintf("termcord v%s", version),
		"Usually takes a few seconds.",
		"Ctrl+C to cancel.",
	)
}

// CalmLoading is a simple history loader.
func CalmLoading(channel string) string {
	if channel == "" {
		channel = "channel"
	}
	return calmLines(
		fmt.Sprintf("Loading %s…", channel),
		"",
		"Fetching from Discord and updating your local cache.",
	)
}

// CalmEmpty is shown in channels with no messages yet.
func CalmEmpty() string {
	return calmLines(
		"No messages here yet.",
		"",
		"Type below to send the first message.",
	)
}

// CalmWelcome is a short first-run hint.
func CalmWelcome(username string, channels, guilds int) string {
	return calmLines(
		fmt.Sprintf("Welcome, @%s.", username),
		"",
		fmt.Sprintf("%d channels across %d servers are ready.", channels, guilds),
		"",
		"Press any key to dismiss and start chatting.",
	)
}

// CalmFarewell is a simple goodbye.
func CalmFarewell() string {
	return calmLines(
		"Goodbye.",
		"",
		"Your local cache is saved. See you next time.",
	)
}

func calmLines(lines ...string) string {
	out := "\n"
	for _, line := range lines {
		if line == "" {
			out += "\n"
			continue
		}
		out += "  " + line + "\n"
	}
	return out
}

// FriendlyFooter is the default calm key reference (2 rows).
func FriendlyFooter() string {
	return ds.ShortcutsFooter()
}
