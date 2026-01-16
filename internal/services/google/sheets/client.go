package sheets

import (
	"context"
	"fmt"

	"google.golang.org/api/sheets/v4"

	googleauth "github.com/zepzeper/vulgar/internal/services/google"
)

type Client struct {
	svc *sheets.Service
}

type Spreadsheet struct {
	ID         string
	Title      string
	SheetCount int
	Sheets     []Sheet
}

type Sheet struct {
	ID    int64
	Title string
	Index int64
}

func NewClient(ctx context.Context) (*Client, error) {
	svc, err := googleauth.NewSheetsService(ctx)
	if err != nil {
		return nil, err
	}
	return &Client{svc: svc}, nil
}

func (c *Client) GetSpreadsheet(ctx context.Context, spreadsheetID string) (*Spreadsheet, error) {
	result, err := c.svc.Spreadsheets.Get(spreadsheetID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get spreadsheet failed: %w", err)
	}

	ss := &Spreadsheet{
		ID:         result.SpreadsheetId,
		Title:      result.Properties.Title,
		SheetCount: len(result.Sheets),
		Sheets:     make([]Sheet, len(result.Sheets)),
	}

	for i, s := range result.Sheets {
		ss.Sheets[i] = Sheet{
			ID:    s.Properties.SheetId,
			Title: s.Properties.Title,
			Index: s.Properties.Index,
		}
	}

	return ss, nil
}

func (c *Client) ReadRange(ctx context.Context, spreadsheetID, rangeStr string) ([][]interface{}, error) {
	result, err := c.svc.Spreadsheets.Values.Get(spreadsheetID, rangeStr).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("read range failed: %w", err)
	}
	return result.Values, nil
}

func (c *Client) WriteRange(ctx context.Context, spreadsheetID, rangeStr string, values [][]interface{}) error {
	vr := &sheets.ValueRange{Values: values}

	_, err := c.svc.Spreadsheets.Values.Update(spreadsheetID, rangeStr, vr).
		ValueInputOption("USER_ENTERED").
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("write range failed: %w", err)
	}

	return nil
}

func (c *Client) AppendRange(ctx context.Context, spreadsheetID, rangeStr string, values [][]interface{}) error {
	vr := &sheets.ValueRange{Values: values}

	_, err := c.svc.Spreadsheets.Values.Append(spreadsheetID, rangeStr, vr).
		ValueInputOption("USER_ENTERED").
		InsertDataOption("INSERT_ROWS").
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("append range failed: %w", err)
	}

	return nil
}

func (c *Client) ClearRange(ctx context.Context, spreadsheetID, rangeStr string) error {
	_, err := c.svc.Spreadsheets.Values.Clear(spreadsheetID, rangeStr, &sheets.ClearValuesRequest{}).
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("clear range failed: %w", err)
	}

	return nil
}
