package git

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zepzeper/vulgar/internal/cli"
	"github.com/zepzeper/vulgar/internal/config"
	"github.com/zepzeper/vulgar/internal/services/github"
)

func GitHubCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "github",
		Short: "Discover GitHub resources",
		Long: `Discover GitHub repositories, issues, pull requests, and more.

Requires GitHub token to be configured. Set up with:
  vulgar init
  
Then configure token in ~/.config/vulgar/config.toml:
  [github]
  token = "ghp_your_personal_access_token"
  default_owner = "your-username"`,
	}

	cmd.AddCommand(githubReposCmd())
	cmd.AddCommand(githubIssuesCmd())
	cmd.AddCommand(githubPRsCmd())
	cmd.AddCommand(githubCheckCmd())
	cmd.AddCommand(githubCommitsCmd())

	return cmd
}

func githubCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check GitHub token and show user info",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := github.NewClientFromConfig()
			if err != nil {
				cli.PrintError("GitHub token not configured")
				fmt.Println()
				fmt.Println("  Run: vulgar init")
				fmt.Println("  Then edit: " + config.ConfigPath())
				fmt.Println()
				fmt.Println("  Set token to your GitHub personal access token (ghp_...)")
				return nil
			}

			cli.PrintLoading("Checking token")

			user, err := client.GetCurrentUser(context.Background())
			if err != nil {
				cli.PrintFailed()
				cli.PrintError("Token verification failed: %v", err)
				return nil
			}

			cli.PrintDone()
			cli.PrintSuccess("Token is valid!")

			userInfo := UserInfo{
				Login:     user.Login,
				Name:      user.Name,
				Email:     user.Email,
				Bio:       user.Bio,
				URL:       user.HTMLURL,
				AvatarURL: user.AvatarURL,
			}

			PrintUserInfo(userInfo, "GitHub")

			fmt.Println()
			rateLimit, err := client.GetRateLimit(context.Background())
			if err == nil && rateLimit != nil {
				fmt.Printf("  Rate Limit: %d / %d remaining\n", rateLimit.Remaining, rateLimit.Limit)
			}

			if owner := client.DefaultOwner(); owner != "" {
				fmt.Printf("  Default Owner: %s\n", owner)
			}

			return nil
		},
	}

	return cmd
}

func githubReposCmd() *cobra.Command {
	var owner string
	var visibility string
	var limit int
	var plain bool

	cmd := &cobra.Command{
		Use:   "repos",
		Short: "List GitHub repositories",
		Long: `List repositories for a user or organization.

Examples:
  vulgar github repos
  vulgar github repos --owner myorg
  vulgar github repos --visibility private
  vulgar github repos --plain    # Non-interactive output`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := github.NewClientFromConfig()
			if err != nil {
				cli.PrintError("%v", err)
				return nil
			}

			cli.PrintLoading("Fetching repositories")

			opts := github.RepoListOptions{
				ListOptions: github.ListOptions{
					PerPage: limit,
					Sort:    "updated",
				},
				Visibility: visibility,
			}

			var repos []github.Repository
			if owner != "" {
				repos, err = client.ListOwnerRepositories(context.Background(), owner, opts)
			} else {
				repos, err = client.ListUserRepositories(context.Background(), opts)
			}

			if err != nil {
				cli.PrintFailed()
				cli.PrintError("Failed to fetch repositories: %v", err)
				return nil
			}

			cli.PrintDone()

			if len(repos) == 0 {
				cli.PrintWarning("No repositories found")
				return nil
			}

			cli.PrintSuccess("Found %d repository(ies)", len(repos))
			fmt.Println()

			repoInfos := make([]RepoInfo, 0, len(repos))
			for _, r := range repos {
				repoInfos = append(repoInfos, RepoInfo{
					ID:          r.ID,
					Name:        r.Name,
					FullName:    r.FullName,
					Description: r.Description,
					Private:     r.Private,
					URL:         r.HTMLURL,
					CloneURL:    r.CloneURL,
					Language:    r.Language,
					Stars:       r.Stars,
					Forks:       r.Forks,
					UpdatedAt:   r.UpdatedAt,
				})
			}

			// Plain mode
			if plain {
				for _, repo := range repoInfos {
					visibility := ""
					if repo.Private {
						visibility = " [P]"
					}
					fmt.Printf("%s%s\n", repo.Name, visibility)
					fmt.Printf("  %s\n", repo.FullName)
					if repo.Description != "" {
						fmt.Printf("  %s\n", repo.Description)
					}
					fmt.Printf("  %s\n", repo.CloneURL)
					fmt.Println()
				}
				return nil
			}

			// Interactive mode
			selected, err := cli.Select("Select a repository", repoInfos, func(r RepoInfo) string {
				visibility := ""
				if r.Private {
					visibility = " [P]"
				}
				desc := r.Description
				if desc == "" {
					desc = "No description"
				}
				return fmt.Sprintf("%s%s  %s", r.Name, visibility, cli.Muted("("+cli.Truncate(desc, 40)+")"))
			})

			if err != nil {
				// User cancelled - show list
				for _, repo := range repoInfos {
					visibility := ""
					if repo.Private {
						visibility = " [P]"
					}
					fmt.Printf("%s%s\n", repo.Name, visibility)
					fmt.Printf("  %s\n", repo.FullName)
					fmt.Println()
				}
				return nil
			}

			// Show selected item details
			fmt.Println()
			cli.PrintHeader("Selected Repository")
			fmt.Println()
			fmt.Printf("  %s %s\n", cli.Muted("Name:"), selected.Name)
			fmt.Printf("  %s %s\n", cli.Muted("Full Name:"), selected.FullName)
			if selected.Description != "" {
				fmt.Printf("  %s %s\n", cli.Muted("Description:"), selected.Description)
			}
			if selected.Language != "" {
				fmt.Printf("  %s %s\n", cli.Muted("Language:"), selected.Language)
			}
			fmt.Printf("  %s %d stars, %d forks\n", cli.Muted("Stats:"), selected.Stars, selected.Forks)
			fmt.Printf("  %s %s\n", cli.Muted("Updated:"), formatTimeAgo(selected.UpdatedAt))
			fmt.Println()
			fmt.Printf("  %s\n", cli.Bold("Clone URL (copy this):"))
			fmt.Printf("  %s\n", selected.CloneURL)
			fmt.Printf("  %s\n", selected.URL)
			fmt.Println()
			cli.PrintTip(fmt.Sprintf("Clone with: git clone %s", selected.CloneURL))

			return nil
		},
	}

	cmd.Flags().StringVar(&owner, "owner", "", "Repository owner (user or org)")
	cmd.Flags().StringVar(&visibility, "visibility", "all", "Filter by visibility (all, public, private)")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of results")
	cmd.Flags().BoolVar(&plain, "plain", false, "Plain output (non-interactive)")

	return cmd
}

