package tui

import "github.com/termcord/termcord/internal/ds"

type Theme = ds.Theme

func LoadTheme(name string) (Theme, string) {
	return ds.LoadTheme(name)
}
