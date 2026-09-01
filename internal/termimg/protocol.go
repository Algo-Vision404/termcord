package termimg

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/termcord/termcord/internal/model"
)

type Protocol string

const (
	ProtocolAuto   Protocol = "auto"
	ProtocolKitty  Protocol = "kitty"
	ProtocolITerm2 Protocol = "iterm2"
	ProtocolNone   Protocol = "none"
)

const maxInlineBytes = 512 * 1024

// Resolve picks the active image protocol from config and environment.
func Resolve(config string) Protocol {
	switch Protocol(strings.ToLower(strings.TrimSpace(config))) {
	case ProtocolKitty:
		return ProtocolKitty
	case ProtocolITerm2:
		return ProtocolITerm2
	case ProtocolNone:
		return ProtocolNone
	case ProtocolAuto:
		return DetectTerminal()
	default:
		return DetectTerminal()
	}
}

// DetectTerminal guesses support from the environment.
func DetectTerminal() Protocol {
	term := strings.ToLower(os.Getenv("TERM"))
	if strings.Contains(term, "xterm-kitty") || os.Getenv("KITTY_WINDOW_ID") != "" {
		return ProtocolKitty
	}
	if os.Getenv("TERM_PROGRAM") == "iTerm.app" || os.Getenv("ITERM_SESSION_ID") != "" {
		return ProtocolITerm2
	}
	if strings.Contains(term, "ghostty") {
		return ProtocolKitty
	}
	return ProtocolNone
}

// FormatAttachments renders attachment lines for chat view.
func FormatAttachments(atts []model.Attachment, proto Protocol, token string, cache map[string]string) string {
	if len(atts) == 0 {
		return ""
	}
	var lines []string
	for _, att := range atts {
		line := formatOne(att, proto, token, cache)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}

func formatOne(att model.Attachment, proto Protocol, token string, cache map[string]string) string {
	label := att.Filename
	if label == "" {
		label = "attachment"
	}
	if !isImage(att) {
		return fmt.Sprintf("  📎 %s", label)
	}
	if proto == ProtocolNone {
		return fmt.Sprintf("  🖼 %s  %s", label, shortURL(att.URL))
	}
	if cached, ok := cache[att.URL]; ok {
		return cached
	}
	inline, err := fetchInline(att, proto, token)
	if err != nil {
		return fmt.Sprintf("  🖼 %s  (%s)", label, err.Error())
	}
	cache[att.URL] = inline
	return inline
}

func isImage(att model.Attachment) bool {
	if strings.HasPrefix(att.ContentType, "image/") {
		return true
	}
	lower := strings.ToLower(att.Filename)
	return strings.HasSuffix(lower, ".png") ||
		strings.HasSuffix(lower, ".jpg") ||
		strings.HasSuffix(lower, ".jpeg") ||
		strings.HasSuffix(lower, ".gif") ||
		strings.HasSuffix(lower, ".webp")
}

func fetchInline(att model.Attachment, proto Protocol, token string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, att.URL, nil)
	if err != nil {
		return "", err
	}
	if token != "" {
		req.Header.Set("Authorization", token)
	}
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("fetch %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxInlineBytes+1))
	if err != nil {
		return "", err
	}
	if len(data) > maxInlineBytes {
		return "", fmt.Errorf("image too large for inline preview")
	}
	b64 := base64.StdEncoding.EncodeToString(data)
	width := att.Width
	height := att.Height
	if width <= 0 {
		width = 400
	}
	if height <= 0 {
		height = 0
	}
	switch proto {
	case ProtocolKitty:
		return kittySequence(b64, width, height), nil
	default:
		return iterm2Sequence(b64, width), nil
	}
}

func iterm2Sequence(b64 string, width int) string {
	args := "inline=1"
	if width > 0 {
		args += fmt.Sprintf(";width=%dpx", min(width, 600))
	}
	return fmt.Sprintf("\033]1337;File=%s:%s\a", args, b64)
}

func kittySequence(b64 string, width, height int) string {
	w := min(width, 600)
	h := height
	if h <= 0 {
		h = 0
	}
	if h > 0 {
		return fmt.Sprintf("\033_Ga=T,f=100,t=d,d=V,i=png,s=%d,v=%d,c=1;%s\033\\", w, h, b64)
	}
	return fmt.Sprintf("\033_Ga=T,f=100,t=d,d=V,i=png,s=%d,c=1;%s\033\\", w, b64)
}

func shortURL(u string) string {
	if len(u) <= 48 {
		return u
	}
	return u[:45] + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
