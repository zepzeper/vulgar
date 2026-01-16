package k8s

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
// client tests
// =============================================================================

func TestClientDefault(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local k8s = require("integrations.k8s")
		local client, err = k8s.client()
		-- May error if no kubeconfig
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestClientWithKubeconfig(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local k8s = require("integrations.k8s")
		local client, err = k8s.client({kubeconfig = "/nonexistent/path"})
		-- Should error for invalid path
		assert(client == nil or err ~= nil, "should error for invalid kubeconfig")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// list_pods tests
// =============================================================================

func TestListPodsNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local k8s = require("integrations.k8s")
		local pods, err = k8s.list_pods(nil, "default")
		assert(pods == nil, "pods should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// get_pod tests
// =============================================================================

func TestGetPodNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local k8s = require("integrations.k8s")
		local pod, err = k8s.get_pod(nil, "default", "pod-name")
		assert(pod == nil, "pod should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// list_deployments tests
// =============================================================================

func TestListDeploymentsNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local k8s = require("integrations.k8s")
		local deployments, err = k8s.list_deployments(nil, "default")
		assert(deployments == nil, "deployments should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// scale tests
// =============================================================================

func TestScaleNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local k8s = require("integrations.k8s")
		local err = k8s.scale(nil, "default", "deployment", "my-app", 3)
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// logs tests
// =============================================================================

func TestLogsNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local k8s = require("integrations.k8s")
		local logs, err = k8s.logs(nil, "default", "pod-name")
		assert(logs == nil, "logs should be nil")
		assert(err ~= nil, "should error without client")
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
		local k8s = require("integrations.k8s")
		local output, err = k8s.exec(nil, "default", "pod-name", {"echo", "hello"})
		assert(output == nil, "output should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// apply tests
// =============================================================================

func TestApplyNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local k8s = require("integrations.k8s")
		local err = k8s.apply(nil, "apiVersion: v1\nkind: ConfigMap")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// delete tests
// =============================================================================

func TestDeleteNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local k8s = require("integrations.k8s")
		local err = k8s.delete(nil, "default", "pod", "pod-name")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// list_namespaces tests
// =============================================================================

func TestListNamespacesNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local k8s = require("integrations.k8s")
		local namespaces, err = k8s.list_namespaces(nil)
		assert(namespaces == nil, "namespaces should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// list_services tests
// =============================================================================

func TestListServicesNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local k8s = require("integrations.k8s")
		local services, err = k8s.list_services(nil, "default")
		assert(services == nil, "services should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
