package gsheets

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules/util"
	"google.golang.org/api/sheets/v4"
)

// luaAddSheet adds a new sheet to an existing spreadsheet
// Usage: local sheet, err = gsheets.add_sheet(client, spreadsheet_id, {title = "New Sheet"})
func luaAddSheet(L *lua.LState) int {
	client := getClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid gsheets client")
	}

	spreadsheetID := L.CheckString(2)
	opts := L.CheckTable(3)

	// Get sheet title
	title := "Sheet"
	if v := opts.RawGetString("title"); v != lua.LNil {
		title = v.String()
	}

	// Create batch update request
	req := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				AddSheet: &sheets.AddSheetRequest{
					Properties: &sheets.SheetProperties{
						Title: title,
					},
				},
			},
		},
	}

	// Call the Sheets API
	resp, err := client.service.Spreadsheets.BatchUpdate(spreadsheetID, req).Do()
	if err != nil {
		return util.PushError(L, "failed to add sheet: %v", err)
	}

	// Return sheet info
	result := L.NewTable()
	if len(resp.Replies) > 0 && resp.Replies[0].AddSheet != nil {
		props := resp.Replies[0].AddSheet.Properties
		L.SetField(result, "sheet_id", lua.LNumber(props.SheetId))
		L.SetField(result, "title", lua.LString(props.Title))
		L.SetField(result, "index", lua.LNumber(props.Index))
	}

	L.Push(result)
	L.Push(lua.LNil)
	return 2
}

// luaDeleteSheet deletes a sheet from a spreadsheet
// Usage: local err = gsheets.delete_sheet(client, spreadsheet_id, sheet_id)
func luaDeleteSheet(L *lua.LState) int {
	client := getClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid gsheets client")
	}

	spreadsheetID := L.CheckString(2)
	sheetID := L.CheckInt64(3)

	// Create batch update request
	req := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				DeleteSheet: &sheets.DeleteSheetRequest{
					SheetId: sheetID,
				},
			},
		},
	}

	// Call the Sheets API
	_, apiErr := client.service.Spreadsheets.BatchUpdate(spreadsheetID, req).Do()
	if apiErr != nil {
		return util.PushError(L, "failed to delete sheet: %v", apiErr)
	}

	L.Push(lua.LNil)
	return 1
}

// luaFindSheet finds a sheet (tab) by name within a spreadsheet
// Usage: local sheet, err = gsheets.find_sheet(client, spreadsheet_id, "Sheet Name")
func luaFindSheet(L *lua.LState) int {
	client := getClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid gsheets client")
	}

	spreadsheetID := L.CheckString(2)
	sheetName := L.CheckString(3)

	// Get spreadsheet to find the sheet
	resp, err := client.service.Spreadsheets.Get(spreadsheetID).Do()
	if err != nil {
		return util.PushError(L, "failed to get spreadsheet: %v", err)
	}

	for _, sheet := range resp.Sheets {
		if sheet.Properties.Title == sheetName {
			result := L.NewTable()
			L.SetField(result, "sheet_id", lua.LNumber(sheet.Properties.SheetId))
			L.SetField(result, "title", lua.LString(sheet.Properties.Title))
			L.SetField(result, "index", lua.LNumber(sheet.Properties.Index))
			if sheet.Properties.GridProperties != nil {
				L.SetField(result, "row_count", lua.LNumber(sheet.Properties.GridProperties.RowCount))
				L.SetField(result, "column_count", lua.LNumber(sheet.Properties.GridProperties.ColumnCount))
			}
			L.Push(result)
			L.Push(lua.LNil)
			return 2
		}
	}

	L.Push(lua.LNil)
	L.Push(lua.LString("sheet not found: " + sheetName))
	return 2
}

// luaListSheets lists all sheets (tabs) in a spreadsheet
// Usage: local sheets, err = gsheets.list_sheets(client, spreadsheet_id)
func luaListSheets(L *lua.LState) int {
	client := getClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid gsheets client")
	}

	spreadsheetID := L.CheckString(2)

	resp, err := client.service.Spreadsheets.Get(spreadsheetID).Do()
	if err != nil {
		return util.PushError(L, "failed to get spreadsheet: %v", err)
	}

	result := L.NewTable()
	for i, sheet := range resp.Sheets {
		sheetTable := L.NewTable()
		L.SetField(sheetTable, "sheet_id", lua.LNumber(sheet.Properties.SheetId))
		L.SetField(sheetTable, "title", lua.LString(sheet.Properties.Title))
		L.SetField(sheetTable, "index", lua.LNumber(sheet.Properties.Index))
		if sheet.Properties.GridProperties != nil {
			L.SetField(sheetTable, "row_count", lua.LNumber(sheet.Properties.GridProperties.RowCount))
			L.SetField(sheetTable, "column_count", lua.LNumber(sheet.Properties.GridProperties.ColumnCount))
		}
		result.RawSetInt(i+1, sheetTable)
	}

	L.Push(result)
	L.Push(lua.LNil)
	return 2
}
