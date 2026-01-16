package docker

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

func skipIfNoDocker(t *testing.T) {
	if os.Getenv("DOCKER_HOST") == "" && os.Getenv("TEST_WITH_DOCKER") == "" {
		t.Skip("Docker not available, skipping integration test")
	}
}

// =============================================================================
// list_containers tests
// =============================================================================

func TestListContainers(t *testing.T) {
	skipIfNoDocker(t)
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local docker = require("integrations.docker")
		local containers, err = docker.list_containers()
		assert(err == nil, "list_containers should not error: " .. tostring(err))
		assert(containers ~= nil, "containers should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestListContainersWithOptions(t *testing.T) {
	skipIfNoDocker(t)
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local docker = require("integrations.docker")
		local containers, err = docker.list_containers({all = true})
		assert(err == nil, "list_containers should not error: " .. tostring(err))
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// inspect_container tests
// =============================================================================

func TestInspectContainerInvalid(t *testing.T) {
	skipIfNoDocker(t)
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local docker = require("integrations.docker")
		local info, err = docker.inspect_container("nonexistent_container_xyz")
		assert(info == nil or err ~= nil, "should error for nonexistent container")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// logs tests
// =============================================================================

func TestLogsInvalidContainer(t *testing.T) {
	skipIfNoDocker(t)
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local docker = require("integrations.docker")
		local logs, err = docker.logs("nonexistent_container_xyz")
		assert(logs == nil or err ~= nil, "should error for nonexistent container")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// list_images tests
// =============================================================================

func TestListImages(t *testing.T) {
	skipIfNoDocker(t)
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local docker = require("integrations.docker")
		local images, err = docker.list_images()
		assert(err == nil, "list_images should not error: " .. tostring(err))
		assert(images ~= nil, "images should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// pull_image tests
// =============================================================================

func TestPullImageInvalid(t *testing.T) {
	skipIfNoDocker(t)
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local docker = require("integrations.docker")
		local err = docker.pull_image("invalid/image/name/xyz123:nonexistent")
		assert(err ~= nil, "should error for invalid image")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// run tests
// =============================================================================

func TestRunMissingImage(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local docker = require("integrations.docker")
		local container, err = docker.run({})
		assert(container == nil or err ~= nil, "should error with missing image")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// stop / start / restart tests
// =============================================================================

func TestStopInvalidContainer(t *testing.T) {
	skipIfNoDocker(t)
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local docker = require("integrations.docker")
		local err = docker.stop("nonexistent_xyz")
		assert(err ~= nil, "should error for nonexistent container")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestStartInvalidContainer(t *testing.T) {
	skipIfNoDocker(t)
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local docker = require("integrations.docker")
		local err = docker.start("nonexistent_xyz")
		assert(err ~= nil, "should error for nonexistent container")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// remove tests
// =============================================================================

func TestRemoveInvalidContainer(t *testing.T) {
	skipIfNoDocker(t)
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local docker = require("integrations.docker")
		local err = docker.remove("nonexistent_xyz")
		assert(err ~= nil, "should error for nonexistent container")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// exec tests
// =============================================================================

func TestExecInvalidContainer(t *testing.T) {
	skipIfNoDocker(t)
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local docker = require("integrations.docker")
		local output, err = docker.exec("nonexistent_xyz", {"echo", "hello"})
		assert(output == nil or err ~= nil, "should error for nonexistent container")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// stats tests
// =============================================================================

func TestStatsInvalidContainer(t *testing.T) {
	skipIfNoDocker(t)
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local docker = require("integrations.docker")
		local stats, err = docker.stats("nonexistent_xyz")
		assert(stats == nil or err ~= nil, "should error for nonexistent container")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
