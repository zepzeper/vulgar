package util

import (
	lua "github.com/yuin/gopher-lua"
)

// RegisterUserDataType registers a userdata type with an optional methods table
// This is a helper for the common pattern of registering userdata types
// Usage: util.RegisterUserDataType(L, "my_type", methods)
func RegisterUserDataType(L *lua.LState, typeName string, methods map[string]lua.LGFunction) {
	mt := L.NewTypeMetatable(typeName)
	if len(methods) > 0 {
		L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), methods))
	} else {
		L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{}))
	}
}

// CheckUserData extracts a typed value from userdata
// Returns the value if it matches the expected type, otherwise returns nil
// This is a helper for the common pattern of checking userdata types
// Usage: value := util.CheckUserData[MyType](L, 1, "my_type")
func CheckUserData[T any](L *lua.LState, index int, typeName string) T {
	var zero T
	val := L.Get(index)
	if val == lua.LNil {
		return zero
	}
	ud, ok := val.(*lua.LUserData)
	if !ok {
		return zero
	}
	if v, ok := ud.Value.(T); ok {
		return v
	}
	return zero
}

// NewUserData creates a new userdata with the given value and type
// Usage: ud := util.NewUserData(L, value, "my_type")
func NewUserData(L *lua.LState, value interface{}, typeName string) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = value
	L.SetMetatable(ud, L.GetTypeMetatable(typeName))
	return ud
}
