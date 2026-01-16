package google

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/zepzeper/vulgar/internal/cli"
)

func SheetsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gsheets",
		Short: "Discover Google Sheets resources",
		Long: `Discover Google Sheets spreadsheets and their contents.

Setup:
  1. Create OAuth 2.0 Client ID (Desktop app) in Google Cloud Console
  2. Add client_id and client_secret to ~/.config/vulgar/config.toml
  3. Run 'vulgar gsheets login' to authenticate`,
	}

	cmd.AddCommand(sheetsLoginCmd())
	cmd.AddCommand(sheetsLogoutCmd())
	cmd.AddCommand(sheetsCheckCmd())
	cmd.AddCommand(sheetsDiscoverCmd())
	cmd.AddCommand(sheetsSearchCmd())
	cmd.AddCommand(sheetsRecentCmd())
	cmd.AddCommand(sheetsInfoCmd())
	cmd.AddCommand(sheetsTabsCmd())

	return cmd
}

func sheetsLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Authenticate with Google",
		Long:  `Opens a browser to authenticate with your Google account.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGoogleLogin()
		},
	}
}

func sheetsLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Log out of Google",
		Long:  `Removes stored Google authentication tokens.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGoogleLogout()
		},
	}
}

func sheetsCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Check Google authentication status",
		Long:  `Shows current authentication status and logged-in user.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGoogleCheck()
		},
	}
}

func sheetsDiscoverCmd() *cobra.Command {
	var maxResults int64
	var plain bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all spreadsheets",
		Long: `List all Google Spreadsheets you have access to.

Examples:
  vulgar gsheets list
  vulgar gsheets list --plain    # Non-interactive output`,
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := GetDriveService()
			if err != nil {
				cli.PrintError("%v", err)
				return nil
			}

			cli.PrintLoading("Fetching spreadsheets")

			query := "mimeType = 'application/vnd.google-apps.spreadsheet' and trashed = false"

			resp, err := service.Files.List().
				Q(query).
				Fields("files(id, name, modifiedTime, webViewLink)").
				OrderBy("modifiedTime desc").
				PageSize(maxResults).
				Do()

			if err != nil {
				cli.PrintFailed()
				cli.PrintError("Failed to list spreadsheets: %v", err)
				return nil
			}

			cli.PrintDone()

			if len(resp.Files) == 0 {
				cli.PrintWarning("No spreadsheets found")
				return nil
			}

			cli.PrintSuccess("Found %d spreadsheet(s)", len(resp.Files))
			fmt.Println()

			type fileOption struct {
				Name     string
				ID       string
				Modified string
			}

			options := make([]fileOption, len(resp.Files))
			for i, file := range resp.Files {
				modified := file.ModifiedTime
				if t, err := time.Parse(time.RFC3339, file.ModifiedTime); err == nil {
					modified = cli.TimeAgo(t)
				}
				options[i] = fileOption{
					Name:     file.Name,
					ID:       file.Id,
					Modified: modified,
				}
			}

			// Plain mode - just print the list
			if plain {
				for _, opt := range options {
					fmt.Printf("%s\n", opt.Name)
					fmt.Printf("  %s\n", opt.ID)
					fmt.Println()
				}
				return nil
			}

			// Interactive mode
			selected, err := cli.Select("Select a spreadsheet", options, func(f fileOption) string {
				return fmt.Sprintf("%s  %s", f.Name, cli.Muted("("+f.Modified+")"))
			})

			if err != nil {
				// User cancelled - show list
				for _, opt := range options {
					fmt.Printf("%s\n", opt.Name)
					fmt.Printf("  %s %s\n", cli.Muted("ID:"), opt.ID)
					fmt.Println()
				}
				return nil
			}

			// Show selected item details
			fmt.Println()
			cli.PrintHeader("Selected Spreadsheet")
			fmt.Println()
			fmt.Printf("  %s %s\n", cli.Muted("Name:"), selected.Name)
			fmt.Printf("  %s %s\n", cli.Muted("Modified:"), selected.Modified)
			fmt.Println()
			fmt.Printf("  %s\n", cli.Bold("ID (copy this):"))
			fmt.Printf("  %s\n", selected.ID)
			fmt.Println()
			cli.PrintTip(fmt.Sprintf("vulgar gsheets info %s", selected.ID))

			return nil
		},
	}

	cmd.Flags().Int64VarP(&maxResults, "limit", "n", 20, "Maximum number of results")
	cmd.Flags().BoolVar(&plain, "plain", false, "Plain output (non-interactive)")

	return cmd
}

