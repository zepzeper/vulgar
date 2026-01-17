package gcalendar

import (
	"context"
	"sync"
	"time"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	gauth "github.com/zepzeper/vulgar/internal/modules/integrations/google"
	"github.com/zepzeper/vulgar/internal/modules/util"
	"google.golang.org/api/calendar/v3"
)

const ModuleName = "integrations.gcalendar"

type clientWrapper struct {
	service *calendar.Service
	ctx     context.Context
}

var (
	clientMutex sync.Mutex
	clients     = make(map[*lua.LUserData]*clientWrapper)
)

func registerClient(L *lua.LState, wrapper *clientWrapper) *lua.LUserData {
	clientMutex.Lock()
	defer clientMutex.Unlock()

	ud := L.NewUserData()
	ud.Value = wrapper
	clients[ud] = wrapper

	return ud
}

func getClient(L *lua.LState, idx int) *clientWrapper {
	ud := L.CheckUserData(idx)
	if ud == nil {
		return nil
	}

	wrapper, ok := ud.Value.(*clientWrapper)
	if !ok {
		return nil
	}

	return wrapper
}

// luaConfigure creates a new Google Calendar client using OAuth authentication
// Usage: local client, err = gcalendar.configure()
// Note: Requires prior authentication via 'vulgar gcalendar login'
func luaConfigure(L *lua.LState) int {
	ctx := context.Background()

	clientOpt, err := gauth.ClientOption(ctx)
	if err != nil {
		return util.PushError(L, "failed to get OAuth credentials: %v (run 'vulgar gcalendar login' first)", err)
	}

	// Create Calendar service
	service, err := calendar.NewService(ctx, clientOpt)
	if err != nil {
		return util.PushError(L, "failed to create calendar service: %v", err)
	}

	// Wrap and return client
	wrapper := &clientWrapper{
		service: service,
		ctx:     ctx,
	}

	L.Push(registerClient(L, wrapper))
	L.Push(lua.LNil)
	return 2
}

// luaListEvents lists events from a calendar
// Usage: local events, err = gcalendar.list_events(client, {calendar_id = "primary", time_min = "...", max_results = 10})
func luaListEvents(L *lua.LState) int {
	client := getClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid gcalendar client")
	}

	opts := L.OptTable(2, L.NewTable())

	// Get calendar ID
	calendarID := "primary"
	if v := opts.RawGetString("calendar_id"); v != lua.LNil {
		calendarID = v.String()
	}

	// Build request
	call := client.service.Events.List(calendarID).
		SingleEvents(true).
		OrderBy("startTime")

	// Apply filters
	if v := opts.RawGetString("time_min"); v != lua.LNil {
		call = call.TimeMin(v.String())
	} else {
		// Default to now
		call = call.TimeMin(time.Now().Format(time.RFC3339))
	}

	if v := opts.RawGetString("time_max"); v != lua.LNil {
		call = call.TimeMax(v.String())
	}

	if v := opts.RawGetString("max_results"); v != lua.LNil {
		if num, ok := v.(lua.LNumber); ok {
			call = call.MaxResults(int64(num))
		}
	} else {
		call = call.MaxResults(50)
	}

	if v := opts.RawGetString("query"); v != lua.LNil {
		call = call.Q(v.String())
	}

	// Execute request
	resp, err := call.Do()
	if err != nil {
		return util.PushError(L, "failed to list events: %v", err)
	}

	// Convert response to Lua table
	result := L.NewTable()
	for i, event := range resp.Items {
		eventTable := eventToLuaTable(L, event)
		result.RawSetInt(i+1, eventTable)
	}

	L.Push(result)
	L.Push(lua.LNil)
	return 2
}

// luaGetEvent gets a specific event
// Usage: local event, err = gcalendar.get_event(client, event_id, {calendar_id = "primary"})
func luaGetEvent(L *lua.LState) int {
	client := getClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid gcalendar client")
	}

	eventID := L.CheckString(2)
	opts := L.OptTable(3, L.NewTable())

	calendarID := "primary"
	if v := opts.RawGetString("calendar_id"); v != lua.LNil {
		calendarID = v.String()
	}

	event, err := client.service.Events.Get(calendarID, eventID).Do()
	if err != nil {
		return util.PushError(L, "failed to get event: %v", err)
	}

	L.Push(eventToLuaTable(L, event))
	L.Push(lua.LNil)
	return 2
}

