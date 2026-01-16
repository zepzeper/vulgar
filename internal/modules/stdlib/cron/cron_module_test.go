package cron

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
// schedule tests
// =============================================================================

func TestScheduleValidCron(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local cron = require("stdlib.cron")
		local executed = false
		
		local job, err = cron.schedule("* * * * *", function()
			executed = true
		end)
		
		assert(err == nil, "schedule should not error: " .. tostring(err))
		assert(job ~= nil, "job should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestScheduleInvalidCron(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local cron = require("stdlib.cron")
		
		local job, err = cron.schedule("invalid cron expression", function() end)
		assert(job == nil, "job should be nil for invalid expression")
		assert(err ~= nil, "should return error for invalid cron expression")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestScheduleHourly(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local cron = require("stdlib.cron")
		
		local job, err = cron.schedule("0 * * * *", function()
			-- Run every hour at minute 0
		end)
		
		assert(err == nil, "schedule hourly should not error: " .. tostring(err))
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestScheduleDaily(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local cron = require("stdlib.cron")
		
		local job, err = cron.schedule("0 0 * * *", function()
			-- Run daily at midnight
		end)
		
		assert(err == nil, "schedule daily should not error: " .. tostring(err))
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// every tests
// =============================================================================

func TestEveryMinutes(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local cron = require("stdlib.cron")
		
		local job, err = cron.every("5m", function()
			-- Run every 5 minutes
		end)
		
		assert(err == nil, "every 5m should not error: " .. tostring(err))
		assert(job ~= nil, "job should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestEverySeconds(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local cron = require("stdlib.cron")
		
		local job, err = cron.every("30s", function()
			-- Run every 30 seconds
		end)
		
		assert(err == nil, "every 30s should not error: " .. tostring(err))
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestEveryHours(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local cron = require("stdlib.cron")
		
		local job, err = cron.every("2h", function()
			-- Run every 2 hours
		end)
		
		assert(err == nil, "every 2h should not error: " .. tostring(err))
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestEveryInvalidDuration(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local cron = require("stdlib.cron")
		
		local job, err = cron.every("invalid", function() end)
		assert(job == nil, "job should be nil for invalid duration")
		assert(err ~= nil, "should return error for invalid duration")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// at tests
// =============================================================================

func TestAtFutureTime(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local cron = require("stdlib.cron")
		
		local job, err = cron.at("2030-01-01T00:00:00Z", function()
			-- Run at specific time
		end)
		
		assert(err == nil, "at future time should not error: " .. tostring(err))
		assert(job ~= nil, "job should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestAtInvalidTime(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local cron = require("stdlib.cron")
		
		local job, err = cron.at("invalid time", function() end)
		assert(job == nil, "job should be nil for invalid time")
		assert(err ~= nil, "should return error for invalid time")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestAtPastTime(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local cron = require("stdlib.cron")
		
		local job, err = cron.at("2000-01-01T00:00:00Z", function() end)
		-- Past time might error or schedule immediately
		-- Behavior depends on implementation
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// cancel tests
// =============================================================================

func TestCancelJob(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local cron = require("stdlib.cron")
		
		local job, _ = cron.every("1h", function() end)
		local err = cron.cancel(job)
		
		assert(err == nil, "cancel should not error: " .. tostring(err))
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestCancelNilJob(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local cron = require("stdlib.cron")
		
		local err = cron.cancel(nil)
		-- Should handle nil gracefully
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// start / stop tests
// =============================================================================

func TestStartScheduler(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local cron = require("stdlib.cron")
		
		local err = cron.start()
		assert(err == nil, "start should not error: " .. tostring(err))
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestStopScheduler(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local cron = require("stdlib.cron")
		
		cron.start()
		local err = cron.stop()
		assert(err == nil, "stop should not error: " .. tostring(err))
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestStartTwice(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local cron = require("stdlib.cron")
		
		cron.start()
		local err = cron.start()
		-- Should handle being started twice gracefully
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// list tests
// =============================================================================

func TestListJobs(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local cron = require("stdlib.cron")
		
		-- Schedule some jobs
		cron.every("1m", function() end)
		cron.every("5m", function() end)
		
		local jobs, err = cron.list()
		assert(err == nil, "list should not error: " .. tostring(err))
		assert(jobs ~= nil, "jobs should not be nil")
		assert(#jobs >= 2, "should have at least 2 jobs")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestListEmpty(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local cron = require("stdlib.cron")
		
		-- Fresh state, no jobs scheduled
		local jobs, err = cron.list()
		assert(err == nil, "list should not error: " .. tostring(err))
		-- jobs might be empty or nil
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// integration tests
// =============================================================================

func TestScheduleStartStop(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local cron = require("stdlib.cron")
		
		-- Schedule a job
		local job, err = cron.every("1m", function()
			-- Do something
		end)
		assert(err == nil, "schedule should not error")
		
		-- Start the scheduler
		err = cron.start()
		assert(err == nil, "start should not error")
		
		-- List jobs
		local jobs, err = cron.list()
		assert(err == nil, "list should not error")
		
		-- Stop the scheduler
		err = cron.stop()
		assert(err == nil, "stop should not error")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
