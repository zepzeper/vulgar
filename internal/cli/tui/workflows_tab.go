package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zepzeper/vulgar/internal/cli"
	"github.com/zepzeper/vulgar/internal/config"
)

type WorkflowsTab struct {
	width         int
	height        int
	list          list.Model
	loading       bool
	error         string
	inspectorMode bool
	inspector     *WorkflowInspector
}

type workflowItem struct {
	workflow WorkflowInfo
}

// FilterValue implements list.Item interface for filtering
// Returns a case-insensitive searchable string containing path and filename
func (i workflowItem) FilterValue() string {
	searchable := strings.ToLower(i.workflow.FullPath)
	return searchable
}

func NewWorkflowsTab() *WorkflowsTab {
	return &WorkflowsTab{
		loading: true,
	}
}

func (t *WorkflowsTab) Init() tea.Cmd {
	return t.discoverWorkflows
}

// discoverWorkflows is a tea.Cmd that discovers workflows
func (t *WorkflowsTab) discoverWorkflows() tea.Msg {
	workflows, err := DiscoverWorkflows()
	if err != nil {
		return workflowsDiscoveredMsg{error: err.Error()}
	}
	return workflowsDiscoveredMsg{workflows: workflows}
}

// workflowsDiscoveredMsg is sent when workflow discovery completes
type workflowsDiscoveredMsg struct {
	workflows []WorkflowInfo
	error     string
}

func (t *WorkflowsTab) Update(msg tea.Msg) (Tab, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	// If in inspector mode, forward most messages to inspector
	if t.inspectorMode && t.inspector != nil {
		// Handle Esc to exit inspector mode
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "esc" {
			t.inspector.Close()
			t.inspector = nil
			t.inspectorMode = false
			return t, nil
		}

		// Forward all messages to inspector
		updatedInspector, inspectorCmd := t.inspector.Update(msg)
		t.inspector = updatedInspector
		return t, inspectorCmd
	}

	switch msg := msg.(type) {
	case workflowsDiscoveredMsg:
		t.loading = false
		if msg.error != "" {
			t.error = msg.error
			t.list = list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
		} else {
			t.error = ""
			// Convert workflows to list items
			items := make([]list.Item, len(msg.workflows))
			for i, wf := range msg.workflows {
				items[i] = workflowItem{workflow: wf}
			}

			// Create custom delegate with better styling
			delegate := newWorkflowDelegate()

			// Calculate list dimensions
			listWidth := t.width - 4
			listHeight := t.height - 8
			if listWidth < 0 {
				listWidth = 0
			}
			if listHeight < 0 {
				listHeight = 0
			}

			t.list = list.New(items, delegate, listWidth, listHeight)
			t.list.Title = ""
			t.list.SetShowStatusBar(true)
			t.list.SetFilteringEnabled(true)
			t.list.Styles.Title = lipgloss.NewStyle().
				Foreground(cli.ColorPrimary).
				Bold(true)
			t.list.Styles.NoItems = lipgloss.NewStyle().
				Foreground(cli.ColorMuted)
		}

	case tea.WindowSizeMsg:
		t.width = msg.Width
		t.height = msg.Height
		// Update list dimensions if list is initialized
		if !t.loading && t.error == "" && len(t.list.Items()) > 0 {
			listWidth := t.width - 4
			listHeight := t.height - 8
			if listWidth < 0 {
				listWidth = 0
			}
			if listHeight < 0 {
				listHeight = 0
			}
			t.list.SetWidth(listWidth)
			t.list.SetHeight(listHeight)
		}

	case tea.KeyMsg:
		// Only process keys if we have a valid list
		if !t.loading && t.error == "" {
			// Don't match any of the keys below if we're actively filtering.
			// This allows the list to handle all input when filtering.
			if t.list.FilterState() == list.Filtering {
				// Break out to let list handle filtering below
				break
			}

			// Handle custom keys only when NOT filtering
			switch msg.String() {
			case "r":
				// Refresh workflows (only when not filtering)
				t.loading = true
				t.error = ""
				return t, t.discoverWorkflows
			case "enter":
				// Open inspector for selected workflow
				if selectedItem := t.list.SelectedItem(); selectedItem != nil {
					if wfItem, ok := selectedItem.(workflowItem); ok {
						t.inspector = NewWorkflowInspector(wfItem.workflow.FullPath)
						t.inspectorMode = true
						return t, t.inspector.Init()
					}
				}
			}
		}
	}

	// This will also call our delegate's update function.
	// Always update the list - this handles filtering, navigation, and all list interactions
	if !t.loading && t.error == "" {
		newListModel, listCmd := t.list.Update(msg)
		t.list = newListModel
		if listCmd != nil {
			cmds = append(cmds, listCmd)
		}
	}

	if len(cmds) > 0 {
		cmd = tea.Batch(cmds...)
	}

	return t, cmd
}

func (t *WorkflowsTab) View(width, height int) string {
	t.width = width
	t.height = height

	contentWidth := width - 4
	contentHeight := height - 6

	if contentWidth < 0 {
		contentWidth = 0
	}
	if contentHeight < 0 {
		contentHeight = 0
	}

	// If in inspector mode, delegate to inspector
	if t.inspectorMode && t.inspector != nil {
		return t.inspector.View(width, height)
	}

	var content string

	if t.loading {
		content = t.renderLoading()
	} else if t.error != "" {
		content = t.renderError()
	} else if t.list.Items() == nil || len(t.list.Items()) == 0 {
		content = t.renderEmpty()
	} else {
		content = t.renderWorkflowsList()
	}

	return lipgloss.NewStyle().
		Width(contentWidth).
		Height(contentHeight).
		Padding(1, 2).
		Render(content)
}

