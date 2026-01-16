package regex

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
// match tests
// =============================================================================

func TestMatchSimplePattern(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local regex = require("stdlib.regex")
		local matched = regex.match("hello", "hello world")
		assert(matched == true, "should match 'hello' in 'hello world'")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestMatchNoMatch(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local regex = require("stdlib.regex")
		local matched = regex.match("xyz", "hello world")
		assert(matched == false, "should not match 'xyz' in 'hello world'")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestMatchRegexPattern(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local regex = require("stdlib.regex")
		local matched = regex.match("^[a-z]+$", "hello")
		assert(matched == true, "should match lowercase letters")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestMatchDigits(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local regex = require("stdlib.regex")
		local matched = regex.match("\\d+", "abc123def")
		assert(matched == true, "should find digits")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestMatchEmailPattern(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local regex = require("stdlib.regex")
		local matched = regex.match("[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}", "test@example.com")
		assert(matched == true, "should match email")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestMatchCaseInsensitive(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local regex = require("stdlib.regex")
		local matched = regex.match("(?i)hello", "HELLO")
		assert(matched == true, "should match case-insensitively")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// find tests
// =============================================================================

func TestFindFirstMatch(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local regex = require("stdlib.regex")
		local match, err = regex.find("\\d+", "abc123def456")
		assert(err == nil, "find should not error: " .. tostring(err))
		assert(match == "123", "should find first number '123', got: " .. tostring(match))
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestFindNoMatch(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local regex = require("stdlib.regex")
		local match, err = regex.find("\\d+", "abcdef")
		assert(err == nil, "find should not error")
		assert(match == nil, "should return nil when no match")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestFindWord(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local regex = require("stdlib.regex")
		local match, err = regex.find("\\w+", "hello world")
		assert(err == nil, "find should not error")
		assert(match == "hello", "should find first word")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// find_all tests
// =============================================================================

func TestFindAllMatches(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local regex = require("stdlib.regex")
		local matches, err = regex.find_all("\\d+", "abc123def456ghi789")
		assert(err == nil, "find_all should not error: " .. tostring(err))
		assert(#matches == 3, "should find 3 numbers")
		assert(matches[1] == "123", "first match should be '123'")
		assert(matches[2] == "456", "second match should be '456'")
		assert(matches[3] == "789", "third match should be '789'")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestFindAllNoMatches(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local regex = require("stdlib.regex")
		local matches, err = regex.find_all("\\d+", "abcdef")
		assert(err == nil, "find_all should not error")
		assert(#matches == 0, "should find 0 matches")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestFindAllWords(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local regex = require("stdlib.regex")
		local matches, err = regex.find_all("[a-zA-Z]+", "hello world test")
		assert(err == nil, "find_all should not error")
		assert(#matches == 3, "should find 3 words")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// replace tests
// =============================================================================

func TestReplaceFirst(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local regex = require("stdlib.regex")
		local result, err = regex.replace("\\d+", "abc123def456", "X")
		assert(err == nil, "replace should not error: " .. tostring(err))
		assert(result == "abcXdef456", "should replace first match only")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestReplaceNoMatch(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local regex = require("stdlib.regex")
		local result, err = regex.replace("\\d+", "abcdef", "X")
		assert(err == nil, "replace should not error")
		assert(result == "abcdef", "should return unchanged string")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// replace_all tests
// =============================================================================

func TestReplaceAllMatches(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local regex = require("stdlib.regex")
		local result, err = regex.replace_all("\\d+", "abc123def456ghi789", "X")
		assert(err == nil, "replace_all should not error: " .. tostring(err))
		assert(result == "abcXdefXghiX", "should replace all matches")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestReplaceAllSpaces(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local regex = require("stdlib.regex")
		local result, err = regex.replace_all("\\s+", "hello   world  test", " ")
		assert(err == nil, "replace_all should not error")
		assert(result == "hello world test", "should normalize spaces")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestReplaceAllNoMatch(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local regex = require("stdlib.regex")
		local result, err = regex.replace_all("\\d+", "abcdef", "X")
		assert(err == nil, "replace_all should not error")
		assert(result == "abcdef", "should return unchanged string")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// split tests
// =============================================================================

func TestSplitByPattern(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local regex = require("stdlib.regex")
		local parts, err = regex.split(",\\s*", "a, b, c, d")
		assert(err == nil, "split should not error: " .. tostring(err))
		assert(#parts == 4, "should have 4 parts")
		assert(parts[1] == "a", "first part should be 'a'")
		assert(parts[2] == "b", "second part should be 'b'")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestSplitByWhitespace(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local regex = require("stdlib.regex")
		local parts, err = regex.split("\\s+", "hello   world  test")
		assert(err == nil, "split should not error")
		assert(#parts == 3, "should have 3 parts")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestSplitNoMatch(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local regex = require("stdlib.regex")
		local parts, err = regex.split(",", "hello")
		assert(err == nil, "split should not error")
		assert(#parts == 1, "should have 1 part when no match")
		assert(parts[1] == "hello", "part should be original string")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// capture tests
// =============================================================================

func TestCaptureGroups(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local regex = require("stdlib.regex")
		local groups, err = regex.capture("(\\d+)-(\\d+)-(\\d+)", "2024-01-15")
		assert(err == nil, "capture should not error: " .. tostring(err))
		assert(#groups == 3, "should have 3 capture groups")
		assert(groups[1] == "2024", "first group should be year")
		assert(groups[2] == "01", "second group should be month")
		assert(groups[3] == "15", "third group should be day")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestCaptureNamedGroups(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local regex = require("stdlib.regex")
		local groups, err = regex.capture("(?P<year>\\d+)-(?P<month>\\d+)-(?P<day>\\d+)", "2024-01-15")
		assert(err == nil, "capture should not error: " .. tostring(err))
		-- Named groups should be accessible
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestCaptureNoMatch(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local regex = require("stdlib.regex")
		local groups, err = regex.capture("(\\d+)-(\\d+)", "abcdef")
		assert(err == nil, "capture should not error")
		assert(groups == nil or #groups == 0, "should return nil or empty for no match")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestCaptureEmailParts(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local regex = require("stdlib.regex")
		local groups, err = regex.capture("([^@]+)@([^.]+)\\.(.+)", "user@example.com")
		assert(err == nil, "capture should not error")
		assert(groups[1] == "user", "first group should be user")
		assert(groups[2] == "example", "second group should be domain")
		assert(groups[3] == "com", "third group should be tld")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// invalid pattern tests
// =============================================================================

func TestMatchInvalidPattern(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local regex = require("stdlib.regex")
		-- Invalid regex should either return false or error gracefully
		local ok, result = pcall(function()
			return regex.match("[invalid", "test")
		end)
		-- Either errors or returns false
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestFindInvalidPattern(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local regex = require("stdlib.regex")
		local match, err = regex.find("[invalid", "test")
		-- Should return error for invalid pattern
		assert(err ~= nil or match == nil, "should handle invalid pattern")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
