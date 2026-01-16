package slack

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zepzeper/vulgar/internal/cli"
)

func channelsCmd() *cobra.Command {
	var showPrivate bool
	var showArchived bool
	var plain bool

	cmd := &cobra.Command{
		Use:   "channels",
		Short: "List Slack channels",
		Long: `List all Slack channels accessible to the bot.

Required Slack OAuth scopes:
  - channels:read (for public channels)
  - groups:read (for private channels, with --private flag)

Examples:
  vulgar slack channels
  vulgar slack channels --private
  vulgar slack channels --archived
  vulgar slack channels --plain    # Non-interactive output`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				cli.PrintError("%v", err)
				return nil
			}

			cli.PrintLoading("Fetching channels")

			channels, err := client.ListChannels(context.Background(), showPrivate, showArchived)

			if err != nil {
				cli.PrintFailed()
				if strings.Contains(err.Error(), "missing_scope") {
					cli.PrintError("Missing required Slack OAuth scopes")
					fmt.Println()
					fmt.Println("  Your bot needs these scopes:")
					fmt.Println("    - " + cli.Code("channels:read") + " - List public channels")
					if showPrivate {
						fmt.Println("    - " + cli.Code("groups:read") + " - List private channels")
					}
					fmt.Println()
					fmt.Println("  Add scopes at: https://api.slack.com/apps -> Your App -> OAuth & Permissions")
					return nil
				}
				cli.PrintError("%v", err)
				return nil
			}

		cli.PrintDone()

		if len(channels) == 0 {
			cli.PrintWarning("No channels found")
			return nil
		}

		cli.PrintSuccess("Found %d channel(s)", len(channels))
		fmt.Println()

		type channelOption struct {
			Name      string
			ID        string
			Type      string
			Members   int
			IsPrivate bool
		}

		options := make([]channelOption, len(channels))
		for i, ch := range channels {
			typeStr := "public"
			if ch.IsPrivate {
				typeStr = "private"
			}
			options[i] = channelOption{
				Name:      ch.Name,
				ID:        ch.ID,
				Type:      typeStr,
				Members:   ch.NumMembers,
				IsPrivate: ch.IsPrivate,
			}
		}

		// Plain mode - just print the list
		if plain {
			for _, opt := range options {
				prefix := "[P]"
				if opt.IsPrivate {
					prefix = "[L]"
				}
				fmt.Printf("%s #%s\n", prefix, opt.Name)
				fmt.Printf("  %s\n", opt.ID)
				fmt.Printf("  %s, %d members\n", opt.Type, opt.Members)
				fmt.Println()
			}
			return nil
		}

		// Interactive mode
		selected, err := cli.Select("Select a channel", options, func(c channelOption) string {
			prefix := "[P]"
			if c.IsPrivate {
				prefix = "[L]"
			}
			return fmt.Sprintf("%s #%s  %s (%d members)", prefix, c.Name, cli.Muted(c.Type), c.Members)
		})

		if err != nil {
			// User cancelled - show list
			for _, opt := range options {
				prefix := "[P]"
				if opt.IsPrivate {
					prefix = "[L]"
				}
				fmt.Printf("%s #%s\n", prefix, opt.Name)
				fmt.Printf("  %s %s\n", cli.Muted("ID:"), opt.ID)
				fmt.Println()
			}
			return nil
		}

		// Show selected item details
		fmt.Println()
		cli.PrintHeader("Selected Channel")
		fmt.Println()
		prefix := "[P]"
		if selected.IsPrivate {
			prefix = "[L]"
		}
		fmt.Printf("  %s %s#%s\n", cli.Muted("Name:"), prefix, selected.Name)
		fmt.Printf("  %s %s\n", cli.Muted("Type:"), selected.Type)
		fmt.Printf("  %s %d\n", cli.Muted("Members:"), selected.Members)
		fmt.Println()
		fmt.Printf("  %s\n", cli.Bold("ID (copy this):"))
		fmt.Printf("  %s\n", selected.ID)
		fmt.Println()
		cli.PrintTip(fmt.Sprintf("Use in workflow: slack.send(client, \"#%s\", \"Hello!\")", selected.Name))

		return nil
	},
}

cmd.Flags().BoolVar(&showPrivate, "private", false, "Include private channels")
cmd.Flags().BoolVar(&showArchived, "archived", false, "Include archived channels")
cmd.Flags().BoolVar(&plain, "plain", false, "Plain output (non-interactive)")

return cmd
}

