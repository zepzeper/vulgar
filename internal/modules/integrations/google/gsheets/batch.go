package gsheets

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules/util"
	"google.golang.org/api/sheets/v4"
)

// luaBatchUpdate performs multiple batch operations
// Usage: local result, err = gsheets.batch_update(client, spreadsheet_id, {requests = {...}})
func luaBatchUpdate(L *lua.LState) int {
	client := getClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid gsheets client")
	}

	spreadsheetID := L.CheckString(2)
	opts := L.CheckTable(3)

	// Build requests from Lua table
	var requests []*sheets.Request

	if v := opts.RawGetString("requests"); v != lua.LNil {
		if reqTable, ok := v.(*lua.LTable); ok {
			reqTable.ForEach(func(_, val lua.LValue) {
				if reqOpts, ok := val.(*lua.LTable); ok {
					req := buildRequest(reqOpts)
					if req != nil {
						requests = append(requests, req)
					}
				}
			})
		}
	}

	if len(requests) == 0 {
		return util.PushError(L, "no valid requests provided")
	}

	// Create batch update request
	batchReq := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: requests,
	}

	// Call the Sheets API
	resp, err := client.service.Spreadsheets.BatchUpdate(spreadsheetID, batchReq).Do()
	if err != nil {
		return util.PushError(L, "failed to batch update: %v", err)
	}

	// Return result
	result := L.NewTable()
	L.SetField(result, "spreadsheet_id", lua.LString(resp.SpreadsheetId))
	L.SetField(result, "replies_count", lua.LNumber(len(resp.Replies)))

	L.Push(result)
	L.Push(lua.LNil)
	return 2
}

// buildRequest builds a sheets.Request from Lua table options
func buildRequest(opts *lua.LTable) *sheets.Request {
	req := &sheets.Request{}

	// Handle different request types
	// Note: update_cells is handled via batch_update directly
	// This function handles the simpler request types

	if v := opts.RawGetString("add_sheet"); v != lua.LNil {
		if addOpts, ok := v.(*lua.LTable); ok {
			title := ""
			if t := addOpts.RawGetString("title"); t != lua.LNil {
				title = t.String()
			}
			req.AddSheet = &sheets.AddSheetRequest{
				Properties: &sheets.SheetProperties{
					Title: title,
				},
			}
			return req
		}
	}

	if v := opts.RawGetString("delete_sheet"); v != lua.LNil {
		if deleteOpts, ok := v.(*lua.LTable); ok {
			if id := deleteOpts.RawGetString("sheet_id"); id != lua.LNil {
				if num, ok := id.(lua.LNumber); ok {
					req.DeleteSheet = &sheets.DeleteSheetRequest{
						SheetId: int64(num),
					}
					return req
				}
			}
		}
	}

	if v := opts.RawGetString("rename_sheet"); v != lua.LNil {
		if renameOpts, ok := v.(*lua.LTable); ok {
			sheetID := int64(0)
			title := ""
			if id := renameOpts.RawGetString("sheet_id"); id != lua.LNil {
				if num, ok := id.(lua.LNumber); ok {
					sheetID = int64(num)
				}
			}
			if t := renameOpts.RawGetString("title"); t != lua.LNil {
				title = t.String()
			}
			req.UpdateSheetProperties = &sheets.UpdateSheetPropertiesRequest{
				Properties: &sheets.SheetProperties{
					SheetId: sheetID,
					Title:   title,
				},
				Fields: "title",
			}
			return req
		}
	}

	return nil
}

