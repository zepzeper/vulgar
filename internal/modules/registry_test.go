package modules

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func TestRegister(t *testing.T) {
	originalRegistry := registry
	registry = make(map[string]lua.LGFunction)
	defer func() { registry = originalRegistry }()

	loader := func(L *lua.LState) int { return 0 }
	Register("test_module", loader)

	if _, ok := registry["test_module"]; !ok {
		t.Error("expected module to be registered")
	}
}

func TestRegisterPreload(t *testing.T) {
	originalPreload := preloadRegistry
	preloadRegistry = make(map[string]func(L *lua.LState))
	defer func() { preloadRegistry = originalPreload }()

	opener := func(L *lua.LState) {}
	RegisterPreload("test_preload", opener)

	if _, ok := preloadRegistry["test_preload"]; !ok {
		t.Error("expected preload module to be registered")
	}
}

func TestGetRegistry(t *testing.T) {
	originalRegistry := registry
	registry = make(map[string]lua.LGFunction)
	defer func() { registry = originalRegistry }()

	loader := func(L *lua.LState) int { return 0 }
	Register("module_a", loader)
	Register("module_b", loader)

	result := GetRegistry()

	if len(result) != 2 {
		t.Errorf("expected 2 modules, got %d", len(result))
	}

	if _, ok := result["module_a"]; !ok {
		t.Error("expected module_a in registry")
	}

	if _, ok := result["module_b"]; !ok {
		t.Error("expected module_b in registry")
	}
}

func TestGetRegistryReturnsCopy(t *testing.T) {
	originalRegistry := registry
	registry = make(map[string]lua.LGFunction)
	defer func() { registry = originalRegistry }()

	loader := func(L *lua.LState) int { return 0 }
	Register("original", loader)

	result := GetRegistry()
	result["modified"] = loader

	// Original should not be modified
	if _, ok := registry["modified"]; ok {
		t.Error("GetRegistry should return a copy, not the original")
	}
}

func TestGetPreloadRegistry(t *testing.T) {
	originalPreload := preloadRegistry
	preloadRegistry = make(map[string]func(L *lua.LState))
	defer func() { preloadRegistry = originalPreload }()

	opener := func(L *lua.LState) {}
	RegisterPreload("preload_a", opener)
	RegisterPreload("preload_b", opener)

	result := GetPreloadRegistry()

	if len(result) != 2 {
		t.Errorf("expected 2 preload modules, got %d", len(result))
	}
}

func TestGetPreloadRegistryReturnsCopy(t *testing.T) {
	originalPreload := preloadRegistry
	preloadRegistry = make(map[string]func(L *lua.LState))
	defer func() { preloadRegistry = originalPreload }()

	opener := func(L *lua.LState) {}
	RegisterPreload("original", opener)

	result := GetPreloadRegistry()
	result["modified"] = opener

	// Original should not be modified
	if _, ok := preloadRegistry["modified"]; ok {
		t.Error("GetPreloadRegistry should return a copy, not the original")
	}
}
