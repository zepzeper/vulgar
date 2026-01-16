package google

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/zepzeper/vulgar/internal/cli"
	"github.com/zepzeper/vulgar/internal/config"
	googleauth "github.com/zepzeper/vulgar/internal/services/google"
)

func runGoogleLogin() error {
	_, _, ok := config.GetGoogleOAuthCredentials()
	if !ok {
		cli.PrintError("Google OAuth credentials not configured")
		fmt.Println()
		fmt.Println("  1. Create OAuth 2.0 Client ID (Desktop app) at:")
		fmt.Println("     https://console.cloud.google.com/apis/credentials")
		fmt.Println()
		fmt.Println("  2. Edit " + config.ConfigPath())
		fmt.Println("     [google]")
		fmt.Println("     client_id = \"your-client-id.apps.googleusercontent.com\"")
		fmt.Println("     client_secret = \"your-client-secret\"")
		fmt.Println()
		fmt.Println("  3. Run 'vulgar gdrive login' again")
		return nil
	}

	cli.PrintInfo("Opening browser for Google authentication...")
	fmt.Println()

	ctx := context.Background()
	_, err := googleauth.Login(ctx)
	if err != nil {
		cli.PrintError("Login failed: %v", err)
		return nil
	}

	userInfo, err := googleauth.GetUserInfo(ctx)
	if err != nil {
		cli.PrintSuccess("Logged in successfully!")
		return nil
	}

	cli.PrintSuccess("Logged in as %s", userInfo.Email)
	fmt.Println()
	cli.PrintTip("Try 'vulgar gdrive list' to see your files")

	return nil
}

func runGoogleLogout() error {
	if !googleauth.IsLoggedIn() {
		cli.PrintWarning("Not logged in")
		return nil
	}

	if err := googleauth.Logout(); err != nil {
		cli.PrintError("Logout failed: %v", err)
		return nil
	}

	cli.PrintSuccess("Logged out successfully")
	return nil
}

func runGoogleCheck() error {
	_, _, ok := config.GetGoogleOAuthCredentials()
	if !ok {
		cli.PrintError("Google OAuth credentials not configured")
		fmt.Println()
		fmt.Println("  Run: vulgar init")
		fmt.Println("  Then edit: " + config.ConfigPath())
		fmt.Println()
		fmt.Println("  [google]")
		fmt.Println("  client_id = \"your-client-id.apps.googleusercontent.com\"")
		fmt.Println("  client_secret = \"your-client-secret\"")
		return nil
	}

	cli.PrintSuccess("OAuth credentials configured")
	fmt.Println()

	if !googleauth.IsLoggedIn() {
		cli.PrintWarning("Not logged in")
		fmt.Println()
		cli.PrintTip("Run 'vulgar gdrive login' to authenticate")
		return nil
	}

	ctx := context.Background()
	userInfo, err := googleauth.GetUserInfo(ctx)
	if err != nil {
		cli.PrintWarning("Token may be expired: %v", err)
		fmt.Println()
		cli.PrintTip("Run 'vulgar gdrive login' to re-authenticate")
		return nil
	}

	cli.PrintHeader("Google Account")
	fmt.Println()
	cli.PrintKeyValue("Email", userInfo.Email)
	if userInfo.Name != "" {
		cli.PrintKeyValue("Name", userInfo.Name)
	}
	cli.PrintKeyValue("Token", config.GoogleTokenPath())
	fmt.Println()
	cli.PrintSuccess("Ready to use Google APIs")

	return nil
}

func DriveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gdrive",
		Short: "Discover Google Drive resources",
		Long: `Discover Google Drive files and folders.

Setup:
  1. Create OAuth 2.0 Client ID (Desktop app) in Google Cloud Console
  2. Add client_id and client_secret to ~/.config/vulgar/config.toml
  3. Run 'vulgar gdrive login' to authenticate`,
	}

	cmd.AddCommand(driveLoginCmd())
	cmd.AddCommand(driveLogoutCmd())
	cmd.AddCommand(driveCheckCmd())
	cmd.AddCommand(driveListCmd())
	cmd.AddCommand(driveFindCmd())
	cmd.AddCommand(drivePathCmd())
	cmd.AddCommand(driveFoldersCmd())
	cmd.AddCommand(driveRecentCmd())
	cmd.AddCommand(driveStarredCmd())

	return cmd
}

func driveLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Authenticate with Google",
		Long:  `Opens a browser to authenticate with your Google account.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGoogleLogin()
		},
	}
}

func driveLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Log out of Google",
		Long:  `Removes stored Google authentication tokens.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGoogleLogout()
		},
	}
}

func driveCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Check Google authentication status",
		Long:  `Shows current authentication status and logged-in user.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGoogleCheck()
		},
	}
}

// mimeTypeMap maps friendly type names to Google Drive mimeTypes
var mimeTypeMap = map[string]string{
	"spreadsheet":  "application/vnd.google-apps.spreadsheet",
	"document":     "application/vnd.google-apps.document",
	"presentation": "application/vnd.google-apps.presentation",
	"folder":       "application/vnd.google-apps.folder",
	"pdf":          "application/pdf",
	"form":         "application/vnd.google-apps.form",
	"drawing":      "application/vnd.google-apps.drawing",
}

func driveListCmd() *cobra.Command {
	var folderName string
	var folderID string
	var maxResults int64
	var showTrashed bool
	var fileType string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List files in Google Drive",
		Long: `List files in Google Drive, optionally filtered by folder or type.

Examples:
  vulgar gdrive list                           # List recent files
  vulgar gdrive list --folder "Reports"        # List files in folder by name
  vulgar gdrive list --folder-id 1abc123xyz    # List files in folder by ID
  vulgar gdrive list --type spreadsheet        # List only spreadsheets
  vulgar gdrive list --type document           # List only documents

Available types: spreadsheet, document, presentation, folder, pdf, form, drawing, image, video`,
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := GetDriveService()
			if err != nil {
				cli.PrintError("%v", err)
				return nil
			}

			cli.PrintLoading("Fetching files")

			query := "trashed = " + fmt.Sprintf("%v", showTrashed)

			// Add type filter
			if fileType != "" {
				if mimeType, ok := mimeTypeMap[fileType]; ok {
					query += fmt.Sprintf(" and mimeType = '%s'", mimeType)
				} else if fileType == "image" {
					query += " and mimeType contains 'image/'"
				} else if fileType == "video" {
					query += " and mimeType contains 'video/'"
				} else {
					cli.PrintFailed()
					cli.PrintError("Unknown type: %s. Available: spreadsheet, document, presentation, folder, pdf, form, drawing, image, video", fileType)
					return nil
				}
			}

			if folderName != "" {
				folderQuery := fmt.Sprintf("name = '%s' and mimeType = 'application/vnd.google-apps.folder' and trashed = false", folderName)
				folders, err := service.Files.List().Q(folderQuery).Fields("files(id, name)").PageSize(1).Do()
				if err != nil {
					cli.PrintFailed()
					cli.PrintError("Failed to find folder: %v", err)
					return nil
				}
				if len(folders.Files) == 0 {
					cli.PrintFailed()
					cli.PrintError("Folder not found: %s", folderName)
					return nil
				}
				folderID = folders.Files[0].Id
			}

			if folderID != "" {
				query += fmt.Sprintf(" and '%s' in parents", folderID)
			}

			resp, err := service.Files.List().
				Q(query).
				Fields("files(id, name, mimeType, size, modifiedTime, webViewLink)").
				OrderBy("modifiedTime desc").
				PageSize(maxResults).
				Do()

			if err != nil {
				cli.PrintFailed()
				cli.PrintError("Failed to list files: %v", err)
				return nil
			}

			cli.PrintDone()

			if len(resp.Files) == 0 {
				cli.PrintWarning("No files found")
				return nil
			}

			if folderName != "" {
				cli.PrintHeader(fmt.Sprintf("Files in \"%s\"", folderName))
				fmt.Println(cli.Muted(fmt.Sprintf("  Folder ID: %s", folderID)))
			} else if folderID != "" {
				cli.PrintHeader("Files in folder")
				fmt.Println(cli.Muted(fmt.Sprintf("  Folder ID: %s", folderID)))
			} else if fileType != "" {
				// Capitalize first letter
				typeName := strings.ToUpper(fileType[:1]) + fileType[1:]
				cli.PrintHeader(fmt.Sprintf("%ss (%d)", typeName, len(resp.Files)))
			} else {
				cli.PrintHeader("Recent Files")
			}
			fmt.Println()

			columns := []cli.Column{
				{Title: "Name", Width: 35},
				{Title: "ID", Width: 35},
				{Title: "Type", Width: 15},
			}

			var rows [][]string
			for _, file := range resp.Files {
				icon := cli.FileTypeIcon(file.MimeType)
				typeName := getTypeName(file.MimeType)
				rows = append(rows, []string{
					icon + " " + cli.Truncate(file.Name, 32),
					file.Id,
					typeName,
				})
			}

			cli.PrintTable(columns, rows)

			if len(resp.Files) > 0 {
				cli.PrintTip(fmt.Sprintf("Use ID in workflow: gdrive.get_file(client, \"%s\")", resp.Files[0].Id))
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&folderName, "folder", "f", "", "Filter by folder name")
	cmd.Flags().StringVar(&folderID, "folder-id", "", "Filter by folder ID")
	cmd.Flags().Int64VarP(&maxResults, "limit", "n", 20, "Maximum number of results")
	cmd.Flags().StringVarP(&fileType, "type", "t", "", "Filter by type: spreadsheet, document, presentation, folder, pdf, image, video")
	cmd.Flags().BoolVar(&showTrashed, "trashed", false, "Show trashed files")

	return cmd
}

func driveFindCmd() *cobra.Command {
	var folderID string

	cmd := &cobra.Command{
		Use:   "find <name>",
		Short: "Find a file by name",
		Long: `Find a file by name in Google Drive.

