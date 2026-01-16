package google

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/zepzeper/vulgar/internal/cli"
)

func CalendarCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gcalendar",
		Short: "Discover Google Calendar resources",
		Long: `Discover Google Calendar events and calendars.

Setup:
  1. Create OAuth 2.0 Client ID (Desktop app) in Google Cloud Console
  2. Add client_id and client_secret to ~/.config/vulgar/config.toml
  3. Run 'vulgar gcalendar login' to authenticate`,
	}

	cmd.AddCommand(calendarLoginCmd())
	cmd.AddCommand(calendarLogoutCmd())
	cmd.AddCommand(calendarCheckCmd())
	cmd.AddCommand(calendarEventsCmd())
	cmd.AddCommand(calendarTodayCmd())
	cmd.AddCommand(calendarUpcomingCmd())
	cmd.AddCommand(calendarRecentCmd())
	cmd.AddCommand(calendarListCmd())

	return cmd
}

func calendarLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Authenticate with Google",
		Long:  `Opens a browser to authenticate with your Google account.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGoogleLogin()
		},
	}
}

func calendarLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Log out of Google",
		Long:  `Removes stored Google authentication tokens.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGoogleLogout()
		},
	}
}

func calendarCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Check Google authentication status",
		Long:  `Shows current authentication status and logged-in user.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGoogleCheck()
		},
	}
}

func calendarEventsCmd() *cobra.Command {
	var calendarID string
	var maxResults int64

	cmd := &cobra.Command{
		Use:   "events",
		Short: "List calendar events",
		Long: `List upcoming events from a Google Calendar.

Examples:
  vulgar gcalendar events
  vulgar gcalendar events --calendar-id primary --max 20`,
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := GetCalendarService()
			if err != nil {
				cli.PrintError("%v", err)
				return nil
			}

			cli.PrintLoading("Fetching events")

			resp, err := service.Events.List(calendarID).
				TimeMin(time.Now().Format(time.RFC3339)).
				MaxResults(maxResults).
				SingleEvents(true).
				OrderBy("startTime").
				Do()

			if err != nil {
				cli.PrintFailed()
				cli.PrintError("Failed to get events: %v", err)
				return nil
			}

			cli.PrintDone()

			if len(resp.Items) == 0 {
				cli.PrintWarning("No upcoming events found")
				return nil
			}

			cli.PrintHeader(fmt.Sprintf("Upcoming Events (%d)", len(resp.Items)))
			fmt.Println()

			columns := []cli.Column{
				{Title: "Event", Width: 35},
				{Title: "When", Width: 25},
				{Title: "ID", Width: 25},
			}

			var rows [][]string
			for _, event := range resp.Items {
				when := ""
				if event.Start.DateTime != "" {
					t, _ := time.Parse(time.RFC3339, event.Start.DateTime)
					when = t.Format("Mon Jan 2, 3:04 PM")
				} else {
					when = event.Start.Date + " (all day)"
				}

				rows = append(rows, []string{
					"[E] " + cli.Truncate(event.Summary, 32),
					when,
					cli.Truncate(event.Id, 23),
				})
			}

			cli.PrintTable(columns, rows)

			return nil
		},
	}

	cmd.Flags().StringVar(&calendarID, "calendar-id", "primary", "Calendar ID")
	cmd.Flags().Int64Var(&maxResults, "max", 10, "Maximum number of events")

	return cmd
}

func calendarTodayCmd() *cobra.Command {
	var calendarID string

	cmd := &cobra.Command{
		Use:   "today",
		Short: "Show today's events",
		Long: `Show all events for today.

