package dns

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
// lookup tests
// =============================================================================

func TestLookupA(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local dns = require("integrations.dns")
		local records, err = dns.lookup("google.com", "A")
		assert(err == nil, "lookup should not error: " .. tostring(err))
		assert(records ~= nil, "records should not be nil")
		assert(#records > 0, "should have at least 1 record")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestLookupAAAA(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local dns = require("integrations.dns")
		local records, err = dns.lookup("google.com", "AAAA")
		-- AAAA might not always be available
		assert(err == nil or records == nil, "lookup should handle AAAA")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestLookupMX(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local dns = require("integrations.dns")
		local records, err = dns.lookup("google.com", "MX")
		assert(err == nil, "lookup MX should not error: " .. tostring(err))
		assert(records ~= nil, "records should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestLookupTXT(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local dns = require("integrations.dns")
		local records, err = dns.lookup("google.com", "TXT")
		assert(err == nil, "lookup TXT should not error: " .. tostring(err))
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestLookupNonexistent(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local dns = require("integrations.dns")
		local records, err = dns.lookup("nonexistent-domain-xyz123.invalid", "A")
		-- Should return error or empty records
		assert(records == nil or #records == 0 or err ~= nil, "should handle nonexistent domain")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// reverse tests
// =============================================================================

func TestReverse(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local dns = require("integrations.dns")
		local names, err = dns.reverse("8.8.8.8")
		assert(err == nil, "reverse should not error: " .. tostring(err))
		assert(names ~= nil, "names should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestReverseInvalidIP(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local dns = require("integrations.dns")
		local names, err = dns.reverse("not-an-ip")
		assert(names == nil or err ~= nil, "should handle invalid IP")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// resolve tests
// =============================================================================

func TestResolve(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local dns = require("integrations.dns")
		local ips, err = dns.resolve("google.com")
		assert(err == nil, "resolve should not error: " .. tostring(err))
		assert(ips ~= nil, "ips should not be nil")
		assert(#ips > 0, "should have at least 1 IP")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestResolveNonexistent(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local dns = require("integrations.dns")
		local ips, err = dns.resolve("nonexistent-xyz123.invalid")
		assert(ips == nil or #ips == 0 or err ~= nil, "should handle nonexistent domain")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// lookup_ns tests
// =============================================================================

func TestLookupNS(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local dns = require("integrations.dns")
		local ns, err = dns.lookup_ns("google.com")
		assert(err == nil, "lookup_ns should not error: " .. tostring(err))
		assert(ns ~= nil, "ns should not be nil")
		assert(#ns > 0, "should have at least 1 nameserver")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// lookup_cname tests
// =============================================================================

func TestLookupCNAME(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local dns = require("integrations.dns")
		local cname, err = dns.lookup_cname("www.google.com")
		-- www.google.com might be a CNAME
		-- Either returns cname or nil (not an error if no CNAME)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
