package cache

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
// get / set tests
// =============================================================================

func TestSetAndGetString(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local cache = require("stdlib.cache")
		cache.set("key1", "value1")
		local value, found = cache.get("key1")
		assert(found == true, "should find cached key")
		assert(value == "value1", "should retrieve correct value")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestSetAndGetNumber(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local cache = require("stdlib.cache")
		cache.set("count", 42)
		local value, found = cache.get("count")
		assert(found == true, "should find cached key")
		assert(value == 42, "should retrieve correct number")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestSetAndGetTable(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local cache = require("stdlib.cache")
		cache.set("user", {name = "John", age = 30})
		local value, found = cache.get("user")
		assert(found == true, "should find cached key")
		assert(value.name == "John", "should retrieve correct table data")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestGetMissingKey(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local cache = require("stdlib.cache")
		local value, found = cache.get("nonexistent")
		assert(found == false, "should not find missing key")
		assert(value == nil, "value should be nil for missing key")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestSetWithTTL(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local cache = require("stdlib.cache")
		cache.set("temp_key", "temp_value", {ttl = 3600})
		local value, found = cache.get("temp_key")
		assert(found == true, "should find key within TTL")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// delete tests
// =============================================================================

func TestDelete(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local cache = require("stdlib.cache")
		cache.set("to_delete", "value")
		local _, found = cache.get("to_delete")
		assert(found == true, "key should exist before delete")
		
		cache.delete("to_delete")
		local _, found2 = cache.get("to_delete")
		assert(found2 == false, "key should not exist after delete")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestDeleteNonexistent(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local cache = require("stdlib.cache")
		-- Should not error when deleting nonexistent key
		cache.delete("nonexistent")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// exists tests
// =============================================================================

func TestExists(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local cache = require("stdlib.cache")
		cache.set("existing", "value")
		assert(cache.exists("existing") == true, "should return true for existing key")
		assert(cache.exists("nonexistent") == false, "should return false for nonexistent key")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// clear tests
// =============================================================================

func TestClear(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local cache = require("stdlib.cache")
		cache.set("key1", "value1")
		cache.set("key2", "value2")
		cache.set("key3", "value3")
		
		cache.clear()
		
		assert(cache.exists("key1") == false, "key1 should be cleared")
		assert(cache.exists("key2") == false, "key2 should be cleared")
		assert(cache.exists("key3") == false, "key3 should be cleared")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// keys tests
// =============================================================================

func TestKeys(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local cache = require("stdlib.cache")
		cache.clear()
		cache.set("key1", "value1")
		cache.set("key2", "value2")
		
		local keys = cache.keys()
		assert(#keys == 2, "should return 2 keys")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestKeysEmpty(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local cache = require("stdlib.cache")
		cache.clear()
		local keys = cache.keys()
		assert(#keys == 0, "should return 0 keys for empty cache")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// size tests
// =============================================================================

func TestSize(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local cache = require("stdlib.cache")
		cache.clear()
		assert(cache.size() == 0, "empty cache should have size 0")
		
		cache.set("key1", "value1")
		assert(cache.size() == 1, "should have size 1")
		
		cache.set("key2", "value2")
		assert(cache.size() == 2, "should have size 2")
		
		cache.delete("key1")
		assert(cache.size() == 1, "should have size 1 after delete")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// get_or_set tests
// =============================================================================

func TestGetOrSetNew(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local cache = require("stdlib.cache")
		cache.clear()
		local computed = false
		local value = cache.get_or_set("new_key", function()
			computed = true
			return "computed_value"
		end)
		assert(computed == true, "should have computed value")
		assert(value == "computed_value", "should return computed value")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestGetOrSetExisting(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local cache = require("stdlib.cache")
		cache.clear()
		cache.set("existing_key", "existing_value")
		
		local computed = false
		local value = cache.get_or_set("existing_key", function()
			computed = true
			return "new_value"
		end)
		assert(computed == false, "should not compute for existing key")
		assert(value == "existing_value", "should return existing value")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// increment / decrement tests
// =============================================================================

func TestIncrement(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local cache = require("stdlib.cache")
		cache.set("counter", 10)
		local new_value = cache.increment("counter", 5)
		assert(new_value == 15, "should increment by 5")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestIncrementNonexistent(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local cache = require("stdlib.cache")
		cache.clear()
		local new_value = cache.increment("new_counter", 1)
		-- Should either create with value 1 or return 0/nil
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestDecrement(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local cache = require("stdlib.cache")
		cache.set("counter", 10)
		local new_value = cache.decrement("counter", 3)
		assert(new_value == 7, "should decrement by 3")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// ttl / expire tests
// =============================================================================

func TestTTL(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local cache = require("stdlib.cache")
		cache.set("ttl_key", "value", {ttl = 3600})
		local ttl = cache.ttl("ttl_key")
		assert(ttl > 0, "TTL should be positive")
		assert(ttl <= 3600, "TTL should not exceed original")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestTTLNonexistent(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local cache = require("stdlib.cache")
		local ttl = cache.ttl("nonexistent")
		assert(ttl == -1 or ttl == nil, "TTL for nonexistent key should be -1 or nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestExpire(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local cache = require("stdlib.cache")
		cache.set("expire_key", "value")
		cache.expire("expire_key", 1800)
		local ttl = cache.ttl("expire_key")
		assert(ttl > 0, "should have positive TTL after expire")
		assert(ttl <= 1800, "TTL should not exceed set value")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