func githubIssuesCmd() *cobra.Command {
	var repo string
	var state string
	var limit int
	var plain bool

	cmd := &cobra.Command{
		Use:   "issues",
		Short: "List GitHub issues",
		Long: `List issues for a repository.

Examples:
  vulgar github issues --repo owner/repo
  vulgar github issues --repo owner/repo --state closed
  vulgar github issues --repo owner/repo --plain    # Non-interactive output`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if repo == "" {
				cli.PrintError("--repo flag is required")
				return nil
			}

			parts := strings.SplitN(repo, "/", 2)
			if len(parts) != 2 {
				cli.PrintError("Invalid repo format. Use owner/repo")
				return nil
			}
			owner, repoName := parts[0], parts[1]

			client, err := github.NewClientFromConfig()
			if err != nil {
				cli.PrintError("%v", err)
				return nil
			}

			cli.PrintLoading("Fetching issues")

			opts := github.IssueListOptions{
				ListOptions: github.ListOptions{
					PerPage: limit,
				},
				State: state,
			}

			issues, err := client.ListIssues(context.Background(), owner, repoName, opts)
			if err != nil {
				cli.PrintFailed()
				cli.PrintError("Failed to fetch issues: %v", err)
				return nil
			}

			cli.PrintDone()

			if len(issues) == 0 {
				cli.PrintWarning("No issues found")
				return nil
			}

			cli.PrintSuccess("Found %d issue(s)", len(issues))
			fmt.Println()

			issueInfos := make([]IssueInfo, 0)
			for _, i := range issues {
				if strings.Contains(i.HTMLURL, "/pull/") {
					continue
				}

				author := ""
				if i.User != nil {
					author = i.User.Login
				}

				issueInfos = append(issueInfos, IssueInfo{
					Number:    i.Number,
					Title:     i.Title,
					State:     i.State,
					Author:    author,
					CreatedAt: i.CreatedAt,
					UpdatedAt: i.UpdatedAt,
					URL:       i.HTMLURL,
				})
			}

			// Plain mode
			if plain {
				for _, issue := range issueInfos {
					fmt.Printf("#%d %s\n", issue.Number, issue.Title)
					fmt.Printf("  %s\n", issue.URL)
					fmt.Println()
				}
				return nil
			}

			// Create comparable wrapper type for selection
			type issueSelect struct {
				Number    int
				Title     string
				State     string
				Author    string
				CreatedAt string
				UpdatedAt string
				URL       string
			}
			selectOptions := make([]issueSelect, len(issueInfos))
			for i, issue := range issueInfos {
				selectOptions[i] = issueSelect{
					Number:    issue.Number,
					Title:     issue.Title,
					State:     issue.State,
					Author:    issue.Author,
					CreatedAt: issue.CreatedAt,
					UpdatedAt: issue.UpdatedAt,
					URL:       issue.URL,
				}
			}

			// Interactive mode
			selected, err := cli.Select("Select an issue", selectOptions, func(i issueSelect) string {
				return fmt.Sprintf("#%d %s  %s", i.Number, i.Title, cli.Muted("("+i.State+", "+formatTimeAgo(i.UpdatedAt)+")"))
			})

			if err != nil {
				// User cancelled - show list
				for _, issue := range issueInfos {
					fmt.Printf("#%d %s\n", issue.Number, issue.Title)
					fmt.Println()
				}
				return nil
			}

			// Show selected item details
			fmt.Println()
			cli.PrintHeader("Selected Issue")
			fmt.Println()
			fmt.Printf("  %s #%d\n", cli.Muted("Number:"), selected.Number)
			fmt.Printf("  %s %s\n", cli.Muted("Title:"), selected.Title)
			fmt.Printf("  %s %s\n", cli.Muted("State:"), selected.State)
			if selected.Author != "" {
				fmt.Printf("  %s %s\n", cli.Muted("Author:"), selected.Author)
			}
			fmt.Printf("  %s %s\n", cli.Muted("Created:"), formatTimeAgo(selected.CreatedAt))
			fmt.Printf("  %s %s\n", cli.Muted("Updated:"), formatTimeAgo(selected.UpdatedAt))
			fmt.Println()
			fmt.Printf("  %s\n", cli.Bold("URL (copy this):"))
			fmt.Printf("  %s\n", selected.URL)
			fmt.Println()
			cli.PrintTip(fmt.Sprintf("View issue: %s", selected.URL))

			return nil
		},
	}

	cmd.Flags().StringVar(&repo, "repo", "", "Repository (owner/repo)")
	cmd.Flags().StringVar(&state, "state", "open", "Issue state (open, closed, all)")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of results")
	cmd.Flags().BoolVar(&plain, "plain", false, "Plain output (non-interactive)")

	return cmd
}

