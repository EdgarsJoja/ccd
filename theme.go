package main

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	primaryBackground   lipgloss.Color
	secondaryBackground lipgloss.Color
	tertiaryBackground  lipgloss.Color
	primaryContent      lipgloss.Color
	success             lipgloss.Color
	info                lipgloss.Color
	warning             lipgloss.Color
	error               lipgloss.Color
}

var darkTheme = Theme{
	primaryBackground:   lipgloss.Color("#090A0C"),
	secondaryBackground: lipgloss.Color("#1B1C22"),
	tertiaryBackground:  lipgloss.Color("#40424F"),
	primaryContent:      lipgloss.Color("#fff"),
	success:             lipgloss.Color("#80b918"),
	info:                lipgloss.Color("#5fa8d3"),
	warning:             lipgloss.Color("#fca311"),
	error:               lipgloss.Color("#d00000"),
}

var lightTheme = Theme{
	primaryBackground:   lipgloss.Color("#fff"),
	secondaryBackground: lipgloss.Color("#eee"),
	tertiaryBackground:  lipgloss.Color("#ddd"),
	primaryContent:      lipgloss.Color("#000"),
	success:             lipgloss.Color("#80b918"),
	info:                lipgloss.Color("#5fa8d3"),
	warning:             lipgloss.Color("#fca311"),
	error:               lipgloss.Color("#d00000"),
}
