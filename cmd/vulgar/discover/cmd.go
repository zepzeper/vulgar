package discover

import (
	"github.com/spf13/cobra"
	"github.com/zepzeper/vulgar/cmd/vulgar/discover/git"
	"github.com/zepzeper/vulgar/cmd/vulgar/discover/google"
	"github.com/zepzeper/vulgar/cmd/vulgar/discover/slack"
)

// RegisterCommands adds all discover subcommands to the root command
func RegisterCommands(rootCmd *cobra.Command) {
	// Add init command
	rootCmd.AddCommand(InitCmd())

	// Add Google discovery commands
	rootCmd.AddCommand(google.DriveCmd())
	rootCmd.AddCommand(google.SheetsCmd())
	rootCmd.AddCommand(google.CalendarCmd())

	// Add communication discovery commands
	rootCmd.AddCommand(slack.Cmd())

	// Add git forge discovery commands
	rootCmd.AddCommand(git.GitHubCmd())
	rootCmd.AddCommand(git.GitLabCmd())
	rootCmd.AddCommand(git.CodebergCmd())
}
