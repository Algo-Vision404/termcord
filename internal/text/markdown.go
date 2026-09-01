package text

import (
	"regexp"
	"strings"
)

var (
	mentionRe   = regexp.MustCompile(`<@!?(\d+)>`)
	channelRe   = regexp.MustCompile(`<#(\d+)>`)
	customEmoji = regexp.MustCompile(`<(a)?:(\w+):\d+>`)
	roleMention = regexp.MustCompile(`<@&(\d+)>`)
	urlRe       = regexp.MustCompile(`https?://[^\s<>]+`)
)

func RenderDiscord(content string, width int) string {
	if content == "" {
		return ""
	}
	s := content
	s = customEmoji.ReplaceAllString(s, ":$2:")
	s = mentionRe.ReplaceAllString(s, "@user")
	s = roleMention.ReplaceAllString(s, "@role")
	s = channelRe.ReplaceAllString(s, "#channel")
	s = strings.ReplaceAll(s, "||", "")
	s = boldify(s)
	s = codify(s)
	if width > 0 {
		s = wrapLines(s, width)
	}
	return s
}

func HighlightLinks(s string, style func(string) string) string {
	if style == nil || s == "" {
		return s
	}
	var b strings.Builder
	last := 0
	for _, idx := range urlRe.FindAllStringIndex(s, -1) {
		b.WriteString(s[last:idx[0]])
		b.WriteString(style(s[idx[0]:idx[1]]))
		last = idx[1]
	}
	b.WriteString(s[last:])
	return b.String()
}

func boldify(s string) string {
	for strings.Contains(s, "**") {
		start := strings.Index(s, "**")
		end := strings.Index(s[start+2:], "**")
		if end < 0 {
			break
		}
		end += start + 2
		inner := s[start+2 : end]
		s = s[:start] + inner + s[end+2:]
	}
	return s
}

func codify(s string) string {
	s = strings.ReplaceAll(s, "```", "")
	for strings.Contains(s, "`") {
		i := strings.Index(s, "`")
		j := strings.Index(s[i+1:], "`")
		if j < 0 {
			break
		}
		j += i + 1
		s = s[:i] + s[i+1:j] + s[j+1:]
	}
	return s
}

func wrapLines(s string, width int) string {
	if width <= 0 {
		return s
	}
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if len(line) > width {
			lines[i] = line[:width-1] + "…"
		}
	}
	return strings.Join(lines, "\n")
}

func FormatEmbeds(embeds []Embed) string {
	if len(embeds) == 0 {
		return ""
	}
	var parts []string
	for _, e := range embeds {
		parts = append(parts, e.String())
	}
	return strings.Join(parts, "\n")
}

type Embed struct {
	Title       string
	Description string
	URL         string
	Author      string
	Fields      []EmbedField
}

type EmbedField struct {
	Name  string
	Value string
}

func (e Embed) String() string {
	var b strings.Builder
	b.WriteString("┌─ embed")
	if e.Title != "" {
		b.WriteString(" · ")
		b.WriteString(e.Title)
	}
	b.WriteString("\n")
	if e.Author != "" {
		b.WriteString("│ ")
		b.WriteString(e.Author)
		b.WriteString("\n")
	}
	if e.Description != "" {
		for _, line := range strings.Split(e.Description, "\n") {
			b.WriteString("│ ")
			b.WriteString(strings.TrimSpace(line))
			b.WriteString("\n")
		}
	}
	for _, f := range e.Fields {
		b.WriteString("│ ")
		b.WriteString(f.Name)
		b.WriteString(": ")
		b.WriteString(strings.ReplaceAll(f.Value, "\n", " "))
		b.WriteString("\n")
	}
	if e.URL != "" {
		b.WriteString("└ ")
		b.WriteString(e.URL)
	} else {
		b.WriteString("└")
	}
	return b.String()
}

func FormatMessage(content string, embeds []Embed, width int) string {
	body := RenderDiscord(content, width)
	extra := FormatEmbeds(embeds)
	if extra == "" {
		return body
	}
	if body == "" {
		return extra
	}
	return body + "\n" + extra
}
