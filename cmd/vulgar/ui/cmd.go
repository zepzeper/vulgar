package ui

import (
	"github.com/spf13/cobra"
)

// RegisterCommands adds all discover subcommands to the root command
func RegisterCommands(rootCmd *cobra.Command) {
	rootCmd.AddCommand(TuiCmd())
}