Examples:
  vulgar gcalendar today
  vulgar gcalendar today --calendar-id work@group.calendar.google.com`,
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := GetCalendarService()
			if err != nil {
				cli.PrintError("%v", err)
				return nil
			}

			cli.PrintLoading("Fetching today's events")

			now := time.Now()
			startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			endOfDay := startOfDay.Add(24 * time.Hour)

			resp, err := service.Events.List(calendarID).
				TimeMin(startOfDay.Format(time.RFC3339)).
				TimeMax(endOfDay.Format(time.RFC3339)).
				SingleEvents(true).
				OrderBy("startTime").
				Do()

			if err != nil {
				cli.PrintFailed()
				cli.PrintError("Failed to get events: %v", err)
				return nil
			}

			cli.PrintDone()

			dateStr := now.Format("Monday, January 2, 2006")

			if len(resp.Items) == 0 {
				cli.PrintHeader(dateStr)
				fmt.Println()
				fmt.Println("  " + cli.Muted("No events scheduled for today"))
				return nil
			}

			cli.PrintHeader(fmt.Sprintf("%s (%d events)", dateStr, len(resp.Items)))
			fmt.Println()

			for _, event := range resp.Items {
				timeStr := "All day"
				if event.Start.DateTime != "" {
					startTime, _ := time.Parse(time.RFC3339, event.Start.DateTime)
					endTime, _ := time.Parse(time.RFC3339, event.End.DateTime)
					timeStr = fmt.Sprintf("%s - %s", startTime.Format("3:04 PM"), endTime.Format("3:04 PM"))
				}

				fmt.Printf("  [T] %s\n", cli.Muted(timeStr))
				fmt.Printf("      [E] %s\n", event.Summary)
				if event.Location != "" {
					fmt.Printf("      [L] %s\n", cli.Muted(event.Location))
				}
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&calendarID, "calendar-id", "primary", "Calendar ID")

	return cmd
}

func calendarUpcomingCmd() *cobra.Command {
	var calendarID string
	var days int

	cmd := &cobra.Command{
		Use:   "upcoming",
		Short: "Show upcoming events",
		Long: `Show events for the next N days.

Examples:
  vulgar gcalendar upcoming
  vulgar gcalendar upcoming --days 14`,
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := GetCalendarService()
			if err != nil {
				cli.PrintError("%v", err)
				return nil
			}

			cli.PrintLoading(fmt.Sprintf("Fetching events for next %d days", days))

			now := time.Now()
			endTime := now.Add(time.Duration(days) * 24 * time.Hour)

			resp, err := service.Events.List(calendarID).
				TimeMin(now.Format(time.RFC3339)).
				TimeMax(endTime.Format(time.RFC3339)).
				SingleEvents(true).
				OrderBy("startTime").
				MaxResults(50).
				Do()

			if err != nil {
				cli.PrintFailed()
				cli.PrintError("Failed to get events: %v", err)
				return nil
			}

			cli.PrintDone()

			if len(resp.Items) == 0 {
				cli.PrintWarning("No events in the next %d days", days)
				return nil
			}

			cli.PrintHeader(fmt.Sprintf("Next %d Days (%d events)", days, len(resp.Items)))
			fmt.Println()

			columns := []cli.Column{
				{Title: "Date", Width: 18},
				{Title: "Time", Width: 15},
				{Title: "Event", Width: 40},
			}

			var rows [][]string
			for _, event := range resp.Items {
				dateStr := ""
				timeStr := "All day"

				if event.Start.DateTime != "" {
					startTime, _ := time.Parse(time.RFC3339, event.Start.DateTime)
					dateStr = startTime.Format("Mon Jan 2")
					timeStr = startTime.Format("3:04 PM")
				} else {
					t, _ := time.Parse("2006-01-02", event.Start.Date)
					dateStr = t.Format("Mon Jan 2")
				}

				rows = append(rows, []string{
					"[D] " + dateStr,
					timeStr,
					cli.Truncate(event.Summary, 38),
				})
			}

			cli.PrintTable(columns, rows)

			return nil
		},
	}

	cmd.Flags().StringVar(&calendarID, "calendar-id", "primary", "Calendar ID")
	cmd.Flags().IntVar(&days, "days", 7, "Number of days to look ahead")

	return cmd
}

func calendarRecentCmd() *cobra.Command {
	var calendarID string
	var days int

	cmd := &cobra.Command{
		Use:   "recent",
		Short: "Show recent past events",
		Long: `Show events from the past N days.

