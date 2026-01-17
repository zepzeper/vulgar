package slack

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zepzeper/vulgar/internal/cli"
	"github.com/zepzeper/vulgar/internal/config"
)

func checkCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Check Slack token and show workspace info",
		Long: `Verify that your Slack token is valid and show workspace information.

This command tests authentication and displays:
  - Bot name and ID
  - Workspace name
  - Team ID`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				cli.PrintError("Slack token not configured")
				fmt.Println()
				fmt.Println("  Run: vulgar init")
				fmt.Println("  Then edit: " + config.ConfigPath())
				fmt.Println()
				fmt.Println("  Set token to your Slack bot token (xoxb-...)")
				return nil
			}

			cli.PrintLoading("Checking token")

			result, err := client.AuthTest(context.Background())

			if err != nil {
				cli.PrintFailed()
				cli.PrintError("Token verification failed: %v", err)
				return nil
			}

			cli.PrintDone()
			cli.PrintSuccess("Token is valid!")
			fmt.Println()

			cli.PrintHeader("Slack Connection")
			fmt.Println()

			botName := ""
			if v, ok := result["user"].(string); ok {
				botName = v
			}
			botID := ""
			if v, ok := result["user_id"].(string); ok {
				botID = v
			}
			teamName := ""
			if v, ok := result["team"].(string); ok {
				teamName = v
			}
			teamID := ""
			if v, ok := result["team_id"].(string); ok {
				teamID = v
			}

			fmt.Printf("  Bot:       %s (%s)\n", botName, botID)
			fmt.Printf("  Workspace: %s (%s)\n", teamName, teamID)

			if defaultChannel := client.DefaultChannel(); defaultChannel != "" {
				fmt.Printf("  Default:   %s\n", defaultChannel)
			}

			fmt.Println()
			cli.PrintTip("Run 'vulgar slack channels' to see available channels")

			return nil
		},
	}
}

func infoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show Slack workspace information",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				cli.PrintError("%v", err)
				return nil
			}

			cli.PrintLoading("Fetching workspace info")

			result, err := client.AuthTest(context.Background())

			if err != nil {
				cli.PrintFailed()
				cli.PrintError("%v", err)
				return nil
			}

			cli.PrintDone()

			cli.PrintHeader("Workspace Info")
			fmt.Println()

			for key, value := range result {
				if key != "ok" {
					fmt.Printf("  %s: %v\n", key, value)
				}
			}

			return nil
		},
	}
}
