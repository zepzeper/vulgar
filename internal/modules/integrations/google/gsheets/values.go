package gsheets

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules/util"
	"google.golang.org/api/sheets/v4"
)

// luaGetValues reads values from a spreadsheet range
// Usage: local values, err = gsheets.get_values(client, spreadsheet_id, "Sheet1!A1:D10")
func luaGetValues(L *lua.LState) int {
	client := getClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid gsheets client")
	}

	spreadsheetID := L.CheckString(2)
	rangeSpec := L.CheckString(3)

	// Call the Sheets API
	resp, err := client.service.Spreadsheets.Values.Get(spreadsheetID, rangeSpec).Do()
	if err != nil {
		return util.PushError(L, "failed to get values: %v", err)
	}

	// Convert response to Lua table
	result := L.NewTable()
	for i, row := range resp.Values {
		rowTable := L.NewTable()
		for j, cell := range row {
			rowTable.RawSetInt(j+1, util.GoToLua(L, cell))
		}
		result.RawSetInt(i+1, rowTable)
	}

	L.Push(result)
	L.Push(lua.LNil)
	return 2
}

// luaSetValues writes values to a spreadsheet range
// Usage: local result, err = gsheets.set_values(client, spreadsheet_id, "Sheet1!A1", {{"Name", "Age"}, {"John", 30}})
func luaSetValues(L *lua.LState) int {
	client := getClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid gsheets client")
	}

	spreadsheetID := L.CheckString(2)
	rangeSpec := L.CheckString(3)
	valuesTable := L.CheckTable(4)

	// Convert Lua table to [][]interface{}
	values := util.LuaTableTo2DSlice(valuesTable)

	// Create value range
	vr := &sheets.ValueRange{
		Values: values,
	}

	// Call the Sheets API
	resp, err := client.service.Spreadsheets.Values.Update(spreadsheetID, rangeSpec, vr).
		ValueInputOption("USER_ENTERED").
		Do()
	if err != nil {
		return util.PushError(L, "failed to set values: %v", err)
	}

	// Return result info
	result := L.NewTable()
	L.SetField(result, "updated_cells", lua.LNumber(resp.UpdatedCells))
	L.SetField(result, "updated_rows", lua.LNumber(resp.UpdatedRows))
	L.SetField(result, "updated_columns", lua.LNumber(resp.UpdatedColumns))
	L.SetField(result, "updated_range", lua.LString(resp.UpdatedRange))

	L.Push(result)
	L.Push(lua.LNil)
	return 2
}

// luaAppendValues appends rows to a sheet
// Usage: local result, err = gsheets.append_values(client, spreadsheet_id, "Sheet1", {{"Jane", 25}})
func luaAppendValues(L *lua.LState) int {
	client := getClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid gsheets client")
	}

	spreadsheetID := L.CheckString(2)
	rangeSpec := L.CheckString(3)
	valuesTable := L.CheckTable(4)

	// Convert Lua table to [][]interface{}
	values := util.LuaTableTo2DSlice(valuesTable)

	// Create value range
	vr := &sheets.ValueRange{
		Values: values,
	}

	// Call the Sheets API
	resp, err := client.service.Spreadsheets.Values.Append(spreadsheetID, rangeSpec, vr).
		ValueInputOption("USER_ENTERED").
		InsertDataOption("INSERT_ROWS").
		Do()
	if err != nil {
		return util.PushError(L, "failed to append values: %v", err)
	}

	// Return result info
	result := L.NewTable()
	if resp.Updates != nil {
		L.SetField(result, "updated_cells", lua.LNumber(resp.Updates.UpdatedCells))
		L.SetField(result, "updated_rows", lua.LNumber(resp.Updates.UpdatedRows))
		L.SetField(result, "updated_range", lua.LString(resp.Updates.UpdatedRange))
	}

	L.Push(result)
	L.Push(lua.LNil)
	return 2
}

// luaClearValues clears values in a range
// Usage: local result, err = gsheets.clear_values(client, spreadsheet_id, "Sheet1!A1:D10")
func luaClearValues(L *lua.LState) int {
	client := getClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid gsheets client")
	}

	spreadsheetID := L.CheckString(2)
	rangeSpec := L.CheckString(3)

	// Call the Sheets API
	resp, err := client.service.Spreadsheets.Values.Clear(spreadsheetID, rangeSpec, &sheets.ClearValuesRequest{}).Do()
	if err != nil {
		return util.PushError(L, "failed to clear values: %v", err)
	}

	// Return cleared range
	result := L.NewTable()
	L.SetField(result, "cleared_range", lua.LString(resp.ClearedRange))

	L.Push(result)
	L.Push(lua.LNil)
	return 2
}

// luaBatchGetValues gets multiple ranges in one request
// Usage: local values, err = gsheets.batch_get_values(client, spreadsheet_id, {"Sheet1!A1:B5", "Sheet2!C1:D5"})
func luaBatchGetValues(L *lua.LState) int {
	client := getClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid gsheets client")
	}

	spreadsheetID := L.CheckString(2)
	rangesTable := L.CheckTable(3)

	// Convert ranges table to slice
	var ranges []string
	rangesTable.ForEach(func(_, val lua.LValue) {
		if s, ok := val.(lua.LString); ok {
			ranges = append(ranges, string(s))
		}
	})

	if len(ranges) == 0 {
		return util.PushError(L, "no ranges provided")
	}

	// Call the Sheets API
	resp, err := client.service.Spreadsheets.Values.BatchGet(spreadsheetID).Ranges(ranges...).Do()
	if err != nil {
		return util.PushError(L, "failed to batch get values: %v", err)
	}

	// Convert response to Lua table
	result := L.NewTable()
	for i, vr := range resp.ValueRanges {
		rangeResult := L.NewTable()
		L.SetField(rangeResult, "range", lua.LString(vr.Range))

		valuesTable := L.NewTable()
		for j, row := range vr.Values {
			rowTable := L.NewTable()
			for k, cell := range row {
				rowTable.RawSetInt(k+1, util.GoToLua(L, cell))
			}
			valuesTable.RawSetInt(j+1, rowTable)
		}
		L.SetField(rangeResult, "values", valuesTable)

		result.RawSetInt(i+1, rangeResult)
	}

	L.Push(result)
	L.Push(lua.LNil)
	return 2
}

