package validator

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
// is_email tests
// =============================================================================

func TestIsEmailValid(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local validator = require("stdlib.validator")
		assert(validator.is_email("user@example.com") == true, "should validate simple email")
		assert(validator.is_email("user.name@example.com") == true, "should validate email with dot")
		assert(validator.is_email("user+tag@example.com") == true, "should validate email with plus")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestIsEmailInvalid(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local validator = require("stdlib.validator")
		assert(validator.is_email("invalid") == false, "should reject string without @")
		assert(validator.is_email("@example.com") == false, "should reject email without local part")
		assert(validator.is_email("user@") == false, "should reject email without domain")
		assert(validator.is_email("") == false, "should reject empty string")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// is_url tests
// =============================================================================

func TestIsURLValid(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local validator = require("stdlib.validator")
		assert(validator.is_url("https://example.com") == true, "should validate HTTPS URL")
		assert(validator.is_url("http://example.com") == true, "should validate HTTP URL")
		assert(validator.is_url("https://example.com/path?query=1") == true, "should validate URL with path and query")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestIsURLInvalid(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local validator = require("stdlib.validator")
		assert(validator.is_url("not a url") == false, "should reject non-URL string")
		assert(validator.is_url("") == false, "should reject empty string")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// is_ip tests
// =============================================================================

func TestIsIPValid(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local validator = require("stdlib.validator")
		assert(validator.is_ip("192.168.1.1") == true, "should validate IPv4")
		assert(validator.is_ip("::1") == true, "should validate IPv6 localhost")
		assert(validator.is_ip("2001:db8::1") == true, "should validate IPv6")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestIsIPInvalid(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local validator = require("stdlib.validator")
		assert(validator.is_ip("256.256.256.256") == false, "should reject invalid IPv4")
		assert(validator.is_ip("not an ip") == false, "should reject non-IP string")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// is_ipv4 / is_ipv6 tests
// =============================================================================

func TestIsIPv4(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local validator = require("stdlib.validator")
		assert(validator.is_ipv4("192.168.1.1") == true, "should validate IPv4")
		assert(validator.is_ipv4("10.0.0.1") == true, "should validate private IPv4")
		assert(validator.is_ipv4("::1") == false, "should reject IPv6")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestIsIPv6(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local validator = require("stdlib.validator")
		assert(validator.is_ipv6("::1") == true, "should validate IPv6 localhost")
		assert(validator.is_ipv6("2001:db8::1") == true, "should validate IPv6")
		assert(validator.is_ipv6("192.168.1.1") == false, "should reject IPv4")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// is_uuid tests
// =============================================================================

func TestIsUUIDValid(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local validator = require("stdlib.validator")
		assert(validator.is_uuid("550e8400-e29b-41d4-a716-446655440000") == true, "should validate UUID v4")
		assert(validator.is_uuid("6ba7b810-9dad-11d1-80b4-00c04fd430c8") == true, "should validate UUID v1")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestIsUUIDInvalid(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local validator = require("stdlib.validator")
		assert(validator.is_uuid("not-a-uuid") == false, "should reject non-UUID string")
		assert(validator.is_uuid("550e8400-e29b-41d4-a716") == false, "should reject incomplete UUID")
		assert(validator.is_uuid("") == false, "should reject empty string")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// is_json tests
// =============================================================================

func TestIsJSONValid(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local validator = require("stdlib.validator")
		assert(validator.is_json('{"key": "value"}') == true, "should validate JSON object")
		assert(validator.is_json('[1, 2, 3]') == true, "should validate JSON array")
		assert(validator.is_json('"string"') == true, "should validate JSON string")
		assert(validator.is_json('123') == true, "should validate JSON number")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestIsJSONInvalid(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local validator = require("stdlib.validator")
		assert(validator.is_json('{invalid}') == false, "should reject invalid JSON")
		assert(validator.is_json('') == false, "should reject empty string")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// is_numeric / is_alpha / is_alphanumeric tests
// =============================================================================

func TestIsNumeric(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local validator = require("stdlib.validator")
		assert(validator.is_numeric("12345") == true, "should validate numeric string")
		assert(validator.is_numeric("0") == true, "should validate zero")
		assert(validator.is_numeric("abc") == false, "should reject alphabetic string")
		assert(validator.is_numeric("12.34") == false, "should reject decimal without option")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestIsAlpha(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local validator = require("stdlib.validator")
		assert(validator.is_alpha("hello") == true, "should validate lowercase alpha")
		assert(validator.is_alpha("HELLO") == true, "should validate uppercase alpha")
		assert(validator.is_alpha("Hello") == true, "should validate mixed case alpha")
		assert(validator.is_alpha("hello123") == false, "should reject alphanumeric")
		assert(validator.is_alpha("") == false, "should reject empty string")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestIsAlphanumeric(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local validator = require("stdlib.validator")
		assert(validator.is_alphanumeric("hello123") == true, "should validate alphanumeric")
		assert(validator.is_alphanumeric("ABC123") == true, "should validate uppercase alphanumeric")
		assert(validator.is_alphanumeric("hello_123") == false, "should reject underscore")
		assert(validator.is_alphanumeric("") == false, "should reject empty string")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// is_credit_card tests
// =============================================================================

func TestIsCreditCardValid(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local validator = require("stdlib.validator")
		-- Test Visa card (Luhn valid)
		assert(validator.is_credit_card("4111111111111111") == true, "should validate Visa test card")
		-- Test Mastercard (Luhn valid)
		assert(validator.is_credit_card("5500000000000004") == true, "should validate Mastercard test card")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestIsCreditCardInvalid(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local validator = require("stdlib.validator")
		assert(validator.is_credit_card("1234567890123456") == false, "should reject invalid Luhn")
		assert(validator.is_credit_card("not a card") == false, "should reject non-numeric")
		assert(validator.is_credit_card("") == false, "should reject empty string")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// is_phone tests
// =============================================================================

func TestIsPhoneValid(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local validator = require("stdlib.validator")
		assert(validator.is_phone("+1-555-555-5555") == true, "should validate US phone with dashes")
		assert(validator.is_phone("+15555555555") == true, "should validate US phone without dashes")
		assert(validator.is_phone("+44 20 7123 4567") == true, "should validate UK phone")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestIsPhoneInvalid(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local validator = require("stdlib.validator")
		assert(validator.is_phone("not a phone") == false, "should reject non-phone string")
		assert(validator.is_phone("123") == false, "should reject too short number")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// is_date tests
// =============================================================================

func TestIsDateValid(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local validator = require("stdlib.validator")
		assert(validator.is_date("2024-01-15", "YYYY-MM-DD") == true, "should validate ISO date")
		assert(validator.is_date("01/15/2024", "MM/DD/YYYY") == true, "should validate US date format")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestIsDateInvalid(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local validator = require("stdlib.validator")
		assert(validator.is_date("not a date", "YYYY-MM-DD") == false, "should reject non-date string")
		assert(validator.is_date("2024-13-45", "YYYY-MM-DD") == false, "should reject invalid date")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// is_base64 / is_hex tests
// =============================================================================

func TestIsBase64Valid(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local validator = require("stdlib.validator")
		assert(validator.is_base64("aGVsbG8=") == true, "should validate base64")
		assert(validator.is_base64("SGVsbG8gV29ybGQ=") == true, "should validate base64 with padding")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestIsBase64Invalid(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local validator = require("stdlib.validator")
		assert(validator.is_base64("not base64!!!") == false, "should reject invalid base64")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestIsHexValid(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local validator = require("stdlib.validator")
		assert(validator.is_hex("deadbeef") == true, "should validate lowercase hex")
		assert(validator.is_hex("DEADBEEF") == true, "should validate uppercase hex")
		assert(validator.is_hex("0123456789abcdef") == true, "should validate full hex charset")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestIsHexInvalid(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local validator = require("stdlib.validator")
		assert(validator.is_hex("ghijkl") == false, "should reject non-hex chars")
		assert(validator.is_hex("0xdeadbeef") == false, "should reject 0x prefix")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// is_slug tests
// =============================================================================

func TestIsSlugValid(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local validator = require("stdlib.validator")
		assert(validator.is_slug("hello-world") == true, "should validate slug with hyphen")
		assert(validator.is_slug("hello-world-123") == true, "should validate slug with numbers")
		assert(validator.is_slug("hello") == true, "should validate simple slug")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestIsSlugInvalid(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local validator = require("stdlib.validator")
		assert(validator.is_slug("Hello World") == false, "should reject spaces and uppercase")
		assert(validator.is_slug("hello_world") == false, "should reject underscores")
		assert(validator.is_slug("hello--world") == false, "should reject consecutive hyphens")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// matches tests
// =============================================================================

func TestMatchesPattern(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local validator = require("stdlib.validator")
		assert(validator.matches("hello123", "^[a-z]+[0-9]+$") == true, "should match pattern")
		assert(validator.matches("HELLO", "^[A-Z]+$") == true, "should match uppercase pattern")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestMatchesPatternFails(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local validator = require("stdlib.validator")
		assert(validator.matches("hello", "^[0-9]+$") == false, "should not match wrong pattern")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// length / range tests
// =============================================================================

func TestLength(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local validator = require("stdlib.validator")
		assert(validator.length("hello", {min = 1, max = 10}) == true, "should be within range")
		assert(validator.length("hi", {min = 5}) == false, "should fail min length")
		assert(validator.length("hello world", {max = 5}) == false, "should fail max length")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestRange(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local validator = require("stdlib.validator")
		assert(validator.range(5, {min = 1, max = 10}) == true, "should be within range")
		assert(validator.range(0, {min = 1}) == false, "should fail min value")
		assert(validator.range(100, {max = 10}) == false, "should fail max value")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
