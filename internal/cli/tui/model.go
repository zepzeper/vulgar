package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zepzeper/vulgar/internal/cli"
)

// RootModel is the main model that manages all tabs and the overall TUI layout
type RootModel struct {
	tabs      []Tab
	activeTab int
	width     int
	height    int
}

// NewRootModel creates a new root model with all tabs initialized
func NewRootModel() RootModel {
	tabs := []Tab{
		NewWorkflowsTab(),
		NewSettingsTab(),
		// Add more tabs here as you create them:
		// NewAnotherTab(),
		// NewYetAnotherTab(),
	}

	// Initialize all tabs
	var cmds []tea.Cmd
	for _, tab := range tabs {
		if cmd := tab.Init(); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return RootModel{
		tabs:      tabs,
		activeTab: 0,
	}
}

func (m RootModel) Init() tea.Cmd {
	// Initialize all tabs
	var cmds []tea.Cmd
	for _, tab := range m.tabs {
		if cmd := tab.Init(); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return tea.Batch(cmds...)
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "tab":
			// Cycle to next tab
			m.activeTab = (m.activeTab + 1) % len(m.tabs)

		case "shift+tab":
			// Cycle to previous tab
			m.activeTab = (m.activeTab - 1 + len(m.tabs)) % len(m.tabs)

		default:
			// Check for numeric shortcuts (1, 2, 3, etc.)
			if len(msg.String()) == 1 {
				for i, tab := range m.tabs {
					if msg.String() == tab.Shortcut() {
						m.activeTab = i
						break
					}
				}
			}

			// Forward the message to the active tab
			updatedTab, cmd := m.tabs[m.activeTab].Update(msg)
			m.tabs[m.activeTab] = updatedTab
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Forward window size to all tabs so they can adjust their layouts
		for i, tab := range m.tabs {
			updatedTab, cmd := tab.Update(msg)
			m.tabs[i] = updatedTab
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

	default:
		// Forward all other messages to the active tab
		updatedTab, cmd := m.tabs[m.activeTab].Update(msg)
		m.tabs[m.activeTab] = updatedTab
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m RootModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	// Calculate layout dimensions
	tabBarHeight := 3
	contentHeight := m.height - tabBarHeight - 2 // Reserve space for border and padding

	// Render tab bar
	tabBar := m.renderTabBar()

	// Render active tab content
	activeTabContent := m.tabs[m.activeTab].View(m.width-4, contentHeight)

	// Combine everything with proper layout
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		tabBar,
		activeTabContent,
	)

	// Wrap in fullscreen border
	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(cli.ColorPrimary).
		Padding(0, 1).
		Render(content)
}

// renderTabBar renders the tab navigation bar at the top
func (m RootModel) renderTabBar() string {
	var tabRenders []string

	for i, tab := range m.tabs {
		tabStyle := lipgloss.NewStyle().
			Padding(0, 2).
			MarginRight(1)

		if i == m.activeTab {
			tabStyle = tabStyle.
				Foreground(cli.ColorPrimary).
				Bold(true).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(cli.ColorPrimary).
				Background(cli.ColorPrimary).
				Foreground(lipgloss.Color("#000000"))
		} else {
			tabStyle = tabStyle.
				Foreground(cli.ColorMuted).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(cli.ColorMuted)
		}

		tabLabel := "[" + tab.Shortcut() + "] " + tab.Title()
		tabRenders = append(tabRenders, tabStyle.Render(tabLabel))
	}

	// Join tabs with spacing
	tabBar := lipgloss.JoinHorizontal(lipgloss.Top, tabRenders...)

	// Add help text on the right side
	helpText := cli.Muted("Tab/Shift+Tab: Switch | Q: Quit")
	helpStyle := lipgloss.NewStyle().
		Align(lipgloss.Right).
		Foreground(cli.ColorMuted)

	// Calculate available width for help text
	helpWidth := m.width - lipgloss.Width(tabBar) - 6
	if helpWidth < 0 {
		helpWidth = 0
	}

	help := helpStyle.Width(helpWidth).Render(helpText)

	// Combine tab bar and help
	return lipgloss.JoinHorizontal(
		lipgloss.Bottom,
		tabBar,
		help,
	)
}
