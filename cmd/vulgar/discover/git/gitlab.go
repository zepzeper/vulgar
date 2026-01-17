package git

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zepzeper/vulgar/internal/cli"
	"github.com/zepzeper/vulgar/internal/config"
	"github.com/zepzeper/vulgar/internal/services/gitlab"
)

func GitLabCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gitlab",
		Short: "Discover GitLab resources",
		Long: `Discover GitLab projects, issues, merge requests, and more.

Requires GitLab token to be configured. Set up with:
  vulgar init
  
Then configure token in ~/.config/vulgar/config.toml:
  [gitlab]
  token = "glpat-your_personal_access_token"
  url = "https://gitlab.com"  # or your self-hosted instance
  projects = ["group/project"]  # optional: default projects`,
	}

	cmd.AddCommand(gitlabProjectsCmd())
	cmd.AddCommand(gitlabIssuesCmd())
	cmd.AddCommand(gitlabMRsCmd())
	cmd.AddCommand(gitlabCheckCmd())
	cmd.AddCommand(gitlabCommitsCmd())
	cmd.AddCommand(gitlabPipelinesCmd())

	return cmd
}

func resolveProject(flagProject string, configuredProjects []string) (string, error) {
	if flagProject != "" {
		return flagProject, nil
	}

	switch len(configuredProjects) {
	case 0:
		return "", fmt.Errorf("no project specified. Use --project flag or configure projects in config.toml")
	case 1:
		return configuredProjects[0], nil
	default:
		return cli.SelectString("Select a project", configuredProjects)
	}
}

func gitlabCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check GitLab token and show user info",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := gitlab.NewClientFromConfig()
			if err != nil {
				cli.PrintError("GitLab token not configured")
				fmt.Println()
				fmt.Println("  Run: vulgar init")
				fmt.Println("  Then edit: " + config.ConfigPath())
				fmt.Println()
				fmt.Println("  Set token to your GitLab personal access token (glpat-...)")
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
				ID:        user.ID,
				Login:     user.Username,
				Name:      user.Name,
				Email:     user.Email,
				URL:       user.WebURL,
				AvatarURL: user.AvatarURL,
			}

			PrintUserInfo(userInfo, "GitLab")

			fmt.Printf("  Instance: %s\n", client.BaseURL())

			if projects := client.Projects(); len(projects) > 0 {
				fmt.Printf("  Configured projects: %d\n", len(projects))
				for _, p := range projects {
					fmt.Printf("    - %s\n", p)
				}
			}

			return nil
		},
	}

	return cmd
}

