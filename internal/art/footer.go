package art

import "github.com/termcord/termcord/internal/ds"

// StaticFooter is the calm footer used by default and when reduce_motion is on.
func StaticFooter() string {
	return ds.ShortcutsFooter()
}
