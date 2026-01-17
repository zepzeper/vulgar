package gsheets

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

// luaFindRow searches for a row containing a specific value
// Usage: local row_index, row_data, err = gsheets.find_row(client, spreadsheet_id, "Sheet1!A:A", "search_value")
func luaFindRow(L *lua.LState) int {
	client := getClient(L, 1)
	if client == nil {
		L.Push(lua.LNil)
		L.Push(lua.LNil)
		L.Push(lua.LString("invalid gsheets client"))
		return 3
	}

	spreadsheetID := L.CheckString(2)
	rangeSpec := L.CheckString(3)
	searchValue := L.CheckString(4)

	// Get the column data
	resp, err := client.service.Spreadsheets.Values.Get(spreadsheetID, rangeSpec).Do()
	if err != nil {
		return util.PushError(L, "failed to get values: %v", err)
	}

	// Search for the value
	for i, row := range resp.Values {
		if len(row) > 0 {
			if fmt.Sprintf("%v", row[0]) == searchValue {
				// Found! Now get the full row
				L.Push(lua.LNumber(i + 1)) // 1-indexed row number

				rowTable := L.NewTable()
				for j, cell := range row {
					rowTable.RawSetInt(j+1, util.GoToLua(L, cell))
				}
				L.Push(rowTable)
				L.Push(lua.LNil)
				return 3
			}
		}
	}

	// Not found
	L.Push(lua.LNil)
	L.Push(lua.LNil)
	L.Push(lua.LString("value not found"))
	return 3
}

// luaFindColumn finds a column by header name and returns the column letter
// Usage: local col, err = gsheets.find_column(client, spreadsheet_id, "Sheet1", "Email")
// Returns: "C" (column letter)
func luaFindColumn(L *lua.LState) int {
	client := getClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid gsheets client")
	}

	spreadsheetID := L.CheckString(2)
	sheetName := L.CheckString(3)
	headerName := L.CheckString(4)

	// Get first row (headers)
	rangeSpec := fmt.Sprintf("%s!1:1", sheetName)
	resp, err := client.service.Spreadsheets.Values.Get(spreadsheetID, rangeSpec).Do()
	if err != nil {
		return util.PushError(L, "failed to get headers: %v", err)
	}

	if len(resp.Values) == 0 || len(resp.Values[0]) == 0 {
		L.Push(lua.LNil)
		L.Push(lua.LString("no headers found"))
		return 2
	}

	for i, cell := range resp.Values[0] {
		if fmt.Sprintf("%v", cell) == headerName {
			colLetter := columnToLetter(i)
			L.Push(lua.LString(colLetter))
			L.Push(lua.LNil)
			return 2
		}
	}

	L.Push(lua.LNil)
	L.Push(lua.LString("column not found: " + headerName))
	return 2
}

// columnToLetter converts a 0-based column index to a letter (0 = A, 25 = Z, 26 = AA)
func columnToLetter(col int) string {
	result := ""
	for col >= 0 {
		result = string(rune('A'+col%26)) + result
		col = col/26 - 1
	}
	return result
}
