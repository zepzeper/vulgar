package csv

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "stdlib.csv"

// parseOptions extracts options from optional Lua table
func parseOptions(L *lua.LState, index int) (header bool, delimiter rune) {
	header = false
	delimiter = ','

	if L.GetTop() >= index && L.Get(index) != lua.LNil {
		opts := L.CheckTable(index)
		if headerVal := opts.RawGetString("header"); headerVal != lua.LNil {
			if lv, ok := headerVal.(lua.LBool); ok {
				header = bool(lv)
			}
		}
		if delimVal := opts.RawGetString("delimiter"); delimVal != lua.LNil {
			delimStr := lua.LVAsString(delimVal)
			if len(delimStr) > 0 {
				delimiter = rune(delimStr[0])
			}
		}
	}

	return header, delimiter
}

// luaParse parses a CSV string into a table
// Usage: local rows, err = csv.parse(csv_string, {header = true, delimiter = ","})
func luaParse(L *lua.LState) int {
	csvString := L.CheckString(1)
	header, delimiter := parseOptions(L, 2)

	reader := csv.NewReader(strings.NewReader(csvString))
	reader.Comma = delimiter

	records, err := reader.ReadAll()
	if err != nil {
		return util.PushError(L, "csv parse error: %v", err)
	}

	if len(records) == 0 {
		return util.PushSuccess(L, L.NewTable())
	}

	result := L.NewTable()

	if header && len(records) > 0 {
		// First row is header, rest are data rows
		headers := records[0]
		for i := 1; i < len(records); i++ {
			row := L.NewTable()
			for j, headerName := range headers {
				if j < len(records[i]) {
					row.RawSetString(headerName, lua.LString(records[i][j]))
				}
			}
			result.Append(row)
		}
	} else {
		// All rows are arrays
		for _, record := range records {
			row := L.NewTable()
			for j, field := range record {
				row.RawSetInt(j+1, lua.LString(field))
			}
			result.Append(row)
		}
	}

	return util.PushSuccess(L, result)
}

// luaEncode encodes a table into a CSV string
// Usage: local csv_string, err = csv.encode(rows, {header = true, delimiter = ","})
func luaEncode(L *lua.LState) int {
	rowsTable := L.CheckTable(1)
	header, delimiter := parseOptions(L, 2)

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	writer.Comma = delimiter

	var records [][]string

	// Convert Lua table to Go records
	rowsTable.ForEach(func(key, value lua.LValue) {
		rowTable, ok := value.(*lua.LTable)
		if !ok {
			return
		}

		var row []string
		if header {
			// Object-like table (map)
			rowMap := make(map[string]string)
			rowTable.ForEach(func(k, v lua.LValue) {
				rowMap[lua.LVAsString(k)] = lua.LVAsString(v)
			})
			// For header mode, we need to collect all keys first
			// This is a simplified version - assumes all rows have same keys
			if len(records) == 0 {
				// First row determines headers
				headers := make([]string, 0)
				rowTable.ForEach(func(k, _ lua.LValue) {
					headers = append(headers, lua.LVAsString(k))
				})
				records = append(records, headers)
			}
			// Build row in header order
			headerRow := records[0]
			row = make([]string, len(headerRow))
			for i, h := range headerRow {
				if val, ok := rowMap[h]; ok {
					row[i] = val
				}
			}
		} else {
			// Array-like table
			rowTable.ForEach(func(k, v lua.LValue) {
				if _, ok := k.(lua.LNumber); ok {
					row = append(row, lua.LVAsString(v))
				}
			})
		}
		records = append(records, row)
	})

	if err := writer.WriteAll(records); err != nil {
		return util.PushError(L, "csv encode error: %v", err)
	}
	writer.Flush()

	return util.PushSuccess(L, lua.LString(buf.String()))
}

// luaReadFile reads and parses a CSV file
// Usage: local rows, err = csv.read_file(path, {header = true, delimiter = ","})
func luaReadFile(L *lua.LState) int {
	filePath := L.CheckString(1)
	header, delimiter := parseOptions(L, 2)

	content, err := os.ReadFile(filePath)
	if err != nil {
		return util.PushError(L, "file read error: %v", err)
	}

	reader := csv.NewReader(bytes.NewReader(content))
	reader.Comma = delimiter

	records, err := reader.ReadAll()
	if err != nil {
		return util.PushError(L, "csv parse error: %v", err)
	}

	if len(records) == 0 {
		return util.PushSuccess(L, L.NewTable())
	}

	result := L.NewTable()

	if header && len(records) > 0 {
		// First row is header, rest are data rows
		headers := records[0]
		for i := 1; i < len(records); i++ {
			row := L.NewTable()
			for j, headerName := range headers {
				if j < len(records[i]) {
					row.RawSetString(headerName, lua.LString(records[i][j]))
				}
			}
			result.Append(row)
		}
	} else {
		// All rows are arrays
		for _, record := range records {
			row := L.NewTable()
			for j, field := range record {
				row.RawSetInt(j+1, lua.LString(field))
			}
			result.Append(row)
		}
	}

	return util.PushSuccess(L, result)
}

// luaWriteFile writes a table to a CSV file
// Usage: local err = csv.write_file(path, rows, {header = true, delimiter = ","})
func luaWriteFile(L *lua.LState) int {
	filePath := L.CheckString(1)
	rowsTable := L.CheckTable(2)
	header, delimiter := parseOptions(L, 3)

	var records [][]string

	// Convert Lua table to Go records
	rowsTable.ForEach(func(key, value lua.LValue) {
		rowTable, ok := value.(*lua.LTable)
		if !ok {
			return
		}

		var row []string
		if header {
			// Object-like table (map)
			rowMap := make(map[string]string)
			rowTable.ForEach(func(k, v lua.LValue) {
				rowMap[lua.LVAsString(k)] = lua.LVAsString(v)
			})
			// For header mode, we need to collect all keys first
			if len(records) == 0 {
				// First row determines headers
				headers := make([]string, 0)
				rowTable.ForEach(func(k, _ lua.LValue) {
					headers = append(headers, lua.LVAsString(k))
				})
				records = append(records, headers)
			}
			// Build row in header order
			headerRow := records[0]
			row = make([]string, len(headerRow))
			for i, h := range headerRow {
				if val, ok := rowMap[h]; ok {
					row[i] = val
				}
			}
		} else {
			// Array-like table
			rowTable.ForEach(func(k, v lua.LValue) {
				if _, ok := k.(lua.LNumber); ok {
					row = append(row, lua.LVAsString(v))
				}
			})
		}
		records = append(records, row)
	})

	file, err := os.Create(filePath)
	if err != nil {
		L.Push(lua.LString(fmt.Sprintf("file create error: %v", err)))
		return 1
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	writer.Comma = delimiter

	if err := writer.WriteAll(records); err != nil {
		L.Push(lua.LString(fmt.Sprintf("csv write error: %v", err)))
		return 1
	}
	writer.Flush()

	if err := writer.Error(); err != nil {
		L.Push(lua.LString(fmt.Sprintf("csv flush error: %v", err)))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

var exports = map[string]lua.LGFunction{
	"parse":      luaParse,
	"encode":     luaEncode,
	"read_file":  luaReadFile,
	"write_file": luaWriteFile,
}

// Loader is called when the module is required via require("csv")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
