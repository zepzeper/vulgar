package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func setupLuaState() *lua.LState {
	L := lua.NewState()
	L.PreloadModule(ModuleName, Loader)
	return L
}

func TestSimpleGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	L := setupLuaState()
	defer L.Close()

	L.SetGlobal("test_url", lua.LString(server.URL))
	err := L.DoString(`
		local http = require("http")
		response, err = http.get(test_url)
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	errVal := L.GetGlobal("err")
	if errVal != lua.LNil {
		t.Fatalf("unexpected error: %v", errVal)
	}

	response := L.GetGlobal("response").(*lua.LTable)
	statusCode := L.GetField(response, "status_code").(lua.LNumber)
	body := L.GetField(response, "body").(lua.LString)

	if statusCode != 200 {
		t.Errorf("expected status 200, got %v", statusCode)
	}
	if string(body) != `{"status":"ok"}` {
		t.Errorf("unexpected body: %s", body)
	}
}

func TestSimplePost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1}`))
	}))
	defer server.Close()

	L := setupLuaState()
	defer L.Close()

	L.SetGlobal("test_url", lua.LString(server.URL))
	err := L.DoString(`
		local http = require("http")
		response, err = http.post(test_url, '{"name":"test"}')
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	response := L.GetGlobal("response").(*lua.LTable)
	statusCode := L.GetField(response, "status_code").(lua.LNumber)

	if statusCode != 201 {
		t.Errorf("expected status 201, got %v", statusCode)
	}
}

func TestSimplePut(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	L := setupLuaState()
	defer L.Close()

	L.SetGlobal("test_url", lua.LString(server.URL))
	err := L.DoString(`
		local http = require("http")
		response, err = http.put(test_url, '{"name":"updated"}')
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	errVal := L.GetGlobal("err")
	if errVal != lua.LNil {
		t.Fatalf("unexpected error: %v", errVal)
	}
}

func TestSimplePatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	L := setupLuaState()
	defer L.Close()

	L.SetGlobal("test_url", lua.LString(server.URL))
	err := L.DoString(`
		local http = require("http")
		response, err = http.patch(test_url, '{"field":"value"}')
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	errVal := L.GetGlobal("err")
	if errVal != lua.LNil {
		t.Fatalf("unexpected error: %v", errVal)
	}
}

func TestSimpleDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	L := setupLuaState()
	defer L.Close()

	L.SetGlobal("test_url", lua.LString(server.URL))
	err := L.DoString(`
		local http = require("http")
		response, err = http.delete(test_url)
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	response := L.GetGlobal("response").(*lua.LTable)
	statusCode := L.GetField(response, "status_code").(lua.LNumber)

	if statusCode != 204 {
		t.Errorf("expected status 204, got %v", statusCode)
	}
}

func TestGenericRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "OPTIONS" {
			t.Errorf("expected OPTIONS, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	L := setupLuaState()
	defer L.Close()

	L.SetGlobal("test_url", lua.LString(server.URL))
	err := L.DoString(`
		local http = require("http")
		response, err = http.request("OPTIONS", test_url)
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	errVal := L.GetGlobal("err")
	if errVal != lua.LNil {
		t.Fatalf("unexpected error: %v", errVal)
	}
}

func TestClientNew(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local http = require("http")
		client = http.new({ timeout = 30 })
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	client := L.GetGlobal("client")
	if client.Type() != lua.LTUserData {
		t.Errorf("expected userdata, got %s", client.Type())
	}
}

func TestClientGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("client get"))
	}))
	defer server.Close()

	L := setupLuaState()
	defer L.Close()

	L.SetGlobal("test_url", lua.LString(server.URL))
	err := L.DoString(`
		local http = require("http")
		local client = http.new({ timeout = 30 })
		response, err = client:get(test_url)
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	response := L.GetGlobal("response").(*lua.LTable)
	body := L.GetField(response, "body").(lua.LString)

	if string(body) != "client get" {
		t.Errorf("unexpected body: %s", body)
	}
}

func TestClientWithBaseURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/users" {
			t.Errorf("expected path /api/users, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	L := setupLuaState()
	defer L.Close()

	L.SetGlobal("base_url", lua.LString(server.URL))
	err := L.DoString(`
		local http = require("http")
		local client = http.new({ base_url = base_url })
		response, err = client:get("/api/users")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	errVal := L.GetGlobal("err")
	if errVal != lua.LNil {
		t.Fatalf("unexpected error: %v", errVal)
	}
}

func TestClientWithHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		customHeader := r.Header.Get("X-Custom-Header")
		if customHeader != "custom-value" {
			t.Errorf("expected X-Custom-Header 'custom-value', got '%s'", customHeader)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	L := setupLuaState()
	defer L.Close()

	L.SetGlobal("test_url", lua.LString(server.URL))
	err := L.DoString(`
		local http = require("http")
		local client = http.new({
			headers = { ["X-Custom-Header"] = "custom-value" }
		})
		response, err = client:get(test_url)
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}
}

func TestClientPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	L := setupLuaState()
	defer L.Close()

	L.SetGlobal("test_url", lua.LString(server.URL))
	err := L.DoString(`
		local http = require("http")
		local client = http.new({})
		response, err = client:post(test_url, '{"data":"test"}')
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	response := L.GetGlobal("response").(*lua.LTable)
	statusCode := L.GetField(response, "status_code").(lua.LNumber)

	if statusCode != 201 {
		t.Errorf("expected status 201, got %v", statusCode)
	}
}

func TestClientPut(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	L := setupLuaState()
	defer L.Close()

	L.SetGlobal("test_url", lua.LString(server.URL))
	err := L.DoString(`
		local http = require("http")
		local client = http.new({})
		response, err = client:put(test_url, '{"data":"update"}')
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	errVal := L.GetGlobal("err")
	if errVal != lua.LNil {
		t.Fatalf("unexpected error: %v", errVal)
	}
}

func TestClientDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	L := setupLuaState()
	defer L.Close()

	L.SetGlobal("test_url", lua.LString(server.URL))
	err := L.DoString(`
		local http = require("http")
		local client = http.new({})
		response, err = client:delete(test_url)
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	errVal := L.GetGlobal("err")
	if errVal != lua.LNil {
		t.Fatalf("unexpected error: %v", errVal)
	}
}

func TestClientRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "HEAD" {
			t.Errorf("expected HEAD, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	L := setupLuaState()
	defer L.Close()

	L.SetGlobal("test_url", lua.LString(server.URL))
	err := L.DoString(`
		local http = require("http")
		local client = http.new({})
		response, err = client:request("HEAD", test_url)
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	errVal := L.GetGlobal("err")
	if errVal != lua.LNil {
		t.Fatalf("unexpected error: %v", errVal)
	}
}

func TestResponseIncludesHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom-Header", "test-value")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	L := setupLuaState()
	defer L.Close()

	L.SetGlobal("test_url", lua.LString(server.URL))
	err := L.DoString(`
		local http = require("http")
		response, err = http.get(test_url)
		headers = response.headers
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	headers := L.GetGlobal("headers").(*lua.LTable)
	customHeader := L.GetField(headers, "X-Custom-Header")

	if customHeader.String() != "test-value" {
		t.Errorf("expected header value 'test-value', got '%s'", customHeader.String())
	}
}

func TestResponseIncludesStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	L := setupLuaState()
	defer L.Close()

	L.SetGlobal("test_url", lua.LString(server.URL))
	err := L.DoString(`
		local http = require("http")
		response, err = http.get(test_url)
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	response := L.GetGlobal("response").(*lua.LTable)
	status := L.GetField(response, "status").(lua.LString)

	if string(status) != "404 Not Found" {
		t.Errorf("expected status '404 Not Found', got '%s'", status)
	}
}

func TestInvalidURL(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local http = require("http")
		response, err = http.get("http://invalid.localhost.invalid:99999/path")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	response := L.GetGlobal("response")
	errVal := L.GetGlobal("err")

	if response != lua.LNil {
		t.Error("expected nil response for invalid URL")
	}
	if errVal == lua.LNil {
		t.Error("expected error for invalid URL")
	}
}

func TestClientWithBearerToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token-123" {
			t.Errorf("expected Authorization 'Bearer test-token-123', got '%s'", auth)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	L := setupLuaState()
	defer L.Close()

	L.SetGlobal("test_url", lua.LString(server.URL))
	err := L.DoString(`
		local http = require("http")
		local client = http.new({
			bearer_token = "test-token-123"
		})
		response, err = client:get(test_url)
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	errVal := L.GetGlobal("err")
	if errVal != lua.LNil {
		t.Fatalf("unexpected error: %v", errVal)
	}
}

func TestClientWithBasicAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok {
			t.Error("expected basic auth")
		}
		if user != "testuser" || pass != "testpass" {
			t.Errorf("expected testuser:testpass, got %s:%s", user, pass)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	L := setupLuaState()
	defer L.Close()

	L.SetGlobal("test_url", lua.LString(server.URL))
	err := L.DoString(`
		local http = require("http")
		local client = http.new({
			basic_auth = { user = "testuser", password = "testpass" }
		})
		response, err = client:get(test_url)
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	errVal := L.GetGlobal("err")
	if errVal != lua.LNil {
		t.Fatalf("unexpected error: %v", errVal)
	}
}

func TestLoaderReturnsModule(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule(ModuleName, Loader)
	err := L.DoString(`http = require("http")`)
	if err != nil {
		t.Fatalf("failed to require module: %v", err)
	}

	mod := L.GetGlobal("http")
	if mod.Type() != lua.LTTable {
		t.Errorf("expected table, got %s", mod.Type())
	}
}

func TestLoaderExportsAllFunctions(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule(ModuleName, Loader)
	err := L.DoString(`http = require("http")`)
	if err != nil {
		t.Fatalf("failed to require module: %v", err)
	}

	tbl := L.GetGlobal("http").(*lua.LTable)

	funcs := []string{"new", "get", "post", "put", "patch", "delete", "request"}
	for _, name := range funcs {
		fn := L.GetField(tbl, name)
		if fn.Type() != lua.LTFunction {
			t.Errorf("expected %s to be a function", name)
		}
	}
}
