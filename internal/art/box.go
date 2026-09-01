package art

import "github.com/termcord/termcord/internal/ds"

const innerWidth = ds.StandardInner
const boxWidth = ds.ComposeBox

func boxTop(title string) string     { return ds.Top(innerWidth, title) }
func boxRow(text string) string      { return ds.Row(innerWidth, text) }
func boxBottom() string              { return ds.Bottom(innerWidth) }
func padRight(s string, n int) string { return ds.PadRight(s, n) }

// BoxTop exports a top rule for external renderers.
func BoxTop(title string) string { return boxTop(title) }

// BoxRow exports a padded box line for external renderers.
func BoxRow(text string) string { return boxRow(text) }

// BoxBottom exports a bottom rule for external renderers.
func BoxBottom() string { return boxBottom() }
