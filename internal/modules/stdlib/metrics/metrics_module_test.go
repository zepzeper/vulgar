package metrics

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
// counter tests
// =============================================================================

func TestCounter(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local metrics = require("stdlib.metrics")
		local counter, err = metrics.counter("requests_total", {
			help = "Total number of requests"
		})
		assert(err == nil, "counter should not error: " .. tostring(err))
		assert(counter ~= nil, "counter should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestCounterInc(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local metrics = require("stdlib.metrics")
		local counter, _ = metrics.counter("test_counter")
		
		metrics.inc(counter)
		metrics.inc(counter)
		
		local value = metrics.value(counter)
		assert(value == 2, "counter should be 2")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestCounterAdd(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local metrics = require("stdlib.metrics")
		local counter, _ = metrics.counter("test_counter")
		
		metrics.add(counter, 5)
		metrics.add(counter, 3)
		
		local value = metrics.value(counter)
		assert(value == 8, "counter should be 8")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// gauge tests
// =============================================================================

func TestGauge(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local metrics = require("stdlib.metrics")
		local gauge, err = metrics.gauge("temperature", {
			help = "Current temperature"
		})
		assert(err == nil, "gauge should not error: " .. tostring(err))
		assert(gauge ~= nil, "gauge should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestGaugeSet(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local metrics = require("stdlib.metrics")
		local gauge, _ = metrics.gauge("test_gauge")
		
		metrics.set(gauge, 42)
		
		local value = metrics.value(gauge)
		assert(value == 42, "gauge should be 42")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestGaugeIncDec(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local metrics = require("stdlib.metrics")
		local gauge, _ = metrics.gauge("test_gauge")
		
		metrics.set(gauge, 10)
		metrics.inc(gauge)
		metrics.dec(gauge)
		metrics.dec(gauge)
		
		local value = metrics.value(gauge)
		assert(value == 9, "gauge should be 9")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// histogram tests
// =============================================================================

func TestHistogram(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local metrics = require("stdlib.metrics")
		local histogram, err = metrics.histogram("request_duration", {
			help = "Request duration in seconds",
			buckets = {0.1, 0.5, 1, 2, 5}
		})
		assert(err == nil, "histogram should not error: " .. tostring(err))
		assert(histogram ~= nil, "histogram should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestHistogramObserve(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local metrics = require("stdlib.metrics")
		local histogram, _ = metrics.histogram("test_histogram")
		
		metrics.observe(histogram, 0.3)
		metrics.observe(histogram, 0.7)
		metrics.observe(histogram, 1.5)
		
		-- Should record observations
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// summary tests
// =============================================================================

func TestSummary(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local metrics = require("stdlib.metrics")
		local summary, err = metrics.summary("response_size", {
			help = "Response size in bytes",
			quantiles = {0.5, 0.9, 0.99}
		})
		assert(err == nil, "summary should not error: " .. tostring(err))
		assert(summary ~= nil, "summary should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// labels tests
// =============================================================================

func TestWithLabels(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local metrics = require("stdlib.metrics")
		local counter, _ = metrics.counter("requests_total", {
			labels = {"method", "status"}
		})
		
		local labeled = metrics.with_labels(counter, {method = "GET", status = "200"})
		metrics.inc(labeled)
		
		-- Should work with labels
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// export tests
// =============================================================================

func TestExport(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local metrics = require("stdlib.metrics")
		
		metrics.counter("test_counter")
		metrics.gauge("test_gauge")
		
		local output, err = metrics.export()
		assert(err == nil, "export should not error: " .. tostring(err))
		assert(output ~= nil, "output should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// reset tests
// =============================================================================

func TestReset(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local metrics = require("stdlib.metrics")
		local counter, _ = metrics.counter("test_counter")
		
		metrics.inc(counter)
		metrics.inc(counter)
		metrics.reset(counter)
		
		local value = metrics.value(counter)
		assert(value == 0, "counter should be 0 after reset")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
