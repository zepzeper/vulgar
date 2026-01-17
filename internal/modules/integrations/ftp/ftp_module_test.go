package ftp

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
// connect tests
// =============================================================================

func TestConnectMissingConfig(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local ftp = require("integrations.ftp")
		local client, err = ftp.connect({})
		assert(client == nil or err ~= nil, "should error with missing config")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestConnectInvalidHost(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local ftp = require("integrations.ftp")
		local client, err = ftp.connect({
			host = "invalid-host-xyz",
			port = 21,
			user = "test",
			password = "test"
		})
		assert(client == nil, "client should be nil for invalid host")
		assert(err ~= nil, "should error for invalid host")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// upload tests
// =============================================================================

func TestUploadNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local ftp = require("integrations.ftp")
		local err = ftp.upload(nil, "/local/path", "/remote/path")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// download tests
// =============================================================================

func TestDownloadNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local ftp = require("integrations.ftp")
		local err = ftp.download(nil, "/remote/path", "/local/path")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// list tests
// =============================================================================

func TestListNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local ftp = require("integrations.ftp")
		local files, err = ftp.list(nil, "/path")
		assert(files == nil, "files should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// mkdir tests
// =============================================================================

func TestMkdirNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local ftp = require("integrations.ftp")
		local err = ftp.mkdir(nil, "/path")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// remove tests
// =============================================================================

func TestRemoveNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local ftp = require("integrations.ftp")
		local err = ftp.remove(nil, "/path")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// rename tests
// =============================================================================

func TestRenameNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local ftp = require("integrations.ftp")
		local err = ftp.rename(nil, "/old", "/new")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// cd / pwd tests
// =============================================================================

func TestCdNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local ftp = require("integrations.ftp")
		local err = ftp.cd(nil, "/path")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestPwdNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local ftp = require("integrations.ftp")
		local path, err = ftp.pwd(nil)
		assert(path == nil, "path should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// close tests
// =============================================================================

func TestCloseNil(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local ftp = require("integrations.ftp")
		local err = ftp.close(nil)
		-- Should handle nil gracefully
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