Examples:
  vulgar gdrive find "report.pdf"
  vulgar gdrive find "budget" --folder-id 1abc123xyz`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			service, err := GetDriveService()
			if err != nil {
				cli.PrintError("%v", err)
				return nil
			}

			cli.PrintLoading(fmt.Sprintf("Searching for \"%s\"", name))

			query := fmt.Sprintf("name contains '%s' and trashed = false", name)
			if folderID != "" {
				query += fmt.Sprintf(" and '%s' in parents", folderID)
			}

			resp, err := service.Files.List().
				Q(query).
				Fields("files(id, name, mimeType, size, modifiedTime, webViewLink, parents)").
				OrderBy("modifiedTime desc").
				PageSize(10).
				Do()

			if err != nil {
				cli.PrintFailed()
				cli.PrintError("Search failed: %v", err)
				return nil
			}

			cli.PrintDone()

			if len(resp.Files) == 0 {
				cli.PrintWarning("No files found matching: %s", name)
				return nil
			}

			cli.PrintHeader(fmt.Sprintf("Found %d file(s) matching \"%s\"", len(resp.Files), name))
			fmt.Println()

			columns := []cli.Column{
				{Title: "Name", Width: 35},
				{Title: "ID", Width: 35},
				{Title: "Type", Width: 15},
			}

			var rows [][]string
			for _, file := range resp.Files {
				icon := cli.FileTypeIcon(file.MimeType)
				typeName := getTypeName(file.MimeType)
				rows = append(rows, []string{
					icon + " " + cli.Truncate(file.Name, 32),
					file.Id,
					typeName,
				})
			}

			cli.PrintTable(columns, rows)

			cli.PrintTip(fmt.Sprintf("Use in workflow: gdrive.get_file(client, \"%s\")", resp.Files[0].Id))

			return nil
		},
	}

	cmd.Flags().StringVar(&folderID, "folder-id", "", "Search within folder ID")

	return cmd
}

func drivePathCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "path <path>",
		Short: "Find a file by path",
		Long: `Navigate to a file using a path-like syntax.

