package gsheets

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules/util"
	"google.golang.org/api/sheets/v4"
)

// luaGetSpreadsheet gets spreadsheet metadata
// Usage: local spreadsheet, err = gsheets.get_spreadsheet(client, spreadsheet_id)
func luaGetSpreadsheet(L *lua.LState) int {
	client := getClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid gsheets client")
	}

	spreadsheetID := L.CheckString(2)

	// Call the Sheets API
	resp, err := client.service.Spreadsheets.Get(spreadsheetID).Do()
	if err != nil {
		return util.PushError(L, "failed to get spreadsheet: %v", err)
	}

	// Return spreadsheet info
	result := L.NewTable()
	L.SetField(result, "spreadsheet_id", lua.LString(resp.SpreadsheetId))
	L.SetField(result, "title", lua.LString(resp.Properties.Title))
	L.SetField(result, "locale", lua.LString(resp.Properties.Locale))
	L.SetField(result, "time_zone", lua.LString(resp.Properties.TimeZone))
	L.SetField(result, "url", lua.LString(resp.SpreadsheetUrl))

	// Add sheets info
	sheetsTable := L.NewTable()
	for i, sheet := range resp.Sheets {
		sheetInfo := L.NewTable()
		L.SetField(sheetInfo, "sheet_id", lua.LNumber(sheet.Properties.SheetId))
		L.SetField(sheetInfo, "title", lua.LString(sheet.Properties.Title))
		L.SetField(sheetInfo, "index", lua.LNumber(sheet.Properties.Index))
		if sheet.Properties.GridProperties != nil {
			L.SetField(sheetInfo, "row_count", lua.LNumber(sheet.Properties.GridProperties.RowCount))
			L.SetField(sheetInfo, "column_count", lua.LNumber(sheet.Properties.GridProperties.ColumnCount))
		}
		sheetsTable.RawSetInt(i+1, sheetInfo)
	}
	L.SetField(result, "sheets", sheetsTable)

	L.Push(result)
	L.Push(lua.LNil)
	return 2
}

// luaCreateSpreadsheet creates a new spreadsheet
// Usage: local spreadsheet, err = gsheets.create_spreadsheet(client, {title = "My Spreadsheet"})
func luaCreateSpreadsheet(L *lua.LState) int {
	client := getClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid gsheets client")
	}

	opts := L.CheckTable(2)

	// Get title
	title := "Untitled"
	if v := opts.RawGetString("title"); v != lua.LNil {
		title = v.String()
	}

	// Create spreadsheet request
	spreadsheet := &sheets.Spreadsheet{
		Properties: &sheets.SpreadsheetProperties{
			Title: title,
		},
	}

	// Add sheets if specified
	if v := opts.RawGetString("sheets"); v != lua.LNil {
		if sheetsTable, ok := v.(*lua.LTable); ok {
			var sheetsList []*sheets.Sheet
			sheetsTable.ForEach(func(_, val lua.LValue) {
				if sheetOpts, ok := val.(*lua.LTable); ok {
					sheet := &sheets.Sheet{
						Properties: &sheets.SheetProperties{},
					}
					if t := sheetOpts.RawGetString("title"); t != lua.LNil {
						sheet.Properties.Title = t.String()
					}
					sheetsList = append(sheetsList, sheet)
				}
			})
			spreadsheet.Sheets = sheetsList
		}
	}

	// Call the Sheets API
	resp, err := client.service.Spreadsheets.Create(spreadsheet).Do()
	if err != nil {
		return util.PushError(L, "failed to create spreadsheet: %v", err)
	}

	// Return spreadsheet info
	result := L.NewTable()
	L.SetField(result, "spreadsheet_id", lua.LString(resp.SpreadsheetId))
	L.SetField(result, "title", lua.LString(resp.Properties.Title))
	L.SetField(result, "url", lua.LString(resp.SpreadsheetUrl))

	L.Push(result)
	L.Push(lua.LNil)
	return 2
}
