package ds

import "github.com/charmbracelet/lipgloss"

// Theme is the termcord design system color + typography palette.
type Theme struct {
	Header       lipgloss.Style
	Banner       lipgloss.Style
	Border       lipgloss.Style
	Status       lipgloss.Style
	Footer       lipgloss.Style
	SidebarTitle lipgloss.Style
	Active       lipgloss.Style
	Unread       lipgloss.Style
	Mention      lipgloss.Style
	User         lipgloss.Style
	Dim          lipgloss.Style
	Selected     lipgloss.Style
	Reaction     lipgloss.Style
	ReactionMe   lipgloss.Style
	Embed        lipgloss.Style
	Reply        lipgloss.Style
	Link         lipgloss.Style
	ChannelBar   lipgloss.Style
	Cordy        lipgloss.Style
}

// LoadTheme returns a named palette. Unknown names fall back to default with a note.
func LoadTheme(name string) (Theme, string) {
	switch name {
	case "dracula":
		return draculaTheme(), ""
	case "default", "":
		return defaultTheme(), ""
	default:
		return defaultTheme(), "unknown theme \"" + name + "\" — using default"
	}
}

func draculaTheme() Theme {
	return Theme{
		Header:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("141")),
		Banner:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")),
		Border:       lipgloss.NewStyle().Foreground(lipgloss.Color("236")),
		Status:       lipgloss.NewStyle().Foreground(lipgloss.Color("246")),
		Footer:       lipgloss.NewStyle().Foreground(lipgloss.Color("239")),
		SidebarTitle: lipgloss.NewStyle().Bold(true).Underline(true).Foreground(lipgloss.Color("117")),
		Active:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229")),
		Unread:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("221")),
		Mention:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("203")),
		User:         lipgloss.NewStyle().Foreground(lipgloss.Color("84")),
		Dim:          lipgloss.NewStyle().Foreground(lipgloss.Color("239")),
		Selected:     lipgloss.NewStyle().Foreground(lipgloss.Color("212")),
		Reaction:     lipgloss.NewStyle().Foreground(lipgloss.Color("246")),
		ReactionMe:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("84")),
		Embed:        lipgloss.NewStyle().Foreground(lipgloss.Color("117")),
		Reply:        lipgloss.NewStyle().Foreground(lipgloss.Color("246")).Italic(true),
		Link:         lipgloss.NewStyle().Underline(true).Foreground(lipgloss.Color("111")),
		ChannelBar:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255")),
		Cordy:        lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("219")).Background(lipgloss.Color("235")).Padding(0, 1),
	}
}

func defaultTheme() Theme {
	return Theme{
		Header:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")),
		Banner:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")),
		Border:       lipgloss.NewStyle().Foreground(lipgloss.Color("238")),
		Status:       lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
		Footer:       lipgloss.NewStyle().Foreground(lipgloss.Color("238")),
		SidebarTitle: lipgloss.NewStyle().Bold(true).Underline(true),
		Active:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229")),
		Unread:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220")),
		Mention:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("203")),
		User:         lipgloss.NewStyle().Foreground(lipgloss.Color("117")),
		Dim:          lipgloss.NewStyle().Foreground(lipgloss.Color("238")),
		Selected:     lipgloss.NewStyle().Foreground(lipgloss.Color("212")),
		Reaction:     lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
		ReactionMe:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("117")),
		Embed:        lipgloss.NewStyle().Foreground(lipgloss.Color("109")),
		Reply:        lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true),
		Link:         lipgloss.NewStyle().Underline(true).Foreground(lipgloss.Color("39")),
		ChannelBar:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255")),
		Cordy:        lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Background(lipgloss.Color("236")).Padding(0, 1),
	}
}
