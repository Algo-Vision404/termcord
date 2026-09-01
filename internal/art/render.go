package art

import (
	"strings"

	"github.com/termcord/termcord/internal/ds"
)

// RenderAnimHeader applies theme colors to the animated header.
func RenderAnimHeader(frame int, version, status string, busy bool, opts SceneOpts, t ds.Theme) string {
	raw := AnimHeader(frame, version, status, busy, opts)
	return ds.ApplyChrome3(t, raw, t.Status)
}

// RenderAnimChannelBar applies theme colors to the channel strip.
func RenderAnimChannelBar(frame int, title string, loading bool, opts SceneOpts, t ds.Theme) string {
	raw := AnimChannelBar(frame, title, loading, opts)
	lines := strings.Split(raw, "\n")
	if len(lines) < 3 {
		return t.Border.Render(raw)
	}
	lines[0] = t.Border.Render(lines[0])
	lines[1] = t.ChannelBar.Render(lines[1])
	lines[2] = t.Border.Render(lines[2])
	return strings.Join(lines, "\n")
}

// RenderAnimFooter applies theme colors to the footer.
func RenderAnimFooter(frame int, t ds.Theme) string {
	return ds.ApplyFooter(t, AnimFooter(frame))
}
