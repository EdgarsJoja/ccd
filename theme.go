package main

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	primaryBackground   lipgloss.Color
	secondaryBackground lipgloss.Color
	tertiaryBackground  lipgloss.Color
	primaryContent      lipgloss.Color
	muted               lipgloss.Color
	accent              lipgloss.Color
	selectedBg          lipgloss.Color
	selectedFg          lipgloss.Color
	border              lipgloss.Color
	title               lipgloss.Color
	success             lipgloss.Color
	info                lipgloss.Color
	warning             lipgloss.Color
	error               lipgloss.Color
}

var darkTheme = Theme{
	primaryBackground:   lipgloss.Color("#0f0f14"),
	secondaryBackground: lipgloss.Color("#1a1b26"),
	tertiaryBackground:  lipgloss.Color("#24283b"),
	primaryContent:      lipgloss.Color("#c0caf5"),
	muted:               lipgloss.Color("#565f89"),
	accent:              lipgloss.Color("#7aa2f7"),
	selectedBg:          lipgloss.Color("#283457"),
	selectedFg:          lipgloss.Color("#7aa2f7"),
	border:              lipgloss.Color("#3b4261"),
	title:               lipgloss.Color("#bb9af7"),
	success:             lipgloss.Color("#9ece6a"),
	info:                lipgloss.Color("#7aa2f7"),
	warning:             lipgloss.Color("#e0af68"),
	error:               lipgloss.Color("#f7768e"),
}

var lightTheme = Theme{
	primaryBackground:   lipgloss.Color("#fafafa"),
	secondaryBackground: lipgloss.Color("#f0f0f0"),
	tertiaryBackground:  lipgloss.Color("#e6e6e6"),
	primaryContent:      lipgloss.Color("#1a1b26"),
	muted:               lipgloss.Color("#888"),
	accent:              lipgloss.Color("#7ab8f5"),
	selectedBg:          lipgloss.Color("#d0e3f7"),
	selectedFg:          lipgloss.Color("#2e7de9"),
	border:              lipgloss.Color("#d0d0d0"),
	title:               lipgloss.Color("#6b4c1e"),
	success:             lipgloss.Color("#4c8023"),
	info:                lipgloss.Color("#2e7de9"),
	warning:             lipgloss.Color("#8c6c3e"),
	error:               lipgloss.Color("#c53a4a"),
}
