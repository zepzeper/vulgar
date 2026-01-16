package url

import (
	"net/url"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "stdlib.url"

// luaParse parses a URL into components
// Usage: local parts, err = url.parse("https://user:pass@example.com:8080/path?query=1#hash")
func luaParse(L *lua.LState) int {
	inputURL := L.CheckString(1)

	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		return util.PushError(L, "%v", err)
	}

	tbl := L.NewTable()

	// Set scheme
	if parsedURL.Scheme != "" {
		tbl.RawSetString("scheme", lua.LString(parsedURL.Scheme))
	}

	// Set host (without port)
	if parsedURL.Hostname() != "" {
		tbl.RawSetString("host", lua.LString(parsedURL.Hostname()))
	}

	// Set port
	if port := parsedURL.Port(); port != "" {
		tbl.RawSetString("port", lua.LString(port))
	}

	// Set path
	if parsedURL.Path != "" {
		tbl.RawSetString("path", lua.LString(parsedURL.Path))
	}

	// Set query string (raw query)
	if parsedURL.RawQuery != "" {
		tbl.RawSetString("query", lua.LString(parsedURL.RawQuery))
	}

	// Set fragment
	if parsedURL.Fragment != "" {
		tbl.RawSetString("fragment", lua.LString(parsedURL.Fragment))
	}

	// Extract user and password from UserInfo
	if parsedURL.User != nil {
		if username := parsedURL.User.Username(); username != "" {
			tbl.RawSetString("user", lua.LString(username))
		}
		if password, ok := parsedURL.User.Password(); ok {
			tbl.RawSetString("password", lua.LString(password))
		}
	}

	// Return table and nil error
	return util.PushSuccess(L, tbl)
}

// luaBuild builds a URL from components
// Usage: local u = url.build({scheme = "https", host = "example.com", path = "/api", query = {foo = "bar"}})
func luaBuild(L *lua.LState) int {
	tbl := L.CheckTable(1)

	u := url.URL{}

	if scheme := tbl.RawGetString("scheme"); scheme != nil {
		if str, ok := scheme.(lua.LString); ok {
			u.Scheme = string(str)
		}
	}

	host := ""
	port := ""
	if hostVal := tbl.RawGetString("host"); hostVal != nil {
		if str, ok := hostVal.(lua.LString); ok {
			host = string(str)
		}
	}

	if portVal := tbl.RawGetString("port"); portVal != nil {
		if str, ok := portVal.(lua.LString); ok {
			port = string(str)
		}
	}

	if port != "" {
		u.Host = host + ":" + port
	} else {
		u.Host = host
	}

	if pathVal := tbl.RawGetString("path"); pathVal != nil {
		if str, ok := pathVal.(lua.LString); ok {
			u.Path = string(str)
		}
	}

	if queryVal := tbl.RawGetString("query"); queryVal != nil {
		switch v := queryVal.(type) {
		case lua.LString:
			u.RawQuery = string(v)
		case *lua.LTable:
			values := url.Values{}
			v.ForEach(func(key, value lua.LValue) {
				keyStr := ""
				if k, ok := key.(lua.LString); ok {
					keyStr = string(k)
				} else if k, ok := key.(lua.LNumber); ok {
					keyStr = k.String()
				}

				valueStr := ""
				if val, ok := value.(lua.LString); ok {
					valueStr = string(val)
				} else if val, ok := value.(lua.LNumber); ok {
					valueStr = val.String()
				}

				if keyStr != "" {
					values.Add(keyStr, valueStr)
				}
			})
			u.RawQuery = values.Encode()
		}
	}

	if fragmentVal := tbl.RawGetString("fragment"); fragmentVal != lua.LNil {
		if str, ok := fragmentVal.(lua.LString); ok {
			u.Fragment = string(str)
		}
	}

	username := ""
	password := ""
	if userVal := tbl.RawGetString("user"); userVal != lua.LNil {
		if str, ok := userVal.(lua.LString); ok {
			username = string(str)
		}
	}
	if passVal := tbl.RawGetString("password"); passVal != lua.LNil {
		if str, ok := passVal.(lua.LString); ok {
			password = string(str)
		}
	}

	if username != "" {
		if password != "" {
			u.User = url.UserPassword(username, password)
		} else {
			u.User = url.User(username)
		}
	}

	L.Push(lua.LString(u.String()))
	return 1
}

