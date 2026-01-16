package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zepzeper/vulgar/internal/cli"
)

// SettingsTab displays and manages application settings
type SettingsTab struct {
	width  int
	height int
	// Add your settings-specific state here
	// e.g., config map[string]interface{}, selectedSetting int, etc.
}

// NewSettingsTab creates a new SettingsTab instance
func NewSettingsTab() *SettingsTab {
	return &SettingsTab{}
}

func (t *SettingsTab) Init() tea.Cmd {
	// Initialize any async operations here
	// e.g., loading settings from config file, etc.
	return nil
}

func (t *SettingsTab) Update(msg tea.Msg) (Tab, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// Handle tab-specific keyboard shortcuts here
		// e.g., arrow keys for navigation, enter to edit, etc.
		}

	case tea.WindowSizeMsg:
		t.width = msg.Width
		t.height = msg.Height
	}

	return t, cmd
}

func (t *SettingsTab) View(width, height int) string {
	// Store dimensions for use in Update
	t.width = width
	t.height = height

	// Calculate available space (accounting for padding)
	contentWidth := width - 4
	contentHeight := height - 6 // Reserve space for header and padding

	if contentWidth < 0 {
		contentWidth = 0
	}
	if contentHeight < 0 {
		contentHeight = 0
	}

	// Build the content
	content := lipgloss.NewStyle().
		Width(contentWidth).
		Height(contentHeight).
		Padding(1, 2).
		Render(
			cli.Title("Settings") + "\n\n" +
				cli.Muted("This is the settings tab. Add your settings management UI here.") + "\n\n" +
				cli.Info("Press Tab to switch tabs, Ctrl+C or Q to quit"),
		)

	return content
}

func (t *SettingsTab) Title() string {
	return "Settings"
}

func (t *SettingsTab) Shortcut() string {
	return "2"
}