func githubPRsCmd() *cobra.Command {
	var repo string
	var state string
	var limit int
	var plain bool

	cmd := &cobra.Command{
		Use:   "prs",
		Short: "List GitHub pull requests",
		Long: `List pull requests for a repository.

Examples:
  vulgar github prs --repo owner/repo
  vulgar github prs --repo owner/repo --state closed
  vulgar github prs --repo owner/repo --plain    # Non-interactive output`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if repo == "" {
				cli.PrintError("--repo flag is required")
				return nil
			}

			parts := strings.SplitN(repo, "/", 2)
			if len(parts) != 2 {
				cli.PrintError("Invalid repo format. Use owner/repo")
				return nil
			}
			owner, repoName := parts[0], parts[1]

			client, err := github.NewClientFromConfig()
			if err != nil {
				cli.PrintError("%v", err)
				return nil
			}

			cli.PrintLoading("Fetching pull requests")

			opts := github.PullRequestListOptions{
				ListOptions: github.ListOptions{
					PerPage: limit,
				},
				State: state,
			}

			prs, err := client.ListPullRequests(context.Background(), owner, repoName, opts)
			if err != nil {
				cli.PrintFailed()
				cli.PrintError("Failed to fetch pull requests: %v", err)
				return nil
			}

			cli.PrintDone()

			if len(prs) == 0 {
				cli.PrintWarning("No pull requests found")
				return nil
			}

			cli.PrintSuccess("Found %d pull request(s)", len(prs))
			fmt.Println()

			prInfos := make([]PRInfo, 0, len(prs))
			for _, pr := range prs {
				author := ""
				if pr.User != nil {
					author = pr.User.Login
				}

				sourceBranch := ""
				if pr.Head != nil {
					sourceBranch = pr.Head.Ref
				}

				targetBranch := ""
				if pr.Base != nil {
					targetBranch = pr.Base.Ref
				}

				prInfos = append(prInfos, PRInfo{
					Number:       pr.Number,
					Title:        pr.Title,
					State:        pr.State,
					Author:       author,
					SourceBranch: sourceBranch,
					TargetBranch: targetBranch,
					IsMerged:     pr.Merged,
					CreatedAt:    pr.CreatedAt,
					UpdatedAt:    pr.UpdatedAt,
					URL:          pr.HTMLURL,
				})
			}

			// Plain mode
			if plain {
				for _, pr := range prInfos {
					state := pr.State
					if pr.IsMerged {
						state = "merged"
					}
					fmt.Printf("#%d %s [%s]\n", pr.Number, pr.Title, state)
					fmt.Printf("  %s\n", pr.URL)
					fmt.Println()
				}
				return nil
			}

			// Interactive mode
			selected, err := cli.Select("Select a pull request", prInfos, func(p PRInfo) string {
				state := p.State
				if p.IsMerged {
					state = "merged"
				}
				return fmt.Sprintf("#%d %s  %s", p.Number, p.Title, cli.Muted("("+state+", "+formatTimeAgo(p.UpdatedAt)+")"))
			})

			if err != nil {
				// User cancelled - show list
				for _, pr := range prInfos {
					state := pr.State
					if pr.IsMerged {
						state = "merged"
					}
					fmt.Printf("#%d %s [%s]\n", pr.Number, pr.Title, state)
					fmt.Println()
				}
				return nil
			}

			// Show selected item details
			fmt.Println()
			cli.PrintHeader("Selected Pull Request")
			fmt.Println()
			fmt.Printf("  %s #%d\n", cli.Muted("Number:"), selected.Number)
			fmt.Printf("  %s %s\n", cli.Muted("Title:"), selected.Title)
			state := selected.State
			if selected.IsMerged {
				state = "merged"
			}
			fmt.Printf("  %s %s\n", cli.Muted("State:"), state)
			if selected.Author != "" {
				fmt.Printf("  %s %s\n", cli.Muted("Author:"), selected.Author)
			}
			if selected.SourceBranch != "" && selected.TargetBranch != "" {
				fmt.Printf("  %s %s -> %s\n", cli.Muted("Branches:"), selected.SourceBranch, selected.TargetBranch)
			}
			fmt.Printf("  %s %s\n", cli.Muted("Created:"), formatTimeAgo(selected.CreatedAt))
			fmt.Printf("  %s %s\n", cli.Muted("Updated:"), formatTimeAgo(selected.UpdatedAt))
			fmt.Println()
			fmt.Printf("  %s\n", cli.Bold("URL (copy this):"))
			fmt.Printf("  %s\n", selected.URL)
			fmt.Println()
			cli.PrintTip(fmt.Sprintf("View PR: %s", selected.URL))

			return nil
		},
	}

	cmd.Flags().StringVar(&repo, "repo", "", "Repository (owner/repo)")
	cmd.Flags().StringVar(&state, "state", "open", "PR state (open, closed, all)")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of results")
	cmd.Flags().BoolVar(&plain, "plain", false, "Plain output (non-interactive)")

	return cmd
}

