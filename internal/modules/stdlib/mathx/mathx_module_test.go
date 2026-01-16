package mathx

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
// round tests
// =============================================================================

func TestRoundToTwoDecimals(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local mathx = require("stdlib.mathx")
		local result = mathx.round(3.14159, 2)
		assert(result == 3.14, "should round to 2 decimals, got: " .. tostring(result))
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestRoundToInteger(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local mathx = require("stdlib.mathx")
		local result = mathx.round(3.7, 0)
		assert(result == 4, "should round to nearest integer")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestRoundNegative(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local mathx = require("stdlib.mathx")
		local result = mathx.round(-3.5, 0)
		assert(result == -4 or result == -3, "should round negative number")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// clamp tests
// =============================================================================

func TestClampWithinRange(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local mathx = require("stdlib.mathx")
		local result = mathx.clamp(5, 0, 10)
		assert(result == 5, "value within range should be unchanged")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestClampBelowMin(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local mathx = require("stdlib.mathx")
		local result = mathx.clamp(-5, 0, 10)
		assert(result == 0, "value below min should be clamped to min")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestClampAboveMax(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local mathx = require("stdlib.mathx")
		local result = mathx.clamp(15, 0, 10)
		assert(result == 10, "value above max should be clamped to max")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// lerp tests
// =============================================================================

func TestLerpStart(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local mathx = require("stdlib.mathx")
		local result = mathx.lerp(0, 100, 0)
		assert(result == 0, "lerp at t=0 should return start value")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestLerpEnd(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local mathx = require("stdlib.mathx")
		local result = mathx.lerp(0, 100, 1)
		assert(result == 100, "lerp at t=1 should return end value")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestLerpMiddle(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local mathx = require("stdlib.mathx")
		local result = mathx.lerp(0, 100, 0.5)
		assert(result == 50, "lerp at t=0.5 should return middle value")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// map tests
// =============================================================================

func TestMapValue(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local mathx = require("stdlib.mathx")
		-- Map 5 from range [0,10] to range [0,100]
		local result = mathx.map(5, 0, 10, 0, 100)
		assert(result == 50, "should map 5 to 50")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestMapValueDifferentRanges(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local mathx = require("stdlib.mathx")
		-- Map 0 from range [-1,1] to range [0,255]
		local result = mathx.map(0, -1, 1, 0, 255)
		assert(result == 127.5 or result == 127 or result == 128, "should map center correctly")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// sign tests
// =============================================================================

func TestSignPositive(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local mathx = require("stdlib.mathx")
		assert(mathx.sign(42) == 1, "positive number should return 1")
		assert(mathx.sign(0.5) == 1, "positive decimal should return 1")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestSignNegative(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local mathx = require("stdlib.mathx")
		assert(mathx.sign(-42) == -1, "negative number should return -1")
		assert(mathx.sign(-0.5) == -1, "negative decimal should return -1")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestSignZero(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local mathx = require("stdlib.mathx")
		assert(mathx.sign(0) == 0, "zero should return 0")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// sum tests
// =============================================================================

func TestSumArgs(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local mathx = require("stdlib.mathx")
		local result = mathx.sum(1, 2, 3, 4, 5)
		assert(result == 15, "sum of 1-5 should be 15")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestSumTable(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local mathx = require("stdlib.mathx")
		local result = mathx.sum({1, 2, 3, 4, 5})
		assert(result == 15, "sum of table should be 15")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestSumEmpty(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local mathx = require("stdlib.mathx")
		local result = mathx.sum({})
		assert(result == 0, "sum of empty table should be 0")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// avg tests
// =============================================================================

func TestAvgArgs(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local mathx = require("stdlib.mathx")
		local result = mathx.avg(10, 20, 30)
		assert(result == 20, "average of 10,20,30 should be 20")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestAvgTable(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local mathx = require("stdlib.mathx")
		local result = mathx.avg({10, 20, 30})
		assert(result == 20, "average of table should be 20")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestAvgSingleValue(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local mathx = require("stdlib.mathx")
		local result = mathx.avg(42)
		assert(result == 42, "average of single value should be itself")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// min / max tests
// =============================================================================

func TestMinArgs(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local mathx = require("stdlib.mathx")
		local result = mathx.min(5, 3, 8, 1, 9)
		assert(result == 1, "min should be 1")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestMinTable(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local mathx = require("stdlib.mathx")
		local result = mathx.min({5, 3, 8, 1, 9})
		assert(result == 1, "min of table should be 1")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestMaxArgs(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local mathx = require("stdlib.mathx")
		local result = mathx.max(5, 3, 8, 1, 9)
		assert(result == 9, "max should be 9")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestMaxTable(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local mathx = require("stdlib.mathx")
		local result = mathx.max({5, 3, 8, 1, 9})
		assert(result == 9, "max of table should be 9")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestMinMaxNegative(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local mathx = require("stdlib.mathx")
		assert(mathx.min(-5, -3, -8) == -8, "min of negatives should be -8")
		assert(mathx.max(-5, -3, -8) == -3, "max of negatives should be -3")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
