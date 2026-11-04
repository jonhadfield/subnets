package cli

import "charm.land/lipgloss/v2"

var (
	colAccent = lipgloss.Color("#7CDEDC")
	colMuted  = lipgloss.Color("#6C7086")
	colText   = lipgloss.Color("#CDD6F4")
	colKey    = lipgloss.Color("#A6ADC8")
	colOk     = lipgloss.Color("#A6E3A1")
	colWarn   = lipgloss.Color("#FAB387")
	colBad    = lipgloss.Color("#F38BA8")
	colBorder = lipgloss.Color("#45475A")
	colHeader = lipgloss.Color("#7CDEDC")
	colInput  = lipgloss.Color("#F9E2AF")
)

var (
	brandStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colAccent)

	mutedStyle = lipgloss.NewStyle().Foreground(colMuted)
	keyStyle   = lipgloss.NewStyle().Foreground(colKey)
	valStyle   = lipgloss.NewStyle().Foreground(colText).Bold(true)
	inputStyle = lipgloss.NewStyle().Foreground(colInput).Bold(true)
	errStyle   = lipgloss.NewStyle().Foreground(colBad).Bold(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colBorder).
			Padding(0, 1).
			MarginTop(1)

	boxTitleStyle = lipgloss.NewStyle().
			Foreground(colHeader).
			Bold(true)

	headerStyle = lipgloss.NewStyle().
			Foreground(colHeader).
			Bold(true).
			Padding(0, 1)

	cellStyle = lipgloss.NewStyle().
			Foreground(colText).
			Padding(0, 1)

	oddCellStyle = lipgloss.NewStyle().
			Foreground(colKey).
			Padding(0, 1)
)
