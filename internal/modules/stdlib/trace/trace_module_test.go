package trace

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
// start tests
// =============================================================================

func TestStart(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local trace = require("stdlib.trace")
		local span, err = trace.start("test_operation")
		assert(err == nil, "start should not error: " .. tostring(err))
		assert(span ~= nil, "span should not be nil")
		trace.finish(span)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestStartWithOptions(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local trace = require("stdlib.trace")
		local span, err = trace.start("test_operation", {
			tags = {service = "test", version = "1.0"}
		})
		assert(err == nil, "start with options should not error: " .. tostring(err))
		assert(span ~= nil, "span should not be nil")
		trace.finish(span)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// child tests
// =============================================================================

func TestChild(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local trace = require("stdlib.trace")
		local parent, _ = trace.start("parent")
		
		local child, err = trace.child(parent, "child_operation")
		assert(err == nil, "child should not error: " .. tostring(err))
		assert(child ~= nil, "child span should not be nil")
		
		trace.finish(child)
		trace.finish(parent)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestChildNoParent(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local trace = require("stdlib.trace")
		
		local child, err = trace.child(nil, "child_operation")
		-- Should either create root span or error
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// finish tests
// =============================================================================

func TestFinish(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local trace = require("stdlib.trace")
		local span, _ = trace.start("test")
		
		local err = trace.finish(span)
		assert(err == nil, "finish should not error: " .. tostring(err))
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestFinishNil(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local trace = require("stdlib.trace")
		local err = trace.finish(nil)
		-- Should handle nil gracefully
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// tag tests
// =============================================================================

func TestTag(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local trace = require("stdlib.trace")
		local span, _ = trace.start("test")
		
		trace.tag(span, "http.method", "GET")
		trace.tag(span, "http.status", 200)
		
		trace.finish(span)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// log tests
// =============================================================================

func TestLog(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local trace = require("stdlib.trace")
		local span, _ = trace.start("test")
		
		trace.log(span, "Processing started")
		trace.log(span, "Processing completed")
		
		trace.finish(span)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// error tests
// =============================================================================

func TestError(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local trace = require("stdlib.trace")
		local span, _ = trace.start("test")
		
		trace.error(span, "Something went wrong")
		
		trace.finish(span)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// context tests
// =============================================================================

func TestContext(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local trace = require("stdlib.trace")
		local span, _ = trace.start("test")
		
		local ctx = trace.context(span)
		assert(ctx ~= nil, "context should not be nil")
		
		trace.finish(span)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// inject / extract tests
// =============================================================================

func TestInjectExtract(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local trace = require("stdlib.trace")
		local span, _ = trace.start("test")
		
		local headers = {}
		trace.inject(span, headers)
		
		-- Headers should contain trace context
		
		local extracted = trace.extract(headers)
		-- Should extract context from headers
		
		trace.finish(span)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
