package ssh

import (
	"os"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func newTestState() *lua.LState {
	L := lua.NewState()
	L.PreloadModule(ModuleName, Loader)
	return L
}

func skipIfNoSSH(t *testing.T) {
	if os.Getenv("SSH_TEST_HOST") == "" {
		t.Skip("SSH_TEST_HOST not set, skipping integration test")
	}
}

// =============================================================================
// connect tests
// =============================================================================

func TestConnectMissingConfig(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local ssh = require("integrations.ssh")
		local client, err = ssh.connect({})
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
		local ssh = require("integrations.ssh")
		local client, err = ssh.connect({
			host = "invalid-host-xyz",
			port = 22,
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
// exec tests
// =============================================================================

func TestExecNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local ssh = require("integrations.ssh")
		local output, err = ssh.exec(nil, "echo hello")
		assert(output == nil, "output should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestExec(t *testing.T) {
	skipIfNoSSH(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("host", lua.LString(os.Getenv("SSH_TEST_HOST")))
	L.SetGlobal("user", lua.LString(os.Getenv("SSH_TEST_USER")))
	L.SetGlobal("pass", lua.LString(os.Getenv("SSH_TEST_PASS")))

	err := L.DoString(`
		local ssh = require("integrations.ssh")
		local client, _ = ssh.connect({
			host = host,
			user = user,
			password = pass
		})
		
		local output, err = ssh.exec(client, "echo hello")
		assert(err == nil, "exec should not error: " .. tostring(err))
		assert(string.find(output, "hello") ~= nil, "should contain hello")
		
		ssh.close(client)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// shell tests
// =============================================================================

func TestShellNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local ssh = require("integrations.ssh")
		local shell, err = ssh.shell(nil)
		assert(shell == nil, "shell should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// upload / download tests
// =============================================================================

func TestUploadNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local ssh = require("integrations.ssh")
		local err = ssh.upload(nil, "/local/path", "/remote/path")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestDownloadNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local ssh = require("integrations.ssh")
		local err = ssh.download(nil, "/remote/path", "/local/path")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// tunnel tests
// =============================================================================

func TestTunnelNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local ssh = require("integrations.ssh")
		local tunnel, err = ssh.tunnel(nil, 8080, "localhost", 80)
		assert(tunnel == nil, "tunnel should be nil")
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
		local ssh = require("integrations.ssh")
		local err = ssh.close(nil)
		-- Should handle nil gracefully
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
