package util

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

// GoToLua converts a Go value to a Lua value
// Supports: nil, bool, float64, string, []interface{}, map[string]interface{}
// Other types are converted to string representation
func GoToLua(L *lua.LState, v interface{}) lua.LValue {
	switch val := v.(type) {
	case nil:
		return lua.LNil
	case bool:
		return lua.LBool(val)
	case float64:
		return lua.LNumber(val)
	case string:
		return lua.LString(val)
	case []interface{}:
		tbl := L.NewTable()
		for i, item := range val {
			tbl.RawSetInt(i+1, GoToLua(L, item))
		}
		return tbl
	case map[string]interface{}:
		tbl := L.NewTable()
		for k, item := range val {
			tbl.RawSetString(k, GoToLua(L, item))
		}
		return tbl
	default:
		return lua.LString(fmt.Sprintf("%v", val))
	}
}

// LuaToGo converts a Lua value to a Go value
// Converts: bool -> bool, number -> float64, string -> string
// Tables are converted to []interface{} (if array-like) or map[string]interface{} (if object-like)
// nil -> nil
func LuaToGo(v lua.LValue) interface{} {
	switch val := v.(type) {
	case lua.LBool:
		return bool(val)
	case lua.LNumber:
		return float64(val)
	case lua.LString:
		return string(val)
	case *lua.LTable:
		// Check if it's an array or object
		maxIndex := 0
		isArray := true
		val.ForEach(func(k, _ lua.LValue) {
			if num, ok := k.(lua.LNumber); ok {
				idx := int(num)
				if idx > maxIndex {
					maxIndex = idx
				}
			} else {
				isArray = false
			}
		})

		if isArray && maxIndex > 0 {
			arr := make([]interface{}, maxIndex)
			val.ForEach(func(k, v lua.LValue) {
				if num, ok := k.(lua.LNumber); ok {
					arr[int(num)-1] = LuaToGo(v)
				}
			})
			return arr
		}

		obj := make(map[string]interface{})
		val.ForEach(func(k, v lua.LValue) {
			obj[lua.LVAsString(k)] = LuaToGo(v)
		})
		return obj
	case *lua.LNilType:
		return nil
	default:
		return nil
	}
}

// LuaTableTo2DSlice converts a Lua table of tables to a 2D Go slice
// Useful for spreadsheet data, CSV rows, database result sets, etc.
// Example: {{1, "a"}, {2, "b"}} -> [][]interface{}{{1, "a"}, {2, "b"}}
func LuaTableTo2DSlice(t *lua.LTable) [][]interface{} {
	var result [][]interface{}

	t.ForEach(func(_, rowVal lua.LValue) {
		if rowTable, ok := rowVal.(*lua.LTable); ok {
			var row []interface{}
			rowTable.ForEach(func(_, cellVal lua.LValue) {
				row = append(row, LuaToGo(cellVal))
			})
			result = append(result, row)
		}
	})

	return result
}

// GoSlice2DToLuaTable converts a 2D Go slice to a Lua table of tables
// Inverse of LuaTableTo2DSlice
func GoSlice2DToLuaTable(L *lua.LState, data [][]interface{}) *lua.LTable {
	result := L.NewTable()
	for i, row := range data {
		rowTable := L.NewTable()
		for j, cell := range row {
			rowTable.RawSetInt(j+1, GoToLua(L, cell))
		}
		result.RawSetInt(i+1, rowTable)
	}
	return result
}
