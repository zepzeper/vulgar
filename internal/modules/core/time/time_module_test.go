package time

import (
	"testing"
	gotime "time"

	lua "github.com/yuin/gopher-lua"
)

func setupLuaState() *lua.LState {
	L := lua.NewState()
	L.PreloadModule(ModuleName, Loader)
	return L
}

func TestNowReturnsTimestamp(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	before := gotime.Now().Unix()

	err := L.DoString(`
		local time = require("time")
		result = time.now()
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	after := gotime.Now().Unix()
	result := int64(L.GetGlobal("result").(lua.LNumber))

	if result < before || result > after {
		t.Errorf("timestamp %d not in expected range [%d, %d]", result, before, after)
	}
}

func TestNowMsReturnsMilliseconds(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	before := gotime.Now().UnixMilli()

	err := L.DoString(`
		local time = require("time")
		result = time.now_ms()
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	after := gotime.Now().UnixMilli()
	result := int64(L.GetGlobal("result").(lua.LNumber))

	if result < before || result > after {
		t.Errorf("timestamp %d not in expected range [%d, %d]", result, before, after)
	}
}

func TestFormatWithTimestamp(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	// Use local time to avoid timezone issues
	localTime := gotime.Date(2024, 1, 15, 12, 30, 45, 0, gotime.Local)
	ts := localTime.Unix()

	L.SetGlobal("test_ts", lua.LNumber(ts))
	err := L.DoString(`
		local time = require("time")
		result = time.format(test_ts, "2006-01-02 15:04:05")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").String()
	expected := localTime.Format("2006-01-02 15:04:05")
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestFormatWithDefaultLayout(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	localTime := gotime.Date(2024, 1, 15, 12, 30, 45, 0, gotime.Local)
	ts := localTime.Unix()

	L.SetGlobal("test_ts", lua.LNumber(ts))
	err := L.DoString(`
		local time = require("time")
		result = time.format(test_ts)
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").String()
	expected := localTime.Format("2006-01-02 15:04:05")
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestFormatWithCurrentTime(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local time = require("time")
		result = time.format()
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").String()
	if len(result) != 19 { // "2006-01-02 15:04:05" length
		t.Errorf("unexpected format length: %s", result)
	}
}

func TestParseValidTime(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local time = require("time")
		result, err = time.parse("2024-01-15", "2006-01-02")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := int64(L.GetGlobal("result").(lua.LNumber))
	errVal := L.GetGlobal("err")

	expected := gotime.Date(2024, 1, 15, 0, 0, 0, 0, gotime.UTC).Unix()
	if result != expected {
		t.Errorf("expected %d, got %d", expected, result)
	}
	if errVal != lua.LNil {
		t.Error("expected no error")
	}
}

func TestParseInvalidTime(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local time = require("time")
		result, err = time.parse("not-a-date", "2006-01-02")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result")
	errVal := L.GetGlobal("err")

	if result != lua.LNil {
		t.Error("expected nil result for invalid time")
	}
	if errVal == lua.LNil {
		t.Error("expected error for invalid time")
	}
}

func TestSleep(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	start := gotime.Now()

	err := L.DoString(`
		local time = require("time")
		time.sleep(0.1)
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	elapsed := gotime.Since(start)
	if elapsed < 100*gotime.Millisecond {
		t.Errorf("sleep too short: %v", elapsed)
	}
}

func TestDateWithTimestamp(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	localTime := gotime.Date(2024, 6, 15, 14, 30, 45, 0, gotime.Local)
	ts := localTime.Unix()

	L.SetGlobal("test_ts", lua.LNumber(ts))
	err := L.DoString(`
		local time = require("time")
		result = time.date(test_ts)
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").(*lua.LTable)

	year := int(L.GetField(result, "year").(lua.LNumber))
	month := int(L.GetField(result, "month").(lua.LNumber))
	day := int(L.GetField(result, "day").(lua.LNumber))
	hour := int(L.GetField(result, "hour").(lua.LNumber))
	minute := int(L.GetField(result, "minute").(lua.LNumber))
	second := int(L.GetField(result, "second").(lua.LNumber))

	if year != 2024 {
		t.Errorf("expected year 2024, got %d", year)
	}
	if month != 6 {
		t.Errorf("expected month 6, got %d", month)
	}
	if day != 15 {
		t.Errorf("expected day 15, got %d", day)
	}
	if hour != 14 {
		t.Errorf("expected hour 14, got %d", hour)
	}
	if minute != 30 {
		t.Errorf("expected minute 30, got %d", minute)
	}
	if second != 45 {
		t.Errorf("expected second 45, got %d", second)
	}
}

func TestDateWithCurrentTime(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local time = require("time")
		result = time.date()
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").(*lua.LTable)
	year := int(L.GetField(result, "year").(lua.LNumber))

	if year < 2024 {
		t.Errorf("unexpected year: %d", year)
	}
}

func TestAddSeconds(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	ts := gotime.Date(2024, 1, 15, 12, 0, 0, 0, gotime.UTC).Unix()

	L.SetGlobal("test_ts", lua.LNumber(ts))
	err := L.DoString(`
		local time = require("time")
		result = time.add(test_ts, 3600)
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := int64(L.GetGlobal("result").(lua.LNumber))
	expected := ts + 3600

	if result != expected {
		t.Errorf("expected %d, got %d", expected, result)
	}
}

func TestAddNegativeSeconds(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	ts := gotime.Date(2024, 1, 15, 12, 0, 0, 0, gotime.UTC).Unix()

	L.SetGlobal("test_ts", lua.LNumber(ts))
	err := L.DoString(`
		local time = require("time")
		result = time.add(test_ts, -3600)
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := int64(L.GetGlobal("result").(lua.LNumber))
	expected := ts - 3600

	if result != expected {
		t.Errorf("expected %d, got %d", expected, result)
	}
}

func TestSubTimestamps(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	ts1 := gotime.Date(2024, 1, 15, 13, 0, 0, 0, gotime.UTC).Unix()
	ts2 := gotime.Date(2024, 1, 15, 12, 0, 0, 0, gotime.UTC).Unix()

	L.SetGlobal("ts1", lua.LNumber(ts1))
	L.SetGlobal("ts2", lua.LNumber(ts2))
	err := L.DoString(`
		local time = require("time")
		result = time.sub(ts1, ts2)
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := float64(L.GetGlobal("result").(lua.LNumber))
	if result != 3600 {
		t.Errorf("expected 3600, got %f", result)
	}
}

func TestUtc(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	ts := gotime.Now().Unix()

	L.SetGlobal("test_ts", lua.LNumber(ts))
	err := L.DoString(`
		local time = require("time")
		result = time.utc(test_ts)
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := int64(L.GetGlobal("result").(lua.LNumber))
	// UTC conversion of Unix timestamp should be the same
	if result != ts {
		t.Errorf("expected %d, got %d", ts, result)
	}
}

func TestLoaderReturnsModule(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule(ModuleName, Loader)
	err := L.DoString(`time = require("time")`)
	if err != nil {
		t.Fatalf("failed to require module: %v", err)
	}

	mod := L.GetGlobal("time")
	if mod.Type() != lua.LTTable {
		t.Errorf("expected table, got %s", mod.Type())
	}
}
