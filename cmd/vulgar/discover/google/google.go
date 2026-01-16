package google

import (
	"context"

	googleauth "github.com/zepzeper/vulgar/internal/services/google"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/sheets/v4"
)

// GetDriveService creates a Google Drive service (original API)
func GetDriveService() (*drive.Service, error) {
	return googleauth.NewDriveService(context.Background())
}

// GetSheetsService creates a Google Sheets service (original API)
func GetSheetsService() (*sheets.Service, error) {
	return googleauth.NewSheetsService(context.Background())
}

// GetCalendarService creates a Google Calendar service (original API)
func GetCalendarService() (*calendar.Service, error) {
	return googleauth.NewCalendarService(context.Background())
}
