package gzip

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
// compress tests
// =============================================================================

func TestCompressString(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local gzip = require("stdlib.gzip")
		local data = "Hello, World! This is some test data for compression."
		local compressed, err = gzip.compress(data)
		assert(err == nil, "compress should not error: " .. tostring(err))
		assert(compressed ~= nil, "compressed data should not be nil")
		assert(#compressed > 0, "compressed data should not be empty")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestCompressLargeData(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local gzip = require("stdlib.gzip")
		-- Create large repetitive data that should compress well
		local data = string.rep("Hello, World! ", 1000)
		local compressed, err = gzip.compress(data)
		assert(err == nil, "compress should not error: " .. tostring(err))
		assert(#compressed < #data, "compressed should be smaller than original for repetitive data")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestCompressEmpty(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local gzip = require("stdlib.gzip")
		local compressed, err = gzip.compress("")
		assert(err == nil, "compress empty should not error")
		assert(compressed ~= nil, "compressed should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestCompressBinaryData(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local gzip = require("stdlib.gzip")
		-- Binary data with null bytes
		local data = "Hello\0World\0\xff\xfe"
		local compressed, err = gzip.compress(data)
		assert(err == nil, "compress should handle binary data")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// decompress tests
// =============================================================================

func TestDecompressValid(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local gzip = require("stdlib.gzip")
		local original = "Hello, World!"
		local compressed, _ = gzip.compress(original)
		local decompressed, err = gzip.decompress(compressed)
		assert(err == nil, "decompress should not error: " .. tostring(err))
		assert(decompressed == original, "decompressed should match original")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestDecompressLargeData(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local gzip = require("stdlib.gzip")
		local original = string.rep("Test data for compression ", 500)
		local compressed, _ = gzip.compress(original)
		local decompressed, err = gzip.decompress(compressed)
		assert(err == nil, "decompress should not error: " .. tostring(err))
		assert(decompressed == original, "decompressed should match original")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestDecompressInvalidData(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local gzip = require("stdlib.gzip")
		local data, err = gzip.decompress("not valid gzip data")
		assert(data == nil, "data should be nil for invalid gzip")
		assert(err ~= nil, "should return error for invalid data")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestDecompressEmpty(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local gzip = require("stdlib.gzip")
		local data, err = gzip.decompress("")
		-- Empty input should either error or return empty
		assert(err ~= nil or data == "" or data == nil, "should handle empty input")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// round trip tests
// =============================================================================

func TestRoundTripText(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local gzip = require("stdlib.gzip")
		local original = "The quick brown fox jumps over the lazy dog."
		local compressed, err = gzip.compress(original)
		assert(err == nil, "compress should not error")
		
		local decompressed, err = gzip.decompress(compressed)
		assert(err == nil, "decompress should not error")
		assert(decompressed == original, "round trip should preserve data exactly")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestRoundTripUnicode(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local gzip = require("stdlib.gzip")
		local original = "Hello, 世界! Привет мир! 🌍"
		local compressed, err = gzip.compress(original)
		assert(err == nil, "compress should handle unicode")
		
		local decompressed, err = gzip.decompress(compressed)
		assert(err == nil, "decompress should not error")
		assert(decompressed == original, "unicode should be preserved")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestRoundTripJSON(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local gzip = require("stdlib.gzip")
		local original = '{"name": "John", "age": 30, "items": [1, 2, 3]}'
		local compressed, _ = gzip.compress(original)
		local decompressed, err = gzip.decompress(compressed)
		assert(err == nil, "decompress should not error")
		assert(decompressed == original, "JSON should be preserved")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
