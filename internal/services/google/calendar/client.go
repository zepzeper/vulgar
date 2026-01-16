package calendar

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/api/calendar/v3"

	googleauth "github.com/zepzeper/vulgar/internal/services/google"
)

type Client struct {
	svc *calendar.Service
}

type Event struct {
	ID          string
	Summary     string
	Description string
	Location    string
	Start       time.Time
	End         time.Time
	AllDay      bool
	HTMLURL     string
	Status      string
	Organizer   string
	Attendees   []string
}

type Calendar struct {
	ID          string
	Summary     string
	Description string
	TimeZone    string
	Primary     bool
}

func NewClient(ctx context.Context) (*Client, error) {
	svc, err := googleauth.NewCalendarService(ctx)
	if err != nil {
		return nil, err
	}
	return &Client{svc: svc}, nil
}

func (c *Client) ListCalendars(ctx context.Context) ([]Calendar, error) {
	result, err := c.svc.CalendarList.List().Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("list calendars failed: %w", err)
	}

	calendars := make([]Calendar, len(result.Items))
	for i, cal := range result.Items {
		calendars[i] = Calendar{
			ID:          cal.Id,
			Summary:     cal.Summary,
			Description: cal.Description,
			TimeZone:    cal.TimeZone,
			Primary:     cal.Primary,
		}
	}

	return calendars, nil
}

func (c *Client) ListEvents(ctx context.Context, calendarID string, timeMin, timeMax time.Time, maxResults int64) ([]Event, error) {
	if calendarID == "" {
		calendarID = "primary"
	}

	call := c.svc.Events.List(calendarID).
		SingleEvents(true).
		OrderBy("startTime").
		Context(ctx)

	if !timeMin.IsZero() {
		call = call.TimeMin(timeMin.Format(time.RFC3339))
	}
	if !timeMax.IsZero() {
		call = call.TimeMax(timeMax.Format(time.RFC3339))
	}
	if maxResults > 0 {
		call = call.MaxResults(maxResults)
	}

	result, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("list events failed: %w", err)
	}

	events := make([]Event, 0, len(result.Items))
	for _, e := range result.Items {
		event := Event{
			ID:          e.Id,
			Summary:     e.Summary,
			Description: e.Description,
			Location:    e.Location,
			HTMLURL:     e.HtmlLink,
			Status:      e.Status,
		}

		if e.Organizer != nil {
			event.Organizer = e.Organizer.Email
		}

		if e.Attendees != nil {
			event.Attendees = make([]string, len(e.Attendees))
			for i, a := range e.Attendees {
				event.Attendees[i] = a.Email
			}
		}

		// Parse start time
		if e.Start != nil {
			if e.Start.DateTime != "" {
				t, _ := time.Parse(time.RFC3339, e.Start.DateTime)
				event.Start = t
			} else if e.Start.Date != "" {
				t, _ := time.Parse("2006-01-02", e.Start.Date)
				event.Start = t
				event.AllDay = true
			}
		}

		// Parse end time
		if e.End != nil {
			if e.End.DateTime != "" {
				t, _ := time.Parse(time.RFC3339, e.End.DateTime)
				event.End = t
			} else if e.End.Date != "" {
				t, _ := time.Parse("2006-01-02", e.End.Date)
				event.End = t
			}
		}

		events = append(events, event)
	}

	return events, nil
}

func (c *Client) GetEvent(ctx context.Context, calendarID, eventID string) (*Event, error) {
	if calendarID == "" {
		calendarID = "primary"
	}

	e, err := c.svc.Events.Get(calendarID, eventID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get event failed: %w", err)
	}

	event := &Event{
		ID:          e.Id,
		Summary:     e.Summary,
		Description: e.Description,
		Location:    e.Location,
		HTMLURL:     e.HtmlLink,
		Status:      e.Status,
	}

	if e.Organizer != nil {
		event.Organizer = e.Organizer.Email
	}

	if e.Start != nil {
		if e.Start.DateTime != "" {
			t, _ := time.Parse(time.RFC3339, e.Start.DateTime)
			event.Start = t
		} else if e.Start.Date != "" {
			t, _ := time.Parse("2006-01-02", e.Start.Date)
			event.Start = t
			event.AllDay = true
		}
	}

	if e.End != nil {
		if e.End.DateTime != "" {
			t, _ := time.Parse(time.RFC3339, e.End.DateTime)
			event.End = t
		} else if e.End.Date != "" {
			t, _ := time.Parse("2006-01-02", e.End.Date)
			event.End = t
		}
	}

	return event, nil
}