func (t *WorkflowsTab) renderLoading() string {
	return cli.Title("Workflows") + "\n\n" +
		cli.Info("Discovering workflows...")
}

func (t *WorkflowsTab) renderError() string {
	return cli.Title("Workflows") + "\n\n" +
		cli.Error("Error: "+t.error) + "\n\n" +
		cli.Muted("Press 'r' to retry, Tab to switch tabs, Ctrl+C or Q to quit")
}

func (t *WorkflowsTab) renderEmpty() string {
	workflowsPath := config.GetWorkflowsPath()
	return cli.Title("Workflows") + "\n\n" +
		cli.Muted(fmt.Sprintf("No workflows found in: %s", workflowsPath)) + "\n\n" +
		cli.Info("Press 'r' to refresh, Tab to switch tabs, Ctrl+C or Q to quit")
}

func (t *WorkflowsTab) renderWorkflowsList() string {
	workflowsPath := config.GetWorkflowsPath()

	// Get filtering information
	isFiltering := t.list.FilterState() == list.Filtering
	isFiltered := t.list.IsFiltered()
	totalItems := len(t.list.Items())
	filteredItems := len(t.list.VisibleItems())

	// Build header with filtering info
	var header string
	if isFiltering || isFiltered {
		// Show filtered count when filtering or filtered
		if isFiltered && filteredItems != totalItems {
			header = cli.Title("Workflows") + " " + cli.Info("(filtered)") + "\n" +
				cli.Muted(fmt.Sprintf("Showing %d of %d workflow(s) in: %s", filteredItems, totalItems, workflowsPath)) + "\n"
		} else {
			header = cli.Title("Workflows") + " " + cli.Info("(filtering)") + "\n" +
				cli.Muted(fmt.Sprintf("Found %d workflow(s) in: %s", totalItems, workflowsPath)) + "\n"
		}
	} else {
		// Normal state
		header = cli.Title("Workflows") + "\n" +
			cli.Muted(fmt.Sprintf("Found %d workflow(s) in: %s", totalItems, workflowsPath)) + "\n"
	}

	listView := t.list.View()

	// Update help text based on filter state
	var helpText string
	if isFiltering {
		helpText = "\n" + cli.Muted("Type to filter | Esc: Clear filter | Tab: Switch tabs | Q: Quit")
	} else if isFiltered {
		helpText = "\n" + cli.Muted("↑/↓: Navigate | Enter: Select | /: Filter again | Esc: Clear filter | r: Refresh | Tab: Switch tabs | Q: Quit")
	} else {
		helpText = "\n" + cli.Muted("↑/↓: Navigate | Enter: Select | /: Filter | r: Refresh | Tab: Switch tabs | Q: Quit")
	}

	return header + "\n" + listView + helpText
}

// workflowDelegate provides custom rendering for workflow items
type workflowDelegate struct {
	list.DefaultDelegate
}

func newWorkflowDelegate() list.ItemDelegate {
	d := list.NewDefaultDelegate()

	// Style for selected items
	d.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(cli.ColorPrimary).
		Bold(true).
		PaddingLeft(0)
	d.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(cli.ColorMuted).
		PaddingLeft(2)

	// Style for unselected items
	d.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		PaddingLeft(0)
	d.Styles.NormalDesc = lipgloss.NewStyle().
		Foreground(cli.ColorMuted).
		PaddingLeft(2)

	return &workflowDelegate{DefaultDelegate: d}
}

// Height returns the height of each item (2 lines: path + full path)
func (d *workflowDelegate) Height() int {
	return 2
}

// Spacing returns spacing between items
func (d *workflowDelegate) Spacing() int {
	return 0
}

// Render renders a workflow item
func (d *workflowDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	workflowItem, ok := item.(workflowItem)
	if !ok {
		d.DefaultDelegate.Render(w, m, index, item)
		return
	}

	wf := workflowItem.workflow
	isSelected := index == m.Index()

	var titleStyle, descStyle lipgloss.Style
	if isSelected {
		titleStyle = d.Styles.SelectedTitle
		descStyle = d.Styles.SelectedDesc
	} else {
		titleStyle = d.Styles.NormalTitle
		descStyle = d.Styles.NormalDesc
	}

	// Get available width from list model
	width := m.Width()
	if width <= 0 {
		width = 80 // fallback
	}

	// Truncate path if too long
	path := wf.Path
	if len(path) > width-4 {
		path = path[:width-7] + "..."
	}

	// Truncate full path if too long
	fullPath := wf.FullPath
	if len(fullPath) > width-6 {
		fullPath = "..." + fullPath[len(fullPath)-(width-9):]
	}

	title := titleStyle.Width(width - 2).Render(path)
	desc := descStyle.Width(width - 2).Render("└─ " + fullPath)

	rendered := lipgloss.JoinVertical(lipgloss.Left, title, desc)
	fmt.Fprint(w, rendered)
}

func (t *WorkflowsTab) Title() string {
	return "Workflows"
}

func (t *WorkflowsTab) Shortcut() string {
	return "1"
}
