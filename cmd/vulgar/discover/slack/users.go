package slack

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zepzeper/vulgar/internal/cli"
)

func usersCmd() *cobra.Command {
	var showBots bool
	var plain bool

	cmd := &cobra.Command{
		Use:   "users",
		Short: "List Slack users",
		Long: `List all Slack users in the workspace.

Required Slack OAuth scopes:
  - users:read

Examples:
  vulgar slack users
  vulgar slack users --bots
  vulgar slack users --plain    # Non-interactive output`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				cli.PrintError("%v", err)
				return nil
			}

			cli.PrintLoading("Fetching users")

			users, err := client.ListUsers(context.Background(), showBots)

			if err != nil {
				cli.PrintFailed()
				if strings.Contains(err.Error(), "missing_scope") {
					cli.PrintError("Missing required Slack OAuth scope")
					fmt.Println()
					fmt.Println("  Your bot needs this scope:")
					fmt.Println("    - " + cli.Code("users:read") + " - View users in workspace")
					fmt.Println()
					fmt.Println("  Add scopes at: https://api.slack.com/apps -> Your App -> OAuth & Permissions")
					return nil
				}
				cli.PrintError("%v", err)
				return nil
			}

		cli.PrintDone()

		if len(users) == 0 {
			cli.PrintWarning("No users found")
			return nil
		}

		cli.PrintSuccess("Found %d user(s)", len(users))
		fmt.Println()

		type userOption struct {
			Name     string
			ID       string
			RealName string
			Status   string
			IsBot    bool
		}

		options := make([]userOption, len(users))
		for i, u := range users {
			status := "active"
			if u.IsBot {
				status = "bot"
			}
			if u.Profile.StatusText != "" {
				status = u.Profile.StatusText
			}
			options[i] = userOption{
				Name:     u.Name,
				ID:       u.ID,
				RealName: u.RealName,
				Status:   status,
				IsBot:    u.IsBot,
			}
		}

		// Plain mode
		if plain {
			for _, opt := range options {
				prefix := "[U]"
				if opt.IsBot {
					prefix = "[B]"
				}
				fmt.Printf("%s @%s\n", prefix, opt.Name)
				fmt.Printf("  %s\n", opt.ID)
				if opt.RealName != "" {
					fmt.Printf("  %s\n", opt.RealName)
				}
				fmt.Println()
			}
			return nil
		}

		// Interactive mode
		selected, err := cli.Select("Select a user", options, func(u userOption) string {
			prefix := "[U]"
			if u.IsBot {
				prefix = "[B]"
			}
			display := fmt.Sprintf("%s @%s", prefix, u.Name)
			if u.RealName != "" {
				display += fmt.Sprintf("  %s", cli.Muted("("+u.RealName+")"))
			}
			return display
		})

		if err != nil {
			// User cancelled - show list
			for _, opt := range options {
				prefix := "[U]"
				if opt.IsBot {
					prefix = "[B]"
				}
				fmt.Printf("%s @%s\n", prefix, opt.Name)
				fmt.Printf("  %s %s\n", cli.Muted("ID:"), opt.ID)
				fmt.Println()
			}
			return nil
		}

		// Show selected item details
		fmt.Println()
		cli.PrintHeader("Selected User")
		fmt.Println()
		prefix := "[U]"
		if selected.IsBot {
			prefix = "[B]"
		}
		fmt.Printf("  %s %s@%s\n", cli.Muted("Name:"), prefix, selected.Name)
		if selected.RealName != "" {
			fmt.Printf("  %s %s\n", cli.Muted("Real Name:"), selected.RealName)
		}
		if selected.Status != "" && selected.Status != "active" && selected.Status != "bot" {
			fmt.Printf("  %s %s\n", cli.Muted("Status:"), selected.Status)
		}
		fmt.Println()
		fmt.Printf("  %s\n", cli.Bold("ID (copy this):"))
		fmt.Printf("  %s\n", selected.ID)
		fmt.Println()
		cli.PrintTip(fmt.Sprintf("Use in workflow: slack.get_user(client, \"%s\")", selected.ID))

		return nil
	},
}

cmd.Flags().BoolVar(&showBots, "bots", false, "Include bot users")
cmd.Flags().BoolVar(&plain, "plain", false, "Plain output (non-interactive)")

return cmd
}

