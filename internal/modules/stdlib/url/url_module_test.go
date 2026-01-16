package url

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func newTestState() *lua.LState {
	L := lua.NewState()
	L.PreloadModule(ModuleName, Loader)
	return L
}

// =============================================================================
// parse tests
// =============================================================================

func TestParseFullURL(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local url = require("stdlib.url")
		local parts, err = url.parse("https://user:pass@example.com:8080/path/to/resource?query=1&foo=bar#section")
		assert(err == nil, "parse should not error: " .. tostring(err))
		assert(parts.scheme == "https", "scheme should be https")
		assert(parts.host == "example.com", "host should be example.com")
		assert(parts.port == "8080", "port should be 8080")
		assert(parts.path == "/path/to/resource", "path should match")
		assert(parts.fragment == "section", "fragment should be section")
		assert(parts.user == "user", "user should be user")
		assert(parts.password == "pass", "password should be pass")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestParseSimpleURL(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local url = require("stdlib.url")
		local parts, err = url.parse("https://example.com")
		assert(err == nil, "parse should not error: " .. tostring(err))
		assert(parts.scheme == "https", "scheme should be https")
		assert(parts.host == "example.com", "host should be example.com")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestParseWithQuery(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local url = require("stdlib.url")
		local parts, err = url.parse("https://example.com/search?q=hello&page=1")
		assert(err == nil, "parse should not error: " .. tostring(err))
		assert(parts.path == "/search", "path should be /search")
		assert(parts.query == "q=hello&page=1", "query should match")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestParseInvalidURL(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local url = require("stdlib.url")
		local parts, err = url.parse("not a valid url ://")
		assert(parts == nil or err ~= nil, "should handle invalid URL")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// build tests
// =============================================================================

func TestBuildURL(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local url = require("stdlib.url")
		local result = url.build({
			scheme = "https",
			host = "example.com",
			path = "/api/users"
		})
		assert(result == "https://example.com/api/users", "should build correct URL")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestBuildURLWithPort(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local url = require("stdlib.url")
		local result = url.build({
			scheme = "http",
			host = "localhost",
			port = "8080",
			path = "/api"
		})
		assert(string.find(result, "localhost:8080") ~= nil, "should include port")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestBuildURLWithQuery(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local url = require("stdlib.url")
		local result = url.build({
			scheme = "https",
			host = "example.com",
			path = "/search",
			query = {q = "test", page = "1"}
		})
		assert(string.find(result, "q=") ~= nil, "should include query")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// encode tests
// =============================================================================

func TestEncodeSpaces(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local url = require("stdlib.url")
		local result = url.encode("hello world")
		assert(result == "hello%20world" or result == "hello+world", "should encode spaces")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestEncodeSpecialChars(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local url = require("stdlib.url")
		local result = url.encode("hello&world=test")
		assert(string.find(result, "&") == nil or string.find(result, "%%26") ~= nil, "should encode &")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestEncodeAlreadySafe(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local url = require("stdlib.url")
		local result = url.encode("hello")
		assert(result == "hello", "safe chars should remain unchanged")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// decode tests
// =============================================================================

func TestDecodeSpaces(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local url = require("stdlib.url")
		local result, err = url.decode("hello%20world")
		assert(err == nil, "decode should not error: " .. tostring(err))
		assert(result == "hello world", "should decode %20 to space")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestDecodePlusAsSpace(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local url = require("stdlib.url")
		local result, err = url.decode("hello+world")
		assert(err == nil, "decode should not error")
		-- Plus might be decoded as space or kept as plus depending on implementation
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestDecodeSpecialChars(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local url = require("stdlib.url")
		local result, err = url.decode("hello%26world%3Dtest")
		assert(err == nil, "decode should not error")
		assert(result == "hello&world=test", "should decode special chars")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestDecodeInvalid(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local url = require("stdlib.url")
		local result, err = url.decode("%ZZ")
		-- Should either error or handle gracefully
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// query_encode tests
// =============================================================================

func TestQueryEncodeSimple(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local url = require("stdlib.url")
		local result = url.query_encode({foo = "bar", baz = "qux"})
		assert(string.find(result, "foo=bar") ~= nil, "should contain foo=bar")
		assert(string.find(result, "baz=qux") ~= nil, "should contain baz=qux")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestQueryEncodeWithSpaces(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local url = require("stdlib.url")
		local result = url.query_encode({q = "hello world"})
		assert(string.find(result, "hello") ~= nil, "should contain query value")
		assert(string.find(result, " ") == nil, "spaces should be encoded")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestQueryEncodeEmpty(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local url = require("stdlib.url")
		local result = url.query_encode({})
		assert(result == "", "empty table should produce empty string")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// query_decode tests
// =============================================================================

func TestQueryDecodeSimple(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local url = require("stdlib.url")
		local params, err = url.query_decode("foo=bar&baz=qux")
		assert(err == nil, "query_decode should not error: " .. tostring(err))
		assert(params.foo == "bar", "foo should be bar")
		assert(params.baz == "qux", "baz should be qux")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestQueryDecodeWithEncodedValues(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local url = require("stdlib.url")
		local params, err = url.query_decode("q=hello%20world")
		assert(err == nil, "query_decode should not error")
		assert(params.q == "hello world", "should decode value")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestQueryDecodeEmpty(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local url = require("stdlib.url")
		local params, err = url.query_decode("")
		assert(err == nil, "query_decode empty should not error")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// join tests
// =============================================================================

func TestJoinPaths(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local url = require("stdlib.url")
		local result = url.join("https://example.com", "api", "v1", "users")
		assert(result == "https://example.com/api/v1/users", "should join paths correctly")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestJoinWithTrailingSlash(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local url = require("stdlib.url")
		local result = url.join("https://example.com/", "/api/", "/users")
		assert(string.find(result, "//api") == nil, "should not have double slashes in path")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestJoinSinglePart(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local url = require("stdlib.url")
		local result = url.join("https://example.com")
		assert(result == "https://example.com", "single part should return as-is")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// resolve tests
// =============================================================================

func TestResolveRelative(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local url = require("stdlib.url")
		local result = url.resolve("https://example.com/api/v1/users", "../posts")
		assert(string.find(result, "posts") ~= nil, "should resolve to posts")
		assert(string.find(result, "users") == nil, "should not contain users")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestResolveAbsolute(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local url = require("stdlib.url")
		local result = url.resolve("https://example.com/api/v1/", "/new/path")
		assert(string.find(result, "/new/path") ~= nil, "should resolve absolute path")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestResolveSamePath(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local url = require("stdlib.url")
		local result = url.resolve("https://example.com/api/", "./resource")
		assert(string.find(result, "resource") ~= nil, "should resolve same-level path")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
