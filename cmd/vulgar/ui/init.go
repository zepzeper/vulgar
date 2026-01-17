package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/zepzeper/vulgar/internal/cli/tui"
)

func TuiCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tui",
		Short: "Vizually explore workflows",
		RunE:  runTui,
	}

	cmd.Flags().Bool("force", false, "Overwrite existing config file")

	return cmd
}

func runTui(cmd *cobra.Command, args []string) error {
	initialModel := tui.NewRootModel()
	p := tea.NewProgram(initialModel, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