// luaCreateEvent creates a new event
//
//	Usage: local event, err = gcalendar.create_event(client, {
//	  calendar_id = "primary",
//	  summary = "Meeting",
//	  start_time = "2025-01-15T10:00:00Z",
//	  end_time = "2025-01-15T11:00:00Z",
//	  attendees = {"user@example.com"}
//	})
func luaCreateEvent(L *lua.LState) int {
	client := getClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid gcalendar client")
	}

	opts := L.CheckTable(2)

	calendarID := "primary"
	if v := opts.RawGetString("calendar_id"); v != lua.LNil {
		calendarID = v.String()
	}

	event := &calendar.Event{}

	// Required fields
	if v := opts.RawGetString("summary"); v != lua.LNil {
		event.Summary = v.String()
	} else {
		return util.PushError(L, "summary is required")
	}

	// Start time
	if v := opts.RawGetString("start_time"); v != lua.LNil {
		event.Start = &calendar.EventDateTime{
			DateTime: v.String(),
		}
	} else if v := opts.RawGetString("start_date"); v != lua.LNil {
		// All-day event
		event.Start = &calendar.EventDateTime{
			Date: v.String(),
		}
	} else {
		return util.PushError(L, "start_time or start_date is required")
	}

	// End time
	if v := opts.RawGetString("end_time"); v != lua.LNil {
		event.End = &calendar.EventDateTime{
			DateTime: v.String(),
		}
	} else if v := opts.RawGetString("end_date"); v != lua.LNil {
		event.End = &calendar.EventDateTime{
			Date: v.String(),
		}
	} else {
		return util.PushError(L, "end_time or end_date is required")
	}

	// Optional fields
	if v := opts.RawGetString("description"); v != lua.LNil {
		event.Description = v.String()
	}

	if v := opts.RawGetString("location"); v != lua.LNil {
		event.Location = v.String()
	}

	if v := opts.RawGetString("timezone"); v != lua.LNil {
		tz := v.String()
		event.Start.TimeZone = tz
		event.End.TimeZone = tz
	}

	// Attendees
	if v := opts.RawGetString("attendees"); v != lua.LNil {
		if attendeesTable, ok := v.(*lua.LTable); ok {
			var attendees []*calendar.EventAttendee
			attendeesTable.ForEach(func(_, val lua.LValue) {
				if email, ok := val.(lua.LString); ok {
					attendees = append(attendees, &calendar.EventAttendee{
						Email: string(email),
					})
				}
			})
			event.Attendees = attendees
		}
	}

	// Reminders
	if v := opts.RawGetString("reminders"); v != lua.LNil {
		if remindersTable, ok := v.(*lua.LTable); ok {
			var overrides []*calendar.EventReminder
			remindersTable.ForEach(func(_, val lua.LValue) {
				if reminderTable, ok := val.(*lua.LTable); ok {
					method := "popup"
					minutes := int64(10)
					if m := reminderTable.RawGetString("method"); m != lua.LNil {
						method = m.String()
					}
					if m := reminderTable.RawGetString("minutes"); m != lua.LNil {
						if num, ok := m.(lua.LNumber); ok {
							minutes = int64(num)
						}
					}
					overrides = append(overrides, &calendar.EventReminder{
						Method:  method,
						Minutes: minutes,
					})
				}
			})
			event.Reminders = &calendar.EventReminders{
				UseDefault: false,
				Overrides:  overrides,
			}
		}
	}

	// Create the event
	createdEvent, err := client.service.Events.Insert(calendarID, event).Do()
	if err != nil {
		return util.PushError(L, "failed to create event: %v", err)
	}

	L.Push(eventToLuaTable(L, createdEvent))
	L.Push(lua.LNil)
	return 2
}