func findChannelCmd() *cobra.Command {
	var plain bool

	cmd := &cobra.Command{
		Use:   "find-channel <name>",
		Short: "Find a channel by name",
		Long: `Find a Slack channel by name.

Examples:
  vulgar slack find-channel general
  vulgar slack find-channel engineering
  vulgar slack find-channel general --plain    # Non-interactive output`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			searchTerm := strings.ToLower(args[0])

			client, err := getClient()
			if err != nil {
				cli.PrintError("%v", err)
				return nil
			}

			cli.PrintLoading(fmt.Sprintf("Searching for \"%s\"", args[0]))

			channels, err := client.ListChannels(context.Background(), true, false)

			if err != nil {
				cli.PrintFailed()
				cli.PrintError("%v", err)
				return nil
			}

			cli.PrintDone()

			var matches []struct {
				Name      string
				ID        string
				IsPrivate bool
				Purpose   string
			}

			for _, ch := range channels {
				name := strings.ToLower(ch.Name)
				purpose := strings.ToLower(ch.Purpose.Value)

					if strings.Contains(name, searchTerm) || strings.Contains(purpose, searchTerm) {
					matches = append(matches, struct {
						Name      string
						ID        string
						IsPrivate bool
						Purpose   string
					}{
						Name:      ch.Name,
						ID:        ch.ID,
						IsPrivate: ch.IsPrivate,
						Purpose:   ch.Purpose.Value,
					})
				}
			}

		if len(matches) == 0 {
			cli.PrintWarning("No channels found matching: %s", args[0])
			return nil
		}

		cli.PrintSuccess("Found %d channel(s) matching \"%s\"", len(matches), args[0])
		fmt.Println()

		type channelOption struct {
			Name      string
			ID        string
			Type      string
			Purpose   string
			IsPrivate bool
		}

		options := make([]channelOption, len(matches))
		for i, ch := range matches {
			typeStr := "public"
			if ch.IsPrivate {
				typeStr = "private"
			}
			options[i] = channelOption{
				Name:      ch.Name,
				ID:        ch.ID,
				Type:      typeStr,
				Purpose:   ch.Purpose,
				IsPrivate: ch.IsPrivate,
			}
		}

		// Plain mode
		if plain {
			for _, opt := range options {
				prefix := "[P]"
				if opt.IsPrivate {
					prefix = "[L]"
				}
				fmt.Printf("%s #%s\n", prefix, opt.Name)
				fmt.Printf("  %s\n", opt.ID)
				if opt.Purpose != "" {
					fmt.Printf("  %s\n", opt.Purpose)
				}
				fmt.Println()
			}
			return nil
		}

		// Interactive mode
		selected, err := cli.Select("Select a channel", options, func(c channelOption) string {
			prefix := "[P]"
			if c.IsPrivate {
				prefix = "[L]"
			}
			purpose := c.Purpose
			if purpose != "" {
				purpose = cli.Truncate(purpose, 40)
			} else {
				purpose = "No purpose"
			}
			return fmt.Sprintf("%s #%s  %s", prefix, c.Name, cli.Muted("("+purpose+")"))
		})

		if err != nil {
			// User cancelled - show list
			for _, opt := range options {
				prefix := "[P]"
				if opt.IsPrivate {
					prefix = "[L]"
				}
				fmt.Printf("%s #%s\n", prefix, opt.Name)
				fmt.Printf("  %s %s\n", cli.Muted("ID:"), opt.ID)
				fmt.Println()
			}
			return nil
		}

		// Show selected item details
		fmt.Println()
		cli.PrintHeader("Selected Channel")
		fmt.Println()
		prefix := "[P]"
		if selected.IsPrivate {
			prefix = "[L]"
		}
		fmt.Printf("  %s %s#%s\n", cli.Muted("Name:"), prefix, selected.Name)
		fmt.Printf("  %s %s\n", cli.Muted("Type:"), selected.Type)
		if selected.Purpose != "" {
			fmt.Printf("  %s %s\n", cli.Muted("Purpose:"), selected.Purpose)
		}
		fmt.Println()
		fmt.Printf("  %s\n", cli.Bold("ID (copy this):"))
		fmt.Printf("  %s\n", selected.ID)
		fmt.Println()
		cli.PrintTip(fmt.Sprintf("Use in workflow: slack.send(client, \"#%s\", \"Hello!\")", selected.Name))

		return nil
	},
}

cmd.Flags().BoolVar(&plain, "plain", false, "Plain output (non-interactive)")

return cmd
}