Examples:
  vulgar gcalendar recent              # Events from last 7 days
  vulgar gcalendar recent --days 30    # Events from last 30 days`,
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := GetCalendarService()
			if err != nil {
				cli.PrintError("%v", err)
				return nil
			}

			cli.PrintLoading(fmt.Sprintf("Fetching events from last %d days", days))

			now := time.Now()
			startTime := now.Add(time.Duration(-days) * 24 * time.Hour)

			resp, err := service.Events.List(calendarID).
				TimeMin(startTime.Format(time.RFC3339)).
				TimeMax(now.Format(time.RFC3339)).
				SingleEvents(true).
				OrderBy("startTime").
				MaxResults(50).
				Do()

			if err != nil {
				cli.PrintFailed()
				cli.PrintError("Failed to get events: %v", err)
				return nil
			}

			cli.PrintDone()

			if len(resp.Items) == 0 {
				cli.PrintWarning("No events in the last %d days", days)
				return nil
			}

			cli.PrintHeader(fmt.Sprintf("Recent Events (%d)", len(resp.Items)))
			fmt.Println()

			columns := []cli.Column{
				{Title: "Date", Width: 18},
				{Title: "Time", Width: 15},
				{Title: "Event", Width: 40},
			}

			var rows [][]string
			for _, event := range resp.Items {
				dateStr := ""
				timeStr := "All day"

				if event.Start.DateTime != "" {
					startTime, _ := time.Parse(time.RFC3339, event.Start.DateTime)
					dateStr = startTime.Format("Mon Jan 2")
					timeStr = startTime.Format("3:04 PM")
				} else {
					t, _ := time.Parse("2006-01-02", event.Start.Date)
					dateStr = t.Format("Mon Jan 2")
				}

				rows = append(rows, []string{
					"[D] " + dateStr,
					timeStr,
					cli.Truncate(event.Summary, 38),
				})
			}

			cli.PrintTable(columns, rows)

			return nil
		},
	}

	cmd.Flags().StringVar(&calendarID, "calendar-id", "primary", "Calendar ID")
	cmd.Flags().IntVar(&days, "days", 7, "Number of days to look back")

	return cmd
}

func calendarListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "calendars",
		Short: "List available calendars",
		Long: `List all calendars accessible to the service account.

Examples:
  vulgar gcalendar calendars`,
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := GetCalendarService()
			if err != nil {
				cli.PrintError("%v", err)
				return nil
			}

			cli.PrintLoading("Fetching calendars")

			resp, err := service.CalendarList.List().Do()

			if err != nil {
				cli.PrintFailed()
				cli.PrintError("Failed to list calendars: %v", err)
				return nil
			}

			cli.PrintDone()

			if len(resp.Items) == 0 {
				cli.PrintWarning("No calendars found")
				return nil
			}

			cli.PrintHeader(fmt.Sprintf("Calendars (%d)", len(resp.Items)))
			fmt.Println()

			columns := []cli.Column{
				{Title: "Name", Width: 35},
				{Title: "ID", Width: 40},
				{Title: "Access", Width: 12},
			}

			var rows [][]string
			for _, cal := range resp.Items {
				icon := "[C]"
				if cal.Primary {
					icon = "[*]"
				}

				rows = append(rows, []string{
					icon + " " + cli.Truncate(cal.Summary, 32),
					cli.Truncate(cal.Id, 38),
					cal.AccessRole,
				})
			}

			cli.PrintTable(columns, rows)

			cli.PrintTip("Use calendar ID: gcalendar.list_events(client, {calendar_id = \"...\"})")

			return nil
		},
	}

	return cmd
}