func gitlabProjectsCmd() *cobra.Command {
	var visibility string
	var limit int
	var owned bool
	var search string
	var plain bool

	cmd := &cobra.Command{
		Use:   "projects",
		Short: "List GitLab projects",
		Long: `List projects for a user or group.

Examples:
  vulgar gitlab projects
  vulgar gitlab projects --owned
  vulgar gitlab projects --visibility private
  vulgar gitlab projects --search myproject
  vulgar gitlab projects --plain    # Non-interactive output`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := gitlab.NewClientFromConfig()
			if err != nil {
				cli.PrintError("%v", err)
				return nil
			}

			cli.PrintLoading("Fetching projects")

			opts := gitlab.ProjectListOptions{
				ListOptions: gitlab.ListOptions{
					PerPage: limit,
				},
				Membership: true,
				Owned:      owned,
				Search:     search,
			}
			if visibility != "all" {
				opts.Visibility = visibility
			}

			projects, err := client.ListProjects(context.Background(), opts)
			if err != nil {
				cli.PrintFailed()
				cli.PrintError("Failed to fetch projects: %v", err)
				return nil
			}

			cli.PrintDone()

			if len(projects) == 0 {
				cli.PrintWarning("No projects found")
				return nil
			}

			cli.PrintSuccess("Found %d project(s)", len(projects))
			fmt.Println()

			repos := make([]RepoInfo, 0, len(projects))
			for _, p := range projects {
				repos = append(repos, RepoInfo{
					ID:          p.ID,
					Name:        p.Name,
					FullName:    p.PathWithNamespace,
					Description: p.Description,
					Private:     p.Visibility == "private",
					URL:         p.WebURL,
					CloneURL:    p.SSHURLToRepo,
					Stars:       p.StarCount,
					Forks:       p.ForksCount,
					UpdatedAt:   p.LastActivityAt,
				})
			}

			// Plain mode
			if plain {
				for _, repo := range repos {
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
			selected, err := cli.Select("Select a project", repos, func(r RepoInfo) string {
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
				for _, repo := range repos {
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
			cli.PrintHeader("Selected Project")
			fmt.Println()
			fmt.Printf("  %s %s\n", cli.Muted("Name:"), selected.Name)
			fmt.Printf("  %s %s\n", cli.Muted("Full Name:"), selected.FullName)
			if selected.Description != "" {
				fmt.Printf("  %s %s\n", cli.Muted("Description:"), selected.Description)
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

	cmd.Flags().StringVar(&visibility, "visibility", "all", "Filter by visibility (all, public, private, internal)")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of results")
	cmd.Flags().BoolVar(&owned, "owned", false, "Only show owned projects")
	cmd.Flags().StringVar(&search, "search", "", "Search projects by name")
	cmd.Flags().BoolVar(&plain, "plain", false, "Plain output (non-interactive)")

	return cmd
}

func gitlabIssuesCmd() *cobra.Command {
	var project string
	var state string
	var limit int
	var plain bool

	cmd := &cobra.Command{
		Use:   "issues",
		Short: "List GitLab issues",
		Long: `List issues for a project.

If no --project flag is provided:
  - Uses the configured project if exactly one is set
  - Prompts for selection if multiple projects are configured

Examples:
  vulgar gitlab issues
  vulgar gitlab issues --project group/project
  vulgar gitlab issues --state closed
  vulgar gitlab issues --plain    # Non-interactive output`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := gitlab.NewClientFromConfig()
			if err != nil {
				cli.PrintError("%v", err)
				return nil
			}

			resolvedProject, err := resolveProject(project, client.Projects())
			if err != nil {
				cli.PrintError("%v", err)
				return nil
			}

			cli.PrintLoading("Fetching issues")

			opts := gitlab.IssueListOptions{
				ListOptions: gitlab.ListOptions{
					PerPage: limit,
					State:   state,
				},
			}

			issues, err := client.ListIssues(context.Background(), resolvedProject, opts)
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

			issueInfos := make([]IssueInfo, 0, len(issues))
			for _, i := range issues {
				author := ""
				if i.Author != nil {
					author = i.Author.Username
				}

				issueInfos = append(issueInfos, IssueInfo{
					Number:    i.IID,
					Title:     i.Title,
					State:     i.State,
					Author:    author,
					CreatedAt: i.CreatedAt,
					UpdatedAt: i.UpdatedAt,
					URL:       i.WebURL,
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

	cmd.Flags().StringVar(&project, "project", "", "Project path (group/project)")
	cmd.Flags().StringVar(&state, "state", "opened", "Issue state (opened, closed, all)")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of results")
	cmd.Flags().BoolVar(&plain, "plain", false, "Plain output (non-interactive)")

	return cmd
}

func gitlabMRsCmd() *cobra.Command {
	var project string
	var state string
	var limit int
	var plain bool

	cmd := &cobra.Command{
		Use:   "mrs",
		Short: "List GitLab merge requests",
		Long: `List merge requests for a project.

If no --project flag is provided:
  - Uses the configured project if exactly one is set
  - Prompts for selection if multiple projects are configured

Examples:
  vulgar gitlab mrs
  vulgar gitlab mrs --project group/project
  vulgar gitlab mrs --state merged
  vulgar gitlab mrs --plain    # Non-interactive output`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := gitlab.NewClientFromConfig()
			if err != nil {
				cli.PrintError("%v", err)
				return nil
			}

			resolvedProject, err := resolveProject(project, client.Projects())
			if err != nil {
				cli.PrintError("%v", err)
				return nil
			}

			cli.PrintLoading("Fetching merge requests")

			opts := gitlab.MergeRequestListOptions{
				ListOptions: gitlab.ListOptions{
					PerPage: limit,
					State:   state,
				},
			}

			mrs, err := client.ListMergeRequests(context.Background(), resolvedProject, opts)
			if err != nil {
				cli.PrintFailed()
				cli.PrintError("Failed to fetch merge requests: %v", err)
				return nil
			}

			cli.PrintDone()

			if len(mrs) == 0 {
				cli.PrintWarning("No merge requests found")
				return nil
			}

			cli.PrintSuccess("Found %d merge request(s)", len(mrs))
			fmt.Println()

			prInfos := make([]PRInfo, 0, len(mrs))
			for _, mr := range mrs {
				author := ""
				if mr.Author != nil {
					author = mr.Author.Username
				}

				prInfos = append(prInfos, PRInfo{
					Number:       mr.IID,
					Title:        mr.Title,
					State:        mr.State,
					Author:       author,
					SourceBranch: mr.SourceBranch,
					TargetBranch: mr.TargetBranch,
					IsMerged:     mr.State == "merged",
					CreatedAt:    mr.CreatedAt,
					UpdatedAt:    mr.UpdatedAt,
					URL:          mr.WebURL,
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
			selected, err := cli.Select("Select a merge request", prInfos, func(p PRInfo) string {
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
			cli.PrintHeader("Selected Merge Request")
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
			cli.PrintTip(fmt.Sprintf("View MR: %s", selected.URL))

			return nil
		},
	}

	cmd.Flags().StringVar(&project, "project", "", "Project path (group/project)")
	cmd.Flags().StringVar(&state, "state", "opened", "MR state (opened, closed, merged, all)")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of results")
	cmd.Flags().BoolVar(&plain, "plain", false, "Plain output (non-interactive)")

	return cmd
}

func gitlabCommitsCmd() *cobra.Command {
	var project string
	var ref string
	var limit int
	var since int

	cmd := &cobra.Command{
		Use:   "commits",
		Short: "List GitLab commits",
		Long: `List commits for a project.

If no --project flag is provided:
  - Uses the configured project if exactly one is set
  - Prompts for selection if multiple projects are configured

Examples:
  vulgar gitlab commits
  vulgar gitlab commits --project group/project
  vulgar gitlab commits --ref main
  vulgar gitlab commits --since 24`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := gitlab.NewClientFromConfig()
			if err != nil {
				cli.PrintError("%v", err)
				return nil
			}

			resolvedProject, err := resolveProject(project, client.Projects())
			if err != nil {
				cli.PrintError("%v", err)
				return nil
			}

			cli.PrintLoading("Fetching commits")

			opts := gitlab.CommitListOptions{
				ListOptions: gitlab.ListOptions{
					PerPage: limit,
				},
				RefName: ref,
			}
			if since > 0 {
				opts.Since = gitlab.SinceHours(since)
			}

			commits, err := client.ListCommits(context.Background(), resolvedProject, opts)
			if err != nil {
				cli.PrintFailed()
				cli.PrintError("Failed to fetch commits: %v", err)
				return nil
			}

			cli.PrintDone()

			cli.PrintHeader(fmt.Sprintf("Commits in %s", resolvedProject))
			printCommitsTable(commits)

			return nil
		},
	}

	cmd.Flags().StringVar(&project, "project", "", "Project path (group/project)")
	cmd.Flags().StringVar(&ref, "ref", "", "Branch or tag name")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of results")
	cmd.Flags().IntVar(&since, "since", 0, "Only show commits from last N hours")

	return cmd
}

func gitlabPipelinesCmd() *cobra.Command {
	var project string
	var status string
	var ref string
	var limit int

	cmd := &cobra.Command{
		Use:   "pipelines",
		Short: "List GitLab pipelines",
		Long: `List CI/CD pipelines for a project.

If no --project flag is provided:
  - Uses the configured project if exactly one is set
  - Prompts for selection if multiple projects are configured

Examples:
  vulgar gitlab pipelines
  vulgar gitlab pipelines --project group/project
  vulgar gitlab pipelines --status failed
  vulgar gitlab pipelines --ref main`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := gitlab.NewClientFromConfig()
			if err != nil {
				cli.PrintError("%v", err)
				return nil
			}

			resolvedProject, err := resolveProject(project, client.Projects())
			if err != nil {
				cli.PrintError("%v", err)
				return nil
			}

			cli.PrintLoading("Fetching pipelines")

			opts := gitlab.PipelineListOptions{
				ListOptions: gitlab.ListOptions{
					PerPage: limit,
				},
				Status: status,
				Ref:    ref,
			}

			pipelines, err := client.ListPipelines(context.Background(), resolvedProject, opts)
			if err != nil {
				cli.PrintFailed()
				cli.PrintError("Failed to fetch pipelines: %v", err)
				return nil
			}

			cli.PrintDone()

			cli.PrintHeader(fmt.Sprintf("Pipelines in %s", resolvedProject))
			printPipelinesTable(pipelines)

			if len(pipelines) > 0 {
				cli.PrintTip(fmt.Sprintf("View pipeline: %s", pipelines[0].WebURL))
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&project, "project", "", "Project path (group/project)")
	cmd.Flags().StringVar(&status, "status", "", "Filter by status (running, pending, success, failed, canceled)")
	cmd.Flags().StringVar(&ref, "ref", "", "Filter by branch/tag")
	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of results")

	return cmd
}

func printCommitsTable(commits []gitlab.Commit) {
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
		rows[i] = []string{
			c.ShortID,
			cli.Truncate(c.Title, 50),
			cli.Truncate(c.AuthorName, 20),
			formatTimeAgo(c.CreatedAt),
		}
	}

	cli.PrintTable(columns, rows)
}

func printPipelinesTable(pipelines []gitlab.Pipeline) {
	if len(pipelines) == 0 {
		cli.PrintWarning("No pipelines found")
		return
	}

	columns := []cli.Column{
		{Title: "ID", Width: 10},
		{Title: "Status", Width: 12},
		{Title: "Ref", Width: 20},
		{Title: "Updated", Width: 12},
	}

	rows := make([][]string, len(pipelines))
	for i, p := range pipelines {
		rows[i] = []string{
			fmt.Sprintf("%d", p.ID),
			cli.FormatStatus(p.Status),
			cli.Truncate(p.Ref, 20),
			formatTimeAgo(p.UpdatedAt),
		}
	}

	cli.PrintTable(columns, rows)
}