func sheetsSearchCmd() *cobra.Command {
	var maxResults int64

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search for spreadsheets by name",
		Long: `Search for Google Spreadsheets by name.

Examples:
  vulgar gsheets search "budget"
  vulgar gsheets search "Q1 Report"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			searchQuery := args[0]

			service, err := GetDriveService()
			if err != nil {
				cli.PrintError("%v", err)
				return nil
			}

			cli.PrintLoading(fmt.Sprintf("Searching for \"%s\"", searchQuery))

			query := fmt.Sprintf("mimeType = 'application/vnd.google-apps.spreadsheet' and name contains '%s' and trashed = false", searchQuery)

			resp, err := service.Files.List().
				Q(query).
				Fields("files(id, name, modifiedTime, webViewLink)").
				OrderBy("modifiedTime desc").
				PageSize(maxResults).
				Do()

			if err != nil {
				cli.PrintFailed()
				cli.PrintError("Search failed: %v", err)
				return nil
			}

			cli.PrintDone()

			if len(resp.Files) == 0 {
				cli.PrintWarning("No spreadsheets found matching: %s", searchQuery)
				return nil
			}

			cli.PrintSuccess("Found %d spreadsheet(s)", len(resp.Files))
			fmt.Println()

			type fileOption struct {
				Name     string
				ID       string
				Modified string
			}

			options := make([]fileOption, len(resp.Files))
			for i, file := range resp.Files {
				modified := file.ModifiedTime
				if t, err := time.Parse(time.RFC3339, file.ModifiedTime); err == nil {
					modified = cli.TimeAgo(t)
				}
				options[i] = fileOption{
					Name:     file.Name,
					ID:       file.Id,
					Modified: modified,
				}
			}

			selected, err := cli.Select("Select a spreadsheet", options, func(f fileOption) string {
				return fmt.Sprintf("%s  %s", f.Name, cli.Muted("("+f.Modified+")"))
			})

			if err != nil {
				fmt.Println()
				for i, opt := range options {
					fmt.Printf("  %s%d%s %s\n", cli.Bold("["), i+1, cli.Bold("]"), opt.Name)
					fmt.Printf("      %s\n", cli.Muted(opt.ID))
				}
				return nil
			}

			fmt.Println()
			cli.PrintHeader("Selected Spreadsheet")
			fmt.Println()
			fmt.Printf("  %s %s\n", cli.Muted("Name:"), selected.Name)
			fmt.Printf("  %s %s\n", cli.Muted("Modified:"), selected.Modified)
			fmt.Println()
			fmt.Printf("  %s\n", cli.Bold("ID (copy this):"))
			fmt.Printf("  %s\n", selected.ID)
			fmt.Println()
			cli.PrintTip(fmt.Sprintf("vulgar gsheets info %s", selected.ID))

			return nil
		},
	}

	cmd.Flags().Int64VarP(&maxResults, "limit", "n", 20, "Maximum number of results")

	return cmd
}

func sheetsRecentCmd() *cobra.Command {
	var days int
	var maxResults int64

	cmd := &cobra.Command{
		Use:   "recent",
		Short: "List recently modified spreadsheets",
		Long: `List spreadsheets that were recently modified.

Examples:
  vulgar gsheets recent              # Modified in last 7 days
  vulgar gsheets recent --days 30    # Modified in last 30 days`,
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := GetDriveService()
			if err != nil {
				cli.PrintError("%v", err)
				return nil
			}

			cli.PrintLoading(fmt.Sprintf("Fetching spreadsheets modified in last %d days", days))

			threshold := time.Now().AddDate(0, 0, -days).Format(time.RFC3339)
			query := fmt.Sprintf("mimeType = 'application/vnd.google-apps.spreadsheet' and modifiedTime > '%s' and trashed = false", threshold)

			resp, err := service.Files.List().
				Q(query).
				Fields("files(id, name, modifiedTime, webViewLink)").
				OrderBy("modifiedTime desc").
				PageSize(maxResults).
				Do()

			if err != nil {
				cli.PrintFailed()
				cli.PrintError("Failed to list spreadsheets: %v", err)
				return nil
			}

			cli.PrintDone()

			if len(resp.Files) == 0 {
				cli.PrintWarning("No recently modified spreadsheets found")
				return nil
			}

			cli.PrintSuccess("Found %d recently modified spreadsheet(s)", len(resp.Files))
			fmt.Println()

			type fileOption struct {
				Name     string
				ID       string
				Modified string
			}

			options := make([]fileOption, len(resp.Files))
			for i, file := range resp.Files {
				modified := file.ModifiedTime
				if t, err := time.Parse(time.RFC3339, file.ModifiedTime); err == nil {
					modified = cli.TimeAgo(t)
				}
				options[i] = fileOption{
					Name:     file.Name,
					ID:       file.Id,
					Modified: modified,
				}
			}

			selected, err := cli.Select("Select a spreadsheet", options, func(f fileOption) string {
				return fmt.Sprintf("%s  %s", f.Name, cli.Muted("("+f.Modified+")"))
			})

			if err != nil {
				fmt.Println()
				for i, opt := range options {
					fmt.Printf("  %s%d%s %s\n", cli.Bold("["), i+1, cli.Bold("]"), opt.Name)
					fmt.Printf("      %s\n", cli.Muted(opt.ID))
				}
				return nil
			}

			fmt.Println()
			cli.PrintHeader("Selected Spreadsheet")
			fmt.Println()
			fmt.Printf("  %s %s\n", cli.Muted("Name:"), selected.Name)
			fmt.Printf("  %s %s\n", cli.Muted("Modified:"), selected.Modified)
			fmt.Println()
			fmt.Printf("  %s\n", cli.Bold("ID (copy this):"))
			fmt.Printf("  %s\n", selected.ID)
			fmt.Println()
			cli.PrintTip(fmt.Sprintf("vulgar gsheets info %s", selected.ID))

			return nil
		},
	}

	cmd.Flags().IntVar(&days, "days", 7, "Number of days to look back")
	cmd.Flags().Int64VarP(&maxResults, "limit", "n", 20, "Maximum number of results")

	return cmd
}

func sheetsInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info <spreadsheet_id>",
		Short: "Show spreadsheet info",
		Long: `Show information about a Google Spreadsheet including all sheets (tabs).

Examples:
  vulgar gsheets info 1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			spreadsheetID := args[0]

			service, err := GetSheetsService()
			if err != nil {
				cli.PrintError("%v", err)
				return nil
			}

			cli.PrintLoading("Fetching spreadsheet info")

			resp, err := service.Spreadsheets.Get(spreadsheetID).Do()

			if err != nil {
				cli.PrintFailed()
				cli.PrintError("Failed to get spreadsheet: %v", err)
				return nil
			}

			cli.PrintDone()

			cli.PrintHeader(resp.Properties.Title)
			fmt.Println()
			fmt.Println("  " + cli.Muted("ID:       ") + cli.Code(resp.SpreadsheetId))
			fmt.Println("  " + cli.Muted("URL:      ") + resp.SpreadsheetUrl)
			fmt.Println("  " + cli.Muted("Locale:   ") + resp.Properties.Locale)
			fmt.Println("  " + cli.Muted("Timezone: ") + resp.Properties.TimeZone)

			fmt.Println()
			cli.PrintHeader(fmt.Sprintf("Sheets (%d)", len(resp.Sheets)))
			fmt.Println()

			columns := []cli.Column{
				{Title: "Title", Width: 30},
				{Title: "Sheet ID", Width: 15},
				{Title: "Rows", Width: 10},
				{Title: "Columns", Width: 10},
			}

			var rows [][]string
			for _, sheet := range resp.Sheets {
				props := sheet.Properties
				rowCount := int64(0)
				colCount := int64(0)
				if props.GridProperties != nil {
					rowCount = props.GridProperties.RowCount
					colCount = props.GridProperties.ColumnCount
				}

				rows = append(rows, []string{
					"[S] " + props.Title,
					fmt.Sprintf("%d", props.SheetId),
					fmt.Sprintf("%d", rowCount),
					fmt.Sprintf("%d", colCount),
				})
			}

			cli.PrintTable(columns, rows)

			cli.PrintTip(fmt.Sprintf("Read data: gsheets.get_values(client, \"%s\", \"Sheet1!A1:D10\")", spreadsheetID))

			return nil
		},
	}

	return cmd
}

func sheetsTabsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tabs <spreadsheet_id>",
		Short: "List tabs in a spreadsheet",
		Long: `List all tabs (sheets) in a Google Spreadsheet with their IDs.

Examples:
  vulgar gsheets tabs 1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			spreadsheetID := args[0]

			service, err := GetSheetsService()
			if err != nil {
				cli.PrintError("%v", err)
				return nil
			}

			cli.PrintLoading("Fetching sheets")

			resp, err := service.Spreadsheets.Get(spreadsheetID).Do()

			if err != nil {
				cli.PrintFailed()
				cli.PrintError("Failed to get spreadsheet: %v", err)
				return nil
			}

			cli.PrintDone()

			cli.PrintHeader(fmt.Sprintf("Sheets in \"%s\"", resp.Properties.Title))
			fmt.Println()

			columns := []cli.Column{
				{Title: "Title", Width: 35},
				{Title: "Sheet ID", Width: 15},
				{Title: "Index", Width: 8},
			}

			var rows [][]string
			for _, sheet := range resp.Sheets {
				props := sheet.Properties
				rows = append(rows, []string{
					"[S] " + props.Title,
					fmt.Sprintf("%d", props.SheetId),
					fmt.Sprintf("%d", props.Index),
				})
			}

			cli.PrintTable(columns, rows)

			if len(resp.Sheets) > 0 {
				firstSheet := resp.Sheets[0].Properties.Title
				cli.PrintTip(fmt.Sprintf("Read: gsheets.get_values(client, \"%s\", \"%s!A1:Z100\")", spreadsheetID, firstSheet))
			}

			return nil
		},
	}

	return cmd
}