// luaEncode URL-encodes a string
// Usage: local encoded = url.encode("hello world")
func luaEncode(L *lua.LState) int {
	inputURL := L.CheckString(1)
	L.Push(lua.LString(url.QueryEscape(inputURL)))
	return 1
}

// luaDecode URL-decodes a string
// Usage: local decoded, err = url.decode("hello%20world")
func luaDecode(L *lua.LState) int {
	inputURL := L.CheckString(1)
	decoded, err := url.QueryUnescape(inputURL)

	if err != nil {
		return util.PushError(L, "%v", err)
	}

	return util.PushSuccess(L, lua.LString(decoded))
}

// luaQueryEncode encodes a table as query string
// Usage: local qs = url.query_encode({foo = "bar", baz = "qux"})
func luaQueryEncode(L *lua.LState) int {
	tbl := L.CheckTable(1)

	values := url.Values{}

	// Iterate over the table and add key-value pairs
	tbl.ForEach(func(key, value lua.LValue) {
		// Convert key to string
		keyStr := ""
		if k, ok := key.(lua.LString); ok {
			keyStr = string(k)
		} else if k, ok := key.(lua.LNumber); ok {
			keyStr = k.String()
		}

		// Convert value to string
		valueStr := ""
		if val, ok := value.(lua.LString); ok {
			valueStr = string(val)
		} else if val, ok := value.(lua.LNumber); ok {
			valueStr = val.String()
		} else if val, ok := value.(lua.LBool); ok {
			if val {
				valueStr = "true"
			} else {
				valueStr = "false"
			}
		}

		if keyStr != "" {
			values.Add(keyStr, valueStr)
		}
	})

	// Encode to query string
	queryString := values.Encode()
	L.Push(lua.LString(queryString))
	return 1
}

// luaQueryDecode decodes a query string into a table
// Usage: local params, err = url.query_decode("foo=bar&baz=qux")
func luaQueryDecode(L *lua.LState) int {
	queryString := L.CheckString(1)

	// Parse the query string
	values, err := url.ParseQuery(queryString)
	if err != nil {
		return util.PushError(L, "%v", err)
	}

	// Create a Lua table
	tbl := L.NewTable()

	// Add each key-value pair to the table
	for key, valSlice := range values {
		// If multiple values for same key, use the first one
		// (or you could make it an array - but test expects single value)
		if len(valSlice) > 0 {
			tbl.RawSetString(key, lua.LString(valSlice[0]))
		}
	}

	return util.PushSuccess(L, tbl)
}

// luaJoin joins URL parts
// Usage: local u = url.join("https://example.com", "api", "v1", "users")
func luaJoin(L *lua.LState) int {
	// Get number of arguments
	n := L.GetTop()
	if n == 0 {
		L.Push(lua.LString(""))
		return 1
	}

	// Collect all string arguments
	var parts []string
	for i := 1; i <= n; i++ {
		if str, ok := L.Get(i).(lua.LString); ok {
			parts = append(parts, string(str))
		}
	}

	if len(parts) == 0 {
		L.Push(lua.LString(""))
		return 1
	}

	// url.JoinPath requires at least one argument, then variadic
	// So we pass first element, then the rest
	var result string
	if len(parts) == 1 {
		result = parts[0]
	} else {
		result, _ = url.JoinPath(parts[0], parts[1:]...)
	}

	L.Push(lua.LString(result))
	return 1
}

// luaResolve resolves a relative URL against a base
// Usage: local u = url.resolve("https://example.com/api/", "../users")
func luaResolve(L *lua.LState) int {
	baseStr := L.CheckString(1)
	refStr := L.CheckString(2)

	// Parse the base URL
	baseURL, err := url.Parse(baseStr)
	if err != nil {
		L.Push(lua.LString(""))
		return 1
	}

	// Parse the reference URL (relative)
	refURL, err := url.Parse(refStr)
	if err != nil {
		L.Push(lua.LString(""))
		return 1
	}

	// Resolve the reference against the base
	resolved := baseURL.ResolveReference(refURL)

	L.Push(lua.LString(resolved.String()))
	return 1
}

var exports = map[string]lua.LGFunction{
	"parse":        luaParse,
	"build":        luaBuild,
	"encode":       luaEncode,
	"decode":       luaDecode,
	"query_encode": luaQueryEncode,
	"query_decode": luaQueryDecode,
	"join":         luaJoin,
	"resolve":      luaResolve,
}

// Loader is called when the module is required via require("url")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