Examples:
  vulgar gdrive path "/Reports/2025/January/report.pdf"
  vulgar gdrive path "/My Folder/subfolder"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			service, err := GetDriveService()
			if err != nil {
				cli.PrintError("%v", err)
				return nil
			}

			parts := splitPath(path)
			if len(parts) == 0 {
				cli.PrintError("Invalid path")
				return nil
			}

			cli.PrintLoading(fmt.Sprintf("Navigating to %s", path))

			parentID := "root"
			var lastFile interface{}

			for i, part := range parts {
				isLastPart := i == len(parts)-1

				query := fmt.Sprintf("name = '%s' and '%s' in parents and trashed = false", part, parentID)
				if !isLastPart {
					query += " and mimeType = 'application/vnd.google-apps.folder'"
				}

				resp, err := service.Files.List().
					Q(query).
					Fields("files(id, name, mimeType, size, webViewLink)").
					PageSize(1).
					Do()

				if err != nil {
					cli.PrintFailed()
					cli.PrintError("Failed to navigate: %v", err)
					return nil
				}

				if len(resp.Files) == 0 {
					cli.PrintFailed()
					cli.PrintError("Not found: %s (at '%s')", path, part)
					return nil
				}

				lastFile = resp.Files[0]
				parentID = resp.Files[0].Id
			}

			cli.PrintDone()

			if file, ok := lastFile.(interface{ GetName() string }); ok {
				_ = file
			}

			if resp, err := service.Files.Get(parentID).Fields("id, name, mimeType, webViewLink").Do(); err == nil {
				cli.PrintHeader("Found")
				fmt.Println()

				icon := cli.FileTypeIcon(resp.MimeType)
				fmt.Printf("  %s %s\n", icon, resp.Name)
				fmt.Println()
				fmt.Println(cli.Muted("  ID: ") + cli.Code(resp.Id))
				fmt.Println(cli.Muted("  Type: ") + getTypeName(resp.MimeType))
				if resp.WebViewLink != "" {
					fmt.Println(cli.Muted("  Link: ") + resp.WebViewLink)
				}

				cli.PrintTip(fmt.Sprintf("gdrive.get_file(client, \"%s\")", resp.Id))
			}

			return nil
		},
	}

	return cmd
}

func driveFoldersCmd() *cobra.Command {
	var parentID string

	cmd := &cobra.Command{
		Use:   "folders",
		Short: "List folders in Google Drive",
		Long: `List only folders in Google Drive.

Examples:
  vulgar gdrive folders
  vulgar gdrive folders --parent-id 1abc123xyz`,
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := GetDriveService()
			if err != nil {
				cli.PrintError("%v", err)
				return nil
			}

			cli.PrintLoading("Fetching folders")

			query := "mimeType = 'application/vnd.google-apps.folder' and trashed = false"
			if parentID != "" {
				query += fmt.Sprintf(" and '%s' in parents", parentID)
			}

			resp, err := service.Files.List().
				Q(query).
				Fields("files(id, name, webViewLink)").
				OrderBy("name").
				PageSize(50).
				Do()

			if err != nil {
				cli.PrintFailed()
				cli.PrintError("Failed to list folders: %v", err)
				return nil
			}

			cli.PrintDone()

			if len(resp.Files) == 0 {
				cli.PrintWarning("No folders found")
				return nil
			}

			cli.PrintHeader(fmt.Sprintf("Folders (%d)", len(resp.Files)))
			fmt.Println()

			columns := []cli.Column{
				{Title: "Name", Width: 40},
				{Title: "ID", Width: 40},
			}

			var rows [][]string
			for _, file := range resp.Files {
				rows = append(rows, []string{
					"[DIR] " + file.Name,
					file.Id,
				})
			}

			cli.PrintTable(columns, rows)

			cli.PrintTip("Use folder ID: gdrive.list_files(client, {folder_id = \"...\"})")

			return nil
		},
	}

	cmd.Flags().StringVar(&parentID, "parent-id", "", "Parent folder ID")

	return cmd
}