func findUserCmd() *cobra.Command {
	var plain bool

	cmd := &cobra.Command{
		Use:   "find-user <name>",
		Short: "Find a user by name",
		Long: `Find a Slack user by name or email.

Examples:
  vulgar slack find-user john
  vulgar slack find-user "John Doe"
  vulgar slack find-user john@example.com
  vulgar slack find-user john --plain    # Non-interactive output`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			searchTerm := strings.ToLower(args[0])

			client, err := getClient()
			if err != nil {
				cli.PrintError("%v", err)
				return nil
			}

			cli.PrintLoading(fmt.Sprintf("Searching for \"%s\"", args[0]))

			users, err := client.ListUsers(context.Background(), true)

			if err != nil {
				cli.PrintFailed()
				cli.PrintError("%v", err)
				return nil
			}

			cli.PrintDone()

			var matches []struct {
				Name     string
				ID       string
				RealName string
				Email    string
				IsBot    bool
			}

			for _, u := range users {
				name := strings.ToLower(u.Name)
				realName := strings.ToLower(u.RealName)
				email := strings.ToLower(u.Email)

					if strings.Contains(name, searchTerm) ||
						strings.Contains(realName, searchTerm) ||
						strings.Contains(email, searchTerm) {
					matches = append(matches, struct {
						Name     string
						ID       string
						RealName string
						Email    string
						IsBot    bool
					}{
						Name:     u.Name,
						ID:       u.ID,
						RealName: u.RealName,
						Email:    u.Email,
						IsBot:    u.IsBot,
					})
				}
			}

		if len(matches) == 0 {
			cli.PrintWarning("No users found matching: %s", args[0])
			return nil
		}

		cli.PrintSuccess("Found %d user(s) matching \"%s\"", len(matches), args[0])
		fmt.Println()

		type userOption struct {
			Name     string
			ID       string
			RealName string
			Email    string
			IsBot    bool
		}

		options := make([]userOption, len(matches))
		for i, u := range matches {
			options[i] = userOption{
				Name:     u.Name,
				ID:       u.ID,
				RealName: u.RealName,
				Email:    u.Email,
				IsBot:    u.IsBot,
			}
		}

		// Plain mode
		if plain {
			for _, opt := range options {
				prefix := "[U]"
				if opt.IsBot {
					prefix = "[B]"
				}
				fmt.Printf("%s @%s\n", prefix, opt.Name)
				fmt.Printf("  %s\n", opt.ID)
				if opt.RealName != "" {
					fmt.Printf("  %s\n", opt.RealName)
				}
				if opt.Email != "" {
					fmt.Printf("  %s\n", opt.Email)
				}
				fmt.Println()
			}
			return nil
		}

		// Interactive mode
		selected, err := cli.Select("Select a user", options, func(u userOption) string {
			prefix := "[U]"
			if u.IsBot {
				prefix = "[B]"
			}
			display := fmt.Sprintf("%s @%s", prefix, u.Name)
			if u.RealName != "" {
				display += fmt.Sprintf("  %s", cli.Muted("("+u.RealName+")"))
			}
			return display
		})

		if err != nil {
			// User cancelled - show list
			for _, opt := range options {
				prefix := "[U]"
				if opt.IsBot {
					prefix = "[B]"
				}
				fmt.Printf("%s @%s\n", prefix, opt.Name)
				fmt.Printf("  %s %s\n", cli.Muted("ID:"), opt.ID)
				fmt.Println()
			}
			return nil
		}

		// Show selected item details
		fmt.Println()
		cli.PrintHeader("Selected User")
		fmt.Println()
		prefix := "[U]"
		if selected.IsBot {
			prefix = "[B]"
		}
		fmt.Printf("  %s %s@%s\n", cli.Muted("Name:"), prefix, selected.Name)
		if selected.RealName != "" {
			fmt.Printf("  %s %s\n", cli.Muted("Real Name:"), selected.RealName)
		}
		if selected.Email != "" {
			fmt.Printf("  %s %s\n", cli.Muted("Email:"), selected.Email)
		}
		fmt.Println()
		fmt.Printf("  %s\n", cli.Bold("ID (copy this):"))
		fmt.Printf("  %s\n", selected.ID)
		fmt.Println()
		cli.PrintTip(fmt.Sprintf("Use in workflow: slack.get_user(client, \"%s\")", selected.ID))

		return nil
	},
}

cmd.Flags().BoolVar(&plain, "plain", false, "Plain output (non-interactive)")

return cmd
}
