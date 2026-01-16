package repl

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/zepzeper/vulgar/internal/cli"
)

var (
	promptStyle = lipgloss.NewStyle().
			Foreground(cli.ColorPrimary).
			Bold(true)

	continueStyle = lipgloss.NewStyle().
			Foreground(cli.ColorMuted)

	errorStyle = lipgloss.NewStyle().
			Foreground(cli.ColorError)

	successStyle = lipgloss.NewStyle().
			Foreground(cli.ColorSecondary)

	resultStyle = lipgloss.NewStyle().
			Foreground(cli.ColorInfo)

	keyStyle = lipgloss.NewStyle().
			Foreground(cli.ColorCode)

	stringStyle = lipgloss.NewStyle().
			Foreground(cli.ColorSecondary)

	numberStyle = lipgloss.NewStyle().
			Foreground(cli.ColorInfo)

	nilStyle = lipgloss.NewStyle().
			Foreground(cli.ColorMuted).
			Italic(true)
)