func driveRecentCmd() *cobra.Command {
	var days int
	var maxResults int64

	cmd := &cobra.Command{
		Use:   "recent",
		Short: "List recently modified files",
		Long: `List files that were recently modified in Google Drive.

Examples:
  vulgar gdrive recent              # Files modified in last 7 days
  vulgar gdrive recent --days 30    # Files modified in last 30 days`,
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := GetDriveService()
			if err != nil {
				cli.PrintError("%v", err)
				return nil
			}

			cli.PrintLoading(fmt.Sprintf("Fetching files modified in last %d days", days))

			// Calculate date threshold
			threshold := time.Now().AddDate(0, 0, -days).Format(time.RFC3339)
			query := fmt.Sprintf("modifiedTime > '%s' and trashed = false", threshold)

			resp, err := service.Files.List().
				Q(query).
				Fields("files(id, name, mimeType, modifiedTime, webViewLink)").
				OrderBy("modifiedTime desc").
				PageSize(maxResults).
				Do()

			if err != nil {
				cli.PrintFailed()
				cli.PrintError("Failed to list files: %v", err)
				return nil
			}

			cli.PrintDone()

			if len(resp.Files) == 0 {
				cli.PrintWarning("No recently modified files found")
				return nil
			}

			cli.PrintHeader(fmt.Sprintf("Recently Modified (%d files)", len(resp.Files)))
			fmt.Println()

			for i, file := range resp.Files {
				icon := cli.FileTypeIcon(file.MimeType)
				typeName := getTypeName(file.MimeType)
				modified := file.ModifiedTime
				if t, err := time.Parse(time.RFC3339, file.ModifiedTime); err == nil {
					modified = cli.TimeAgo(t)
				}
				fmt.Printf("  %s%s%s %s %s\n", cli.Bold("["), cli.Bold(fmt.Sprintf("%d", i+1)), cli.Bold("]"), icon, file.Name)
				fmt.Printf("      %s %s\n", cli.Muted("ID:"), file.Id)
				fmt.Printf("      %s %s  %s %s\n", cli.Muted("Type:"), typeName, cli.Muted("Modified:"), modified)
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&days, "days", 7, "Number of days to look back")
	cmd.Flags().Int64VarP(&maxResults, "limit", "n", 20, "Maximum number of results")

	return cmd
}

func driveStarredCmd() *cobra.Command {
	var maxResults int64

	cmd := &cobra.Command{
		Use:   "starred",
		Short: "List starred files",
		Long: `List files that are starred in Google Drive.

Examples:
  vulgar gdrive starred`,
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := GetDriveService()
			if err != nil {
				cli.PrintError("%v", err)
				return nil
			}

			cli.PrintLoading("Fetching starred files")

			query := "starred = true and trashed = false"

			resp, err := service.Files.List().
				Q(query).
				Fields("files(id, name, mimeType, modifiedTime, webViewLink)").
				OrderBy("modifiedTime desc").
				PageSize(maxResults).
				Do()

			if err != nil {
				cli.PrintFailed()
				cli.PrintError("Failed to list files: %v", err)
				return nil
			}

			cli.PrintDone()

			if len(resp.Files) == 0 {
				cli.PrintWarning("No starred files found")
				return nil
			}

			cli.PrintHeader(fmt.Sprintf("Starred Files (%d)", len(resp.Files)))
			fmt.Println()

			for i, file := range resp.Files {
				icon := cli.FileTypeIcon(file.MimeType)
				typeName := getTypeName(file.MimeType)
				fmt.Printf("  %s%s%s %s %s\n", cli.Bold("["), cli.Bold(fmt.Sprintf("%d", i+1)), cli.Bold("]"), icon, file.Name)
				fmt.Printf("      %s %s\n", cli.Muted("ID:"), file.Id)
				fmt.Printf("      %s %s\n", cli.Muted("Type:"), typeName)
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().Int64VarP(&maxResults, "limit", "n", 50, "Maximum number of results")

	return cmd
}

func splitPath(path string) []string {
	var parts []string
	current := ""
	for _, c := range path {
		if c == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

func getTypeName(mimeType string) string {
	switch {
	case strings.Contains(mimeType, "folder"):
		return "folder"
	case strings.Contains(mimeType, "spreadsheet"):
		return "spreadsheet"
	case strings.Contains(mimeType, "document"):
		return "document"
	case strings.Contains(mimeType, "presentation"):
		return "presentation"
	case strings.Contains(mimeType, "pdf"):
		return "pdf"
	case strings.Contains(mimeType, "image"):
		return "image"
	case strings.Contains(mimeType, "video"):
		return "video"
	default:
		parts := strings.Split(mimeType, "/")
		if len(parts) > 1 {
			return parts[1]
		}
		return "file"
	}
}