// luaUpdateEvent updates an existing event
// Usage: local event, err = gcalendar.update_event(client, event_id, {summary = "New Title", ...})
func luaUpdateEvent(L *lua.LState) int {
	client := getClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid gcalendar client")
	}

	eventID := L.CheckString(2)
	opts := L.CheckTable(3)

	calendarID := "primary"
	if v := opts.RawGetString("calendar_id"); v != lua.LNil {
		calendarID = v.String()
	}

	// Get existing event
	existingEvent, err := client.service.Events.Get(calendarID, eventID).Do()
	if err != nil {
		return util.PushError(L, "failed to get event: %v", err)
	}

	// Update fields
	if v := opts.RawGetString("summary"); v != lua.LNil {
		existingEvent.Summary = v.String()
	}

	if v := opts.RawGetString("description"); v != lua.LNil {
		existingEvent.Description = v.String()
	}

	if v := opts.RawGetString("location"); v != lua.LNil {
		existingEvent.Location = v.String()
	}

	if v := opts.RawGetString("start_time"); v != lua.LNil {
		existingEvent.Start.DateTime = v.String()
	}

	if v := opts.RawGetString("end_time"); v != lua.LNil {
		existingEvent.End.DateTime = v.String()
	}

	// Update the event
	updatedEvent, err := client.service.Events.Update(calendarID, eventID, existingEvent).Do()
	if err != nil {
		return util.PushError(L, "failed to update event: %v", err)
	}

	L.Push(eventToLuaTable(L, updatedEvent))
	L.Push(lua.LNil)
	return 2
}

// luaDeleteEvent deletes an event
// Usage: local err = gcalendar.delete_event(client, event_id, {calendar_id = "primary"})
func luaDeleteEvent(L *lua.LState) int {
	client := getClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid gcalendar client")
	}

	eventID := L.CheckString(2)
	opts := L.OptTable(3, L.NewTable())

	calendarID := "primary"
	if v := opts.RawGetString("calendar_id"); v != lua.LNil {
		calendarID = v.String()
	}

	err := client.service.Events.Delete(calendarID, eventID).Do()
	if err != nil {
		return util.PushError(L, "failed to delete event: %v", err)
	}

	L.Push(lua.LNil)
	return 1
}

// luaListCalendars lists all calendars
// Usage: local calendars, err = gcalendar.list_calendars(client)
func luaListCalendars(L *lua.LState) int {
	client := getClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid gcalendar client")
	}

	resp, err := client.service.CalendarList.List().Do()
	if err != nil {
		return util.PushError(L, "failed to list calendars: %v", err)
	}

	result := L.NewTable()
	for i, cal := range resp.Items {
		calTable := L.NewTable()
		L.SetField(calTable, "id", lua.LString(cal.Id))
		L.SetField(calTable, "summary", lua.LString(cal.Summary))
		L.SetField(calTable, "description", lua.LString(cal.Description))
		L.SetField(calTable, "timezone", lua.LString(cal.TimeZone))
		L.SetField(calTable, "primary", lua.LBool(cal.Primary))
		L.SetField(calTable, "access_role", lua.LString(cal.AccessRole))
		result.RawSetInt(i+1, calTable)
	}

	L.Push(result)
	L.Push(lua.LNil)
	return 2
}

// luaQuickAdd quickly adds an event using natural language
// Usage: local event, err = gcalendar.quick_add(client, "Meeting tomorrow at 10am", {calendar_id = "primary"})
func luaQuickAdd(L *lua.LState) int {
	client := getClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid gcalendar client")
	}

	text := L.CheckString(2)
	opts := L.OptTable(3, L.NewTable())

	calendarID := "primary"
	if v := opts.RawGetString("calendar_id"); v != lua.LNil {
		calendarID = v.String()
	}

	event, err := client.service.Events.QuickAdd(calendarID, text).Do()
	if err != nil {
		return util.PushError(L, "failed to quick add event: %v", err)
	}

	L.Push(eventToLuaTable(L, event))
	L.Push(lua.LNil)
	return 2
}

// luaFreebusy checks free/busy information
//
//	Usage: local busy, err = gcalendar.freebusy(client, {
//	  time_min = "...",
//	  time_max = "...",
//	  calendars = {"primary", "other@example.com"}
//	})
func luaFreebusy(L *lua.LState) int {
	client := getClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid gcalendar client")
	}

	opts := L.CheckTable(2)

	timeMin := ""
	if v := opts.RawGetString("time_min"); v != lua.LNil {
		timeMin = v.String()
	} else {
		timeMin = time.Now().Format(time.RFC3339)
	}

	timeMax := ""
	if v := opts.RawGetString("time_max"); v != lua.LNil {
		timeMax = v.String()
	} else {
		timeMax = time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	}

	var items []*calendar.FreeBusyRequestItem
	if v := opts.RawGetString("calendars"); v != lua.LNil {
		if calsTable, ok := v.(*lua.LTable); ok {
			calsTable.ForEach(func(_, val lua.LValue) {
				if calID, ok := val.(lua.LString); ok {
					items = append(items, &calendar.FreeBusyRequestItem{Id: string(calID)})
				}
			})
		}
	} else {
		items = append(items, &calendar.FreeBusyRequestItem{Id: "primary"})
	}

	req := &calendar.FreeBusyRequest{
		TimeMin: timeMin,
		TimeMax: timeMax,
		Items:   items,
	}

	resp, err := client.service.Freebusy.Query(req).Do()
	if err != nil {
		return util.PushError(L, "failed to query freebusy: %v", err)
	}

	result := L.NewTable()
	for calID, cal := range resp.Calendars {
		calTable := L.NewTable()
		busyTable := L.NewTable()
		for i, busy := range cal.Busy {
			busyPeriod := L.NewTable()
			L.SetField(busyPeriod, "start", lua.LString(busy.Start))
			L.SetField(busyPeriod, "end", lua.LString(busy.End))
			busyTable.RawSetInt(i+1, busyPeriod)
		}
		L.SetField(calTable, "busy", busyTable)
		L.SetField(result, calID, calTable)
	}

	L.Push(result)
	L.Push(lua.LNil)
	return 2
}