func githubCommitsCmd() *cobra.Command {
	var repo string
	var sha string
	var limit int

	cmd := &cobra.Command{
		Use:   "commits",
		Short: "List GitHub commits",
		Long: `List commits for a repository.

Examples:
  vulgar github commits --repo owner/repo
  vulgar github commits --repo owner/repo --sha main`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if repo == "" {
				cli.PrintError("--repo flag is required")
				return nil
			}

			parts := strings.SplitN(repo, "/", 2)
			if len(parts) != 2 {
				cli.PrintError("Invalid repo format. Use owner/repo")
				return nil
			}
			owner, repoName := parts[0], parts[1]

			client, err := github.NewClientFromConfig()
			if err != nil {
				cli.PrintError("%v", err)
				return nil
			}

			cli.PrintLoading("Fetching commits")

			opts := github.CommitListOptions{
				ListOptions: github.ListOptions{
					PerPage: limit,
				},
				SHA: sha,
			}

			commits, err := client.ListCommits(context.Background(), owner, repoName, opts)
			if err != nil {
				cli.PrintFailed()
				cli.PrintError("Failed to fetch commits: %v", err)
				return nil
			}

			cli.PrintDone()

			cli.PrintHeader(fmt.Sprintf("Commits in %s", repo))
			printGitHubCommitsTable(commits)

			return nil
		},
	}

	cmd.Flags().StringVar(&repo, "repo", "", "Repository (owner/repo)")
	cmd.Flags().StringVar(&sha, "sha", "", "SHA or branch to start listing from")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of results")

	return cmd
}

func printGitHubCommitsTable(commits []github.Commit) {
	if len(commits) == 0 {
		cli.PrintWarning("No commits found")
		return
	}

	columns := []cli.Column{
		{Title: "SHA", Width: 8},
		{Title: "Message", Width: 50},
		{Title: "Author", Width: 20},
		{Title: "Date", Width: 12},
	}

	rows := make([][]string, len(commits))
	for i, c := range commits {
		sha := c.SHA
		if len(sha) > 7 {
			sha = sha[:7]
		}

		message := ""
		author := ""
		date := ""

		if c.Commit != nil {
			message = c.Commit.Message
			if idx := strings.Index(message, "\n"); idx > 0 {
				message = message[:idx]
			}

			if c.Commit.Author != nil {
				author = c.Commit.Author.Name
				date = c.Commit.Author.Date
			}
		}

		rows[i] = []string{
			sha,
			cli.Truncate(message, 50),
			cli.Truncate(author, 20),
			formatTimeAgo(date),
		}
	}

	cli.PrintTable(columns, rows)
}
