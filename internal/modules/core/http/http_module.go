package http

import (
	"context"
	"strings"
	"time"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/httpclient"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "http"
const luaHTTPClientTypeName = "http_client"

// luaHTTPClient wraps the httpclient.Client for Lua
type luaHTTPClient struct {
	client *httpclient.Client
}

// registerHTTPClientType registers the HTTP client userdata type
func registerHTTPClientType(L *lua.LState) {
	mt := L.NewTypeMetatable(luaHTTPClientTypeName)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), clientMethods))
}

// clientMethods are the methods available on http_client instances
var clientMethods = map[string]lua.LGFunction{
	"get":     clientGet,
	"post":    clientPost,
	"put":     clientPut,
	"patch":   clientPatch,
	"delete":  clientDelete,
	"request": clientRequest,
}

// checkHTTPClient extracts the http client from userdata
func checkHTTPClient(L *lua.LState) *luaHTTPClient {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*luaHTTPClient); ok {
		return v
	}
	L.ArgError(1, "http_client expected")
	return nil
}

// parseClientOptions extracts httpclient options from a Lua table
func parseClientOptions(L *lua.LState, tbl *lua.LTable) []httpclient.Option {
	opts := make([]httpclient.Option, 0)

	if tbl == nil {
		return opts
	}

	// Parse timeout (in seconds)
	if timeout := L.GetField(tbl, "timeout"); timeout != lua.LNil {
		if num, ok := timeout.(lua.LNumber); ok {
			opts = append(opts, httpclient.WithTimeout(time.Duration(num)*time.Second))
		}
	}

	// Parse base_url
	if baseURL := L.GetField(tbl, "base_url"); baseURL != lua.LNil {
		opts = append(opts, httpclient.WithBaseURL(lua.LVAsString(baseURL)))
	}

	// Parse follow_redirects
	if follow := L.GetField(tbl, "follow_redirects"); follow != lua.LNil {
		if !lua.LVAsBool(follow) {
			opts = append(opts, httpclient.WithNoRedirects())
		}
	}

	// Parse retry
	if retry := L.GetField(tbl, "retry"); retry != lua.LNil {
		if num, ok := retry.(lua.LNumber); ok {
			opts = append(opts, httpclient.WithRetry(int(num)))
		}
	}

	// Parse headers
	if headers := L.GetField(tbl, "headers"); headers != lua.LNil {
		if tblHeaders, ok := headers.(*lua.LTable); ok {
			headerMap := make(map[string]string)
			tblHeaders.ForEach(func(k, v lua.LValue) {
				headerMap[lua.LVAsString(k)] = lua.LVAsString(v)
			})
			opts = append(opts, httpclient.WithHeaders(headerMap))
		}
	}

	// Parse bearer token
	if token := L.GetField(tbl, "bearer_token"); token != lua.LNil {
		opts = append(opts, httpclient.WithBearerToken(lua.LVAsString(token)))
	}

	// Parse basic auth
	if auth := L.GetField(tbl, "basic_auth"); auth != lua.LNil {
		if authTbl, ok := auth.(*lua.LTable); ok {
			user := lua.LVAsString(L.GetField(authTbl, "user"))
			pass := lua.LVAsString(L.GetField(authTbl, "password"))
			if user != "" {
				opts = append(opts, httpclient.WithBasicAuth(user, pass))
			}
		}
	}

	// Parse rate limit
	if rateLimit := L.GetField(tbl, "rate_limit"); rateLimit != lua.LNil {
		if num, ok := rateLimit.(lua.LNumber); ok {
			opts = append(opts, httpclient.WithRateLimit(float64(num)))
		}
	}

	return opts
}

// buildLuaResponse creates a Lua table from an httpclient.Response
func buildLuaResponse(L *lua.LState, resp *httpclient.Response) *lua.LTable {
	tbl := L.NewTable()
	tbl.RawSetString("status_code", lua.LNumber(resp.StatusCode))
	tbl.RawSetString("status", lua.LString(resp.Status))
	tbl.RawSetString("body", lua.LString(resp.String()))

	// Add headers
	headers := L.NewTable()
	for k, v := range resp.Header {
		if len(v) > 0 {
			headers.RawSetString(k, lua.LString(v[0]))
		}
	}
	tbl.RawSetString("headers", headers)

	return tbl
}

