package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Tab interface {
	// Init returns any initial commands that should be run when the tab is first created
	Init() tea.Cmd

	// Update handles messages and returns the updated model and any commands
	Update(msg tea.Msg) (Tab, tea.Cmd)

	// View renders the tab's content
	View(width, height int) string

	// Title returns the display name of the tab
	Title() string

	// Shortcut returns the keyboard shortcut key for this tab (e.g., "1", "2")
	Shortcut() string
}
