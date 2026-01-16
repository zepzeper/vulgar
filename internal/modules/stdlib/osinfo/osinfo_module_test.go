package osinfo

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
// hostname tests
// =============================================================================

func TestHostname(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local osinfo = require("stdlib.osinfo")
		local hostname, err = osinfo.hostname()
		assert(err == nil, "hostname should not error: " .. tostring(err))
		assert(hostname ~= nil, "hostname should not be nil")
		assert(#hostname > 0, "hostname should not be empty")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// platform tests
// =============================================================================

func TestPlatform(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local osinfo = require("stdlib.osinfo")
		local platform = osinfo.platform()
		assert(platform ~= nil, "platform should not be nil")
		assert(#platform > 0, "platform should not be empty")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// arch tests
// =============================================================================

func TestArch(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local osinfo = require("stdlib.osinfo")
		local arch = osinfo.arch()
		assert(arch ~= nil, "arch should not be nil")
		assert(#arch > 0, "arch should not be empty")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// cpus tests
// =============================================================================

func TestCpus(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local osinfo = require("stdlib.osinfo")
		local cpus = osinfo.cpus()
		assert(cpus ~= nil, "cpus should not be nil")
		assert(cpus > 0, "should have at least 1 CPU")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// memory tests
// =============================================================================

func TestMemory(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local osinfo = require("stdlib.osinfo")
		local mem, err = osinfo.memory()
		assert(err == nil, "memory should not error: " .. tostring(err))
		assert(mem ~= nil, "memory should not be nil")
		assert(mem.total ~= nil, "should have total memory")
		assert(mem.total > 0, "total memory should be positive")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// disk tests
// =============================================================================

func TestDisk(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local osinfo = require("stdlib.osinfo")
		local disk, err = osinfo.disk("/")
		assert(err == nil, "disk should not error: " .. tostring(err))
		assert(disk ~= nil, "disk should not be nil")
		assert(disk.total ~= nil, "should have total disk")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestDiskInvalidPath(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local osinfo = require("stdlib.osinfo")
		local disk, err = osinfo.disk("/nonexistent/path/xyz")
		-- May error or return zeroes depending on implementation
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// uptime tests
// =============================================================================

func TestUptime(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local osinfo = require("stdlib.osinfo")
		local uptime, err = osinfo.uptime()
		assert(err == nil, "uptime should not error: " .. tostring(err))
		assert(uptime ~= nil, "uptime should not be nil")
		assert(uptime >= 0, "uptime should be non-negative")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// load_avg tests
// =============================================================================

func TestLoadAvg(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local osinfo = require("stdlib.osinfo")
		local load, err = osinfo.load_avg()
		-- May not be available on all platforms
		if err == nil then
			assert(load ~= nil, "load should not be nil")
		end
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// processes tests
// =============================================================================

func TestProcesses(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local osinfo = require("stdlib.osinfo")
		local procs, err = osinfo.processes()
		assert(err == nil, "processes should not error: " .. tostring(err))
		assert(procs ~= nil, "processes should not be nil")
		assert(#procs > 0, "should have at least 1 process")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// network_interfaces tests
// =============================================================================

func TestNetworkInterfaces(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local osinfo = require("stdlib.osinfo")
		local interfaces, err = osinfo.network_interfaces()
		assert(err == nil, "network_interfaces should not error: " .. tostring(err))
		assert(interfaces ~= nil, "interfaces should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// user tests
// =============================================================================

func TestUser(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local osinfo = require("stdlib.osinfo")
		local user, err = osinfo.user()
		assert(err == nil, "user should not error: " .. tostring(err))
		assert(user ~= nil, "user should not be nil")
		assert(user.username ~= nil, "should have username")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// env tests
// =============================================================================

func TestEnv(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local osinfo = require("stdlib.osinfo")
		local env = osinfo.env()
		assert(env ~= nil, "env should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