// doRequest performs the actual HTTP request using httpclient
func doRequest(L *lua.LState, client *httpclient.Client, method, urlPath string, body string, opts *lua.LTable) int {
	ctx := context.Background()

	// Parse per-request options for additional headers or timeout
	var extraOpts []httpclient.Option
	if opts != nil {
		// Per-request timeout override
		if timeout := L.GetField(opts, "timeout"); timeout != lua.LNil {
			if num, ok := timeout.(lua.LNumber); ok {
				extraOpts = append(extraOpts, httpclient.WithTimeout(time.Duration(num)*time.Second))
			}
		}
		// Per-request headers
		if headers := L.GetField(opts, "headers"); headers != lua.LNil {
			if tblHeaders, ok := headers.(*lua.LTable); ok {
				headerMap := make(map[string]string)
				tblHeaders.ForEach(func(k, v lua.LValue) {
					headerMap[lua.LVAsString(k)] = lua.LVAsString(v)
				})
				extraOpts = append(extraOpts, httpclient.WithHeaders(headerMap))
			}
		}
	}

	// Create request client with overrides if needed
	reqClient := client
	if len(extraOpts) > 0 {
		reqClient = client.With(extraOpts...)
	}

	var resp *httpclient.Response
	var err error

	if body != "" {
		resp, err = reqClient.NewRequest(method, urlPath).
			Context(ctx).
			BodyString(body).
			Do()
	} else {
		resp, err = reqClient.Request(ctx, method, urlPath, nil)
	}

	if err != nil {
		return util.PushError(L, "request failed: %v", err)
	}

	return util.PushSuccess(L, buildLuaResponse(L, resp))
}

// =============================================================================
// Module Functions
// =============================================================================

// luaNew creates a new HTTP client with configuration
// Usage: local client = http.new({ timeout = 30, base_url = "https://api.example.com" })
func luaNew(L *lua.LState) int {
	opts := parseClientOptions(L, L.OptTable(1, nil))
	client := httpclient.New(opts...)

	luaClient := &luaHTTPClient{client: client}

	ud := L.NewUserData()
	ud.Value = luaClient
	L.SetMetatable(ud, L.GetTypeMetatable(luaHTTPClientTypeName))
	L.Push(ud)
	return 1
}

// =============================================================================
// Client Instance Methods (called with : syntax)
// =============================================================================

func clientGet(L *lua.LState) int {
	client := checkHTTPClient(L)
	url := L.CheckString(2)
	opts := L.OptTable(3, nil)
	return doRequest(L, client.client, "GET", url, "", opts)
}

func clientPost(L *lua.LState) int {
	client := checkHTTPClient(L)
	url := L.CheckString(2)
	body := L.OptString(3, "")
	opts := L.OptTable(4, nil)
	return doRequest(L, client.client, "POST", url, body, opts)
}

func clientPut(L *lua.LState) int {
	client := checkHTTPClient(L)
	url := L.CheckString(2)
	body := L.OptString(3, "")
	opts := L.OptTable(4, nil)
	return doRequest(L, client.client, "PUT", url, body, opts)
}

func clientPatch(L *lua.LState) int {
	client := checkHTTPClient(L)
	url := L.CheckString(2)
	body := L.OptString(3, "")
	opts := L.OptTable(4, nil)
	return doRequest(L, client.client, "PATCH", url, body, opts)
}

func clientDelete(L *lua.LState) int {
	client := checkHTTPClient(L)
	url := L.CheckString(2)
	opts := L.OptTable(3, nil)
	return doRequest(L, client.client, "DELETE", url, "", opts)
}

func clientRequest(L *lua.LState) int {
	client := checkHTTPClient(L)
	method := L.CheckString(2)
	url := L.CheckString(3)
	body := L.OptString(4, "")
	opts := L.OptTable(5, nil)
	return doRequest(L, client.client, strings.ToUpper(method), url, body, opts)
}

// =============================================================================
// Convenience Functions (one-off requests without creating a client)
// =============================================================================

func simpleRequest(L *lua.LState, method string) int {
	url := L.CheckString(1)
	body := ""
	var opts *lua.LTable

	if method == "POST" || method == "PUT" || method == "PATCH" {
		body = L.OptString(2, "")
		opts = L.OptTable(3, nil)
	} else {
		opts = L.OptTable(2, nil)
	}

	clientOpts := parseClientOptions(L, opts)
	client := httpclient.New(clientOpts...)

	return doRequest(L, client, method, url, body, opts)
}

func luaGet(L *lua.LState) int {
	return simpleRequest(L, "GET")
}

func luaPost(L *lua.LState) int {
	return simpleRequest(L, "POST")
}

func luaPut(L *lua.LState) int {
	return simpleRequest(L, "PUT")
}

func luaPatch(L *lua.LState) int {
	return simpleRequest(L, "PATCH")
}

func luaDelete(L *lua.LState) int {
	return simpleRequest(L, "DELETE")
}

// luaRequest is a general-purpose request function
// Usage: local resp, err = http.request("GET", url, body, { timeout = 10 })
func luaRequest(L *lua.LState) int {
	method := L.CheckString(1)
	url := L.CheckString(2)
	body := L.OptString(3, "")
	opts := L.OptTable(4, nil)

	clientOpts := parseClientOptions(L, opts)
	client := httpclient.New(clientOpts...)

	return doRequest(L, client, strings.ToUpper(method), url, body, opts)
}

// =============================================================================
// Module Registration
// =============================================================================

var exports = map[string]lua.LGFunction{
	"new":     luaNew,
	"get":     luaGet,
	"post":    luaPost,
	"put":     luaPut,
	"patch":   luaPatch,
	"delete":  luaDelete,
	"request": luaRequest,
}

// Loader is called when the module is required via require("http")
func Loader(L *lua.LState) int {
	registerHTTPClientType(L)
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

func init() {
	modules.Register(ModuleName, Loader)
}