// Helper function to convert event to Lua table
func eventToLuaTable(L *lua.LState, event *calendar.Event) *lua.LTable {
	result := L.NewTable()
	L.SetField(result, "id", lua.LString(event.Id))
	L.SetField(result, "summary", lua.LString(event.Summary))
	L.SetField(result, "description", lua.LString(event.Description))
	L.SetField(result, "location", lua.LString(event.Location))
	L.SetField(result, "status", lua.LString(event.Status))
	L.SetField(result, "html_link", lua.LString(event.HtmlLink))

	if event.Start != nil {
		if event.Start.DateTime != "" {
			L.SetField(result, "start_time", lua.LString(event.Start.DateTime))
		} else {
			L.SetField(result, "start_date", lua.LString(event.Start.Date))
		}
	}

	if event.End != nil {
		if event.End.DateTime != "" {
			L.SetField(result, "end_time", lua.LString(event.End.DateTime))
		} else {
			L.SetField(result, "end_date", lua.LString(event.End.Date))
		}
	}

	if event.Organizer != nil {
		L.SetField(result, "organizer", lua.LString(event.Organizer.Email))
	}

	if len(event.Attendees) > 0 {
		attendees := L.NewTable()
		for i, att := range event.Attendees {
			attTable := L.NewTable()
			L.SetField(attTable, "email", lua.LString(att.Email))
			L.SetField(attTable, "response_status", lua.LString(att.ResponseStatus))
			L.SetField(attTable, "optional", lua.LBool(att.Optional))
			attendees.RawSetInt(i+1, attTable)
		}
		L.SetField(result, "attendees", attendees)
	}

	L.SetField(result, "created", lua.LString(event.Created))
	L.SetField(result, "updated", lua.LString(event.Updated))

	return result
}

// luaFindCalendar finds a calendar by name
// Usage: local calendar, err = gcalendar.find_calendar(client, "Work Calendar")
func luaFindCalendar(L *lua.LState) int {
	client := getClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid gcalendar client")
	}

	calendarName := L.CheckString(2)

	resp, err := client.service.CalendarList.List().Do()
	if err != nil {
		return util.PushError(L, "failed to list calendars: %v", err)
	}

	for _, cal := range resp.Items {
		if cal.Summary == calendarName {
			result := L.NewTable()
			L.SetField(result, "id", lua.LString(cal.Id))
			L.SetField(result, "summary", lua.LString(cal.Summary))
			L.SetField(result, "description", lua.LString(cal.Description))
			L.SetField(result, "timezone", lua.LString(cal.TimeZone))
			L.SetField(result, "primary", lua.LBool(cal.Primary))
			L.Push(result)
			L.Push(lua.LNil)
			return 2
		}
	}

	L.Push(lua.LNil)
	L.Push(lua.LString("calendar not found: " + calendarName))
	return 2
}

