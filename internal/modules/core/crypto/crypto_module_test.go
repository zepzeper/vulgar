package crypto

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func setupLuaState() *lua.LState {
	L := lua.NewState()
	L.PreloadModule(ModuleName, Loader)
	return L
}

func TestSha256(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local crypto = require("crypto")
		result = crypto.sha256("hello world")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").String()
	expected := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestSha256EmptyString(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local crypto = require("crypto")
		result = crypto.sha256("")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").String()
	expected := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestSha512(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local crypto = require("crypto")
		result = crypto.sha512("hello world")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").String()
	if len(result) != 128 { // SHA-512 produces 64 bytes = 128 hex chars
		t.Errorf("expected 128 char hex string, got %d chars", len(result))
	}
}

func TestMd5(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local crypto = require("crypto")
		result = crypto.md5("hello world")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").String()
	expected := "5eb63bbbe01eeed093cb22bb8f5acdc3"
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestHmacSha256(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local crypto = require("crypto")
		result = crypto.hmac_sha256("message", "secret")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").String()
	if len(result) != 64 { // HMAC-SHA256 produces 32 bytes = 64 hex chars
		t.Errorf("expected 64 char hex string, got %d chars", len(result))
	}
}

func TestHmacSha256Consistency(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local crypto = require("crypto")
		result1 = crypto.hmac_sha256("data", "key")
		result2 = crypto.hmac_sha256("data", "key")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result1 := L.GetGlobal("result1").String()
	result2 := L.GetGlobal("result2").String()

	if result1 != result2 {
		t.Error("HMAC should produce consistent results for same input")
	}
}

func TestHmacSha256DifferentKeys(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local crypto = require("crypto")
		result1 = crypto.hmac_sha256("data", "key1")
		result2 = crypto.hmac_sha256("data", "key2")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result1 := L.GetGlobal("result1").String()
	result2 := L.GetGlobal("result2").String()

	if result1 == result2 {
		t.Error("HMAC with different keys should produce different results")
	}
}

func TestBase64Encode(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local crypto = require("crypto")
		result = crypto.base64_encode("Hello, World!")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").String()
	expected := "SGVsbG8sIFdvcmxkIQ=="
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestBase64Decode(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local crypto = require("crypto")
		result, err = crypto.base64_decode("SGVsbG8sIFdvcmxkIQ==")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").String()
	errVal := L.GetGlobal("err")

	if result != "Hello, World!" {
		t.Errorf("expected 'Hello, World!', got '%s'", result)
	}
	if errVal != lua.LNil {
		t.Error("expected no error")
	}
}

func TestBase64DecodeInvalid(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local crypto = require("crypto")
		result, err = crypto.base64_decode("not valid base64!!!")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result")
	errVal := L.GetGlobal("err")

	if result != lua.LNil {
		t.Error("expected nil result for invalid base64")
	}
	if errVal == lua.LNil {
		t.Error("expected error for invalid base64")
	}
}

func TestBase64RoundTrip(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local crypto = require("crypto")
		local original = "Test data with special chars: !@#$%"
		local encoded = crypto.base64_encode(original)
		result, err = crypto.base64_decode(encoded)
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").String()
	if result != "Test data with special chars: !@#$%" {
		t.Errorf("round trip failed, got '%s'", result)
	}
}

func TestHexEncode(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local crypto = require("crypto")
		result = crypto.hex_encode("ABC")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").String()
	expected := "414243"
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestHexDecode(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local crypto = require("crypto")
		result, err = crypto.hex_decode("414243")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").String()
	errVal := L.GetGlobal("err")

	if result != "ABC" {
		t.Errorf("expected 'ABC', got '%s'", result)
	}
	if errVal != lua.LNil {
		t.Error("expected no error")
	}
}

func TestHexDecodeInvalid(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local crypto = require("crypto")
		result, err = crypto.hex_decode("not hex!")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result")
	errVal := L.GetGlobal("err")

	if result != lua.LNil {
		t.Error("expected nil result for invalid hex")
	}
	if errVal == lua.LNil {
		t.Error("expected error for invalid hex")
	}
}

func TestHexRoundTrip(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local crypto = require("crypto")
		local original = "Test 123"
		local encoded = crypto.hex_encode(original)
		result, err = crypto.hex_decode(encoded)
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").String()
	if result != "Test 123" {
		t.Errorf("round trip failed, got '%s'", result)
	}
}

func TestRandomBytes(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local crypto = require("crypto")
		result, err = crypto.random_bytes(16)
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").String()
	errVal := L.GetGlobal("err")

	if len(result) != 32 { // 16 bytes = 32 hex chars
		t.Errorf("expected 32 hex chars, got %d", len(result))
	}
	if errVal != lua.LNil {
		t.Error("expected no error")
	}
}

func TestRandomBytesUnique(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local crypto = require("crypto")
		result1, _ = crypto.random_bytes(16)
		result2, _ = crypto.random_bytes(16)
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result1 := L.GetGlobal("result1").String()
	result2 := L.GetGlobal("result2").String()

	if result1 == result2 {
		t.Error("random bytes should be unique")
	}
}

func TestRandomBytesInvalidLength(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local crypto = require("crypto")
		result, err = crypto.random_bytes(0)
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result")
	errVal := L.GetGlobal("err")

	if result != lua.LNil {
		t.Error("expected nil result for zero length")
	}
	if errVal == lua.LNil {
		t.Error("expected error for zero length")
	}
}

func TestLoaderReturnsModule(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule(ModuleName, Loader)
	err := L.DoString(`crypto = require("crypto")`)
	if err != nil {
		t.Fatalf("failed to require module: %v", err)
	}

	mod := L.GetGlobal("crypto")
	if mod.Type() != lua.LTTable {
		t.Errorf("expected table, got %s", mod.Type())
	}
}

func TestLoaderExportsAllFunctions(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule(ModuleName, Loader)
	err := L.DoString(`crypto = require("crypto")`)
	if err != nil {
		t.Fatalf("failed to require module: %v", err)
	}

	tbl := L.GetGlobal("crypto").(*lua.LTable)

	funcs := []string{"sha256", "sha512", "md5", "hmac_sha256", "base64_encode", "base64_decode", "hex_encode", "hex_decode", "random_bytes"}
	for _, name := range funcs {
		fn := L.GetField(tbl, name)
		if fn.Type() != lua.LTFunction {
			t.Errorf("expected %s to be a function", name)
		}
	}
}
