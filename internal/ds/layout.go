package ds

// PanelInner is the content width for full-width chrome (header, compose, footer).
func PanelInner(terminalWidth int) int {
	if terminalWidth <= 0 {
		return StandardInner
	}
	return ClampInner(terminalWidth)
}

// ChatInner is the content width for the message stream beside an optional sidebar.
func ChatInner(terminalWidth, sidebarWidth int, sidebarOpen bool) int {
	used := 0
	if sidebarOpen {
		used = sidebarWidth + 1 // sidebar + separator
	}
	inner := terminalWidth - used
	if inner < MinInner {
		inner = MinInner
	}
	return inner
}