// luaFindEvent finds an event by title within a time range
// Usage: local events, err = gcalendar.find_event(client, "Team Meeting", {calendar_id = "primary", time_min = "..."})
func luaFindEvent(L *lua.LState) int {
	client := getClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid gcalendar client")
	}

	eventTitle := L.CheckString(2)
	opts := L.OptTable(3, L.NewTable())

	calendarID := "primary"
	if v := opts.RawGetString("calendar_id"); v != lua.LNil {
		calendarID = v.String()
	}

	call := client.service.Events.List(calendarID).
		Q(eventTitle).
		SingleEvents(true).
		OrderBy("startTime")

	if v := opts.RawGetString("time_min"); v != lua.LNil {
		call = call.TimeMin(v.String())
	} else {
		call = call.TimeMin(time.Now().Format(time.RFC3339))
	}

	if v := opts.RawGetString("time_max"); v != lua.LNil {
		call = call.TimeMax(v.String())
	}

	if v := opts.RawGetString("max_results"); v != lua.LNil {
		if num, ok := v.(lua.LNumber); ok {
			call = call.MaxResults(int64(num))
		}
	} else {
		call = call.MaxResults(10)
	}

	resp, err := call.Do()
	if err != nil {
		return util.PushError(L, "failed to find events: %v", err)
	}

	if len(resp.Items) == 0 {
		L.Push(lua.LNil)
		L.Push(lua.LString("no events found matching: " + eventTitle))
		return 2
	}

	// Return all matching events
	result := L.NewTable()
	for i, event := range resp.Items {
		result.RawSetInt(i+1, eventToLuaTable(L, event))
	}

	L.Push(result)
	L.Push(lua.LNil)
	return 2
}

// luaTodaysEvents gets all events for today
// Usage: local events, err = gcalendar.todays_events(client, {calendar_id = "primary"})
func luaTodaysEvents(L *lua.LState) int {
	client := getClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid gcalendar client")
	}

	opts := L.OptTable(2, L.NewTable())

	calendarID := "primary"
	if v := opts.RawGetString("calendar_id"); v != lua.LNil {
		calendarID = v.String()
	}

	// Get start and end of today
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	resp, err := client.service.Events.List(calendarID).
		TimeMin(startOfDay.Format(time.RFC3339)).
		TimeMax(endOfDay.Format(time.RFC3339)).
		SingleEvents(true).
		OrderBy("startTime").
		Do()
	if err != nil {
		return util.PushError(L, "failed to get today's events: %v", err)
	}

	result := L.NewTable()
	for i, event := range resp.Items {
		result.RawSetInt(i+1, eventToLuaTable(L, event))
	}

	L.Push(result)
	L.Push(lua.LNil)
	return 2
}

// luaUpcomingEvents gets upcoming events
// Usage: local events, err = gcalendar.upcoming_events(client, {days = 7, calendar_id = "primary"})
func luaUpcomingEvents(L *lua.LState) int {
	client := getClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid gcalendar client")
	}

	opts := L.OptTable(2, L.NewTable())

	calendarID := "primary"
	if v := opts.RawGetString("calendar_id"); v != lua.LNil {
		calendarID = v.String()
	}

	days := 7
	if v := opts.RawGetString("days"); v != lua.LNil {
		if num, ok := v.(lua.LNumber); ok {
			days = int(num)
		}
	}

	maxResults := int64(50)
	if v := opts.RawGetString("max_results"); v != lua.LNil {
		if num, ok := v.(lua.LNumber); ok {
			maxResults = int64(num)
		}
	}

	now := time.Now()
	endTime := now.Add(time.Duration(days) * 24 * time.Hour)

	resp, err := client.service.Events.List(calendarID).
		TimeMin(now.Format(time.RFC3339)).
		TimeMax(endTime.Format(time.RFC3339)).
		SingleEvents(true).
		OrderBy("startTime").
		MaxResults(maxResults).
		Do()
	if err != nil {
		return util.PushError(L, "failed to get upcoming events: %v", err)
	}

	result := L.NewTable()
	for i, event := range resp.Items {
		result.RawSetInt(i+1, eventToLuaTable(L, event))
	}

	L.Push(result)
	L.Push(lua.LNil)
	return 2
}

var exports = map[string]lua.LGFunction{
	"configure":       luaConfigure,
	"list_events":     luaListEvents,
	"get_event":       luaGetEvent,
	"create_event":    luaCreateEvent,
	"update_event":    luaUpdateEvent,
	"delete_event":    luaDeleteEvent,
	"list_calendars":  luaListCalendars,
	"quick_add":       luaQuickAdd,
	"freebusy":        luaFreebusy,
	"find_calendar":   luaFindCalendar,
	"find_event":      luaFindEvent,
	"todays_events":   luaTodaysEvents,
	"upcoming_events": luaUpcomingEvents,
}

// Loader is called when the module is required via require("integrations.gcalendar")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
