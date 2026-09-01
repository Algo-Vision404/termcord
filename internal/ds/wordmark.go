package ds

import "strings"

// Boot tagline shown under the wordmark.
const BootTagline = "the terminal-native discord experience"

// Big figlet art for "termcord" (smushed, no 888 styling).
var wordmarkLines = []string{
	" _                                        _ ",
	"| |                                      | |",
	"| |_ ___ _ __ _ __ ___   ___ ___  _ __ __| |",
	"| __/ _ \\ '__| '_ ` _ \\ / __/ _ \\| '__/ _` |",
	"| ||  __/ |  | | | | | | (_| (_) | | | (_| |",
	" \\__\\___|_|  |_| |_| |_|\\___\\___/|_|  \\__,_|",
}

const wordmarkPromptLine = 3

func renderWordmark() string {
	lines := make([]string, len(wordmarkLines))
	copy(lines, wordmarkLines)
	if wordmarkPromptLine < len(lines) {
		lines[wordmarkPromptLine] = "> " + strings.TrimLeft(lines[wordmarkPromptLine], " ")
	}
	return strings.Join(lines, "\n")
}
