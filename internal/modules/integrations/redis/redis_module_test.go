package redis

import (
	"os"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func newTestState() *lua.LState {
	L := lua.NewState()
	L.PreloadModule(ModuleName, Loader)
	return L
}

func skipIfNoRedis(t *testing.T) (string, string) {
	host := os.Getenv("REDIS_TEST_HOST")
	port := os.Getenv("REDIS_TEST_PORT")
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "6379"
	}
	if os.Getenv("REDIS_TEST_SKIP") != "" {
		t.Skip("REDIS_TEST_SKIP set, skipping integration test")
	}
	return host, port
}

// =============================================================================
// connect tests
// =============================================================================

func TestConnectDefault(t *testing.T) {
	host, port := skipIfNoRedis(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("redis_host", lua.LString(host))
	L.SetGlobal("redis_port", lua.LString(port))

	err := L.DoString(`
		local redis = require("integrations.redis")
		local client, err = redis.connect({host = redis_host, port = tonumber(redis_port)})
		assert(err == nil, "connect should not error: " .. tostring(err))
		assert(client ~= nil, "client should not be nil")
		redis.close(client)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestConnectInvalidHost(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local redis = require("integrations.redis")
		local client, err = redis.connect({host = "invalid-host-xyz", port = 6379})
		assert(client == nil, "client should be nil for invalid host")
		assert(err ~= nil, "should error for invalid host")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// get / set tests
// =============================================================================

func TestSetAndGet(t *testing.T) {
	host, port := skipIfNoRedis(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("redis_host", lua.LString(host))
	L.SetGlobal("redis_port", lua.LString(port))

	err := L.DoString(`
		local redis = require("integrations.redis")
		local client, _ = redis.connect({host = redis_host, port = tonumber(redis_port)})
		
		local err = redis.set(client, "test_key", "test_value")
		assert(err == nil, "set should not error: " .. tostring(err))
		
		local value, err = redis.get(client, "test_key")
		assert(err == nil, "get should not error: " .. tostring(err))
		assert(value == "test_value", "value should match")
		
		redis.close(client)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestGetMissingKey(t *testing.T) {
	host, port := skipIfNoRedis(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("redis_host", lua.LString(host))
	L.SetGlobal("redis_port", lua.LString(port))

	err := L.DoString(`
		local redis = require("integrations.redis")
		local client, _ = redis.connect({host = redis_host, port = tonumber(redis_port)})
		
		local value, err = redis.get(client, "nonexistent_key_xyz_123")
		-- Should return nil for missing key without erroring
		assert(value == nil, "value should be nil for missing key")
		
		redis.close(client)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestSetWithExpiration(t *testing.T) {
	host, port := skipIfNoRedis(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("redis_host", lua.LString(host))
	L.SetGlobal("redis_port", lua.LString(port))

	err := L.DoString(`
		local redis = require("integrations.redis")
		local client, _ = redis.connect({host = redis_host, port = tonumber(redis_port)})
		
		local err = redis.set(client, "expiring_key", "value", {ex = 3600})
		assert(err == nil, "set with expiration should not error: " .. tostring(err))
		
		redis.close(client)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// del tests
// =============================================================================

func TestDel(t *testing.T) {
	host, port := skipIfNoRedis(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("redis_host", lua.LString(host))
	L.SetGlobal("redis_port", lua.LString(port))

	err := L.DoString(`
		local redis = require("integrations.redis")
		local client, _ = redis.connect({host = redis_host, port = tonumber(redis_port)})
		
		redis.set(client, "to_delete", "value")
		local err = redis.del(client, "to_delete")
		assert(err == nil, "del should not error: " .. tostring(err))
		
		local value, _ = redis.get(client, "to_delete")
		assert(value == nil, "value should be nil after delete")
		
		redis.close(client)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// exists tests
// =============================================================================

func TestExists(t *testing.T) {
	host, port := skipIfNoRedis(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("redis_host", lua.LString(host))
	L.SetGlobal("redis_port", lua.LString(port))

	err := L.DoString(`
		local redis = require("integrations.redis")
		local client, _ = redis.connect({host = redis_host, port = tonumber(redis_port)})
		
		redis.set(client, "exists_key", "value")
		local exists, err = redis.exists(client, "exists_key")
		assert(err == nil, "exists should not error: " .. tostring(err))
		assert(exists == true, "key should exist")
		
		local exists2, _ = redis.exists(client, "nonexistent_key_abc")
		assert(exists2 == false, "nonexistent key should not exist")
		
		redis.close(client)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// expire tests
// =============================================================================

func TestExpire(t *testing.T) {
	host, port := skipIfNoRedis(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("redis_host", lua.LString(host))
	L.SetGlobal("redis_port", lua.LString(port))

	err := L.DoString(`
		local redis = require("integrations.redis")
		local client, _ = redis.connect({host = redis_host, port = tonumber(redis_port)})
		
		redis.set(client, "expire_test", "value")
		local err = redis.expire(client, "expire_test", 3600)
		assert(err == nil, "expire should not error: " .. tostring(err))
		
		redis.close(client)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// hash tests
// =============================================================================

func TestHSetAndHGet(t *testing.T) {
	host, port := skipIfNoRedis(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("redis_host", lua.LString(host))
	L.SetGlobal("redis_port", lua.LString(port))

	err := L.DoString(`
		local redis = require("integrations.redis")
		local client, _ = redis.connect({host = redis_host, port = tonumber(redis_port)})
		
		local err = redis.hset(client, "user:1", "name", "John")
		assert(err == nil, "hset should not error: " .. tostring(err))
		
		local value, err = redis.hget(client, "user:1", "name")
		assert(err == nil, "hget should not error: " .. tostring(err))
		assert(value == "John", "value should be John")
		
		redis.close(client)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestHGetAll(t *testing.T) {
	host, port := skipIfNoRedis(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("redis_host", lua.LString(host))
	L.SetGlobal("redis_port", lua.LString(port))

	err := L.DoString(`
		local redis = require("integrations.redis")
		local client, _ = redis.connect({host = redis_host, port = tonumber(redis_port)})
		
		redis.hset(client, "user:2", "name", "Jane")
		redis.hset(client, "user:2", "age", "30")
		
		local fields, err = redis.hgetall(client, "user:2")
		assert(err == nil, "hgetall should not error: " .. tostring(err))
		assert(fields.name == "Jane", "name should be Jane")
		assert(fields.age == "30", "age should be 30")
		
		redis.close(client)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// list tests
// =============================================================================

func TestListOperations(t *testing.T) {
	host, port := skipIfNoRedis(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("redis_host", lua.LString(host))
	L.SetGlobal("redis_port", lua.LString(port))

	err := L.DoString(`
		local redis = require("integrations.redis")
		local client, _ = redis.connect({host = redis_host, port = tonumber(redis_port)})
		
		-- Clear list first
		redis.del(client, "mylist")
		
		-- Push items
		local err = redis.lpush(client, "mylist", "first")
		assert(err == nil, "lpush should not error: " .. tostring(err))
		
		err = redis.rpush(client, "mylist", "last")
		assert(err == nil, "rpush should not error: " .. tostring(err))
		
		-- Pop items
		local value, err = redis.lpop(client, "mylist")
		assert(err == nil, "lpop should not error: " .. tostring(err))
		assert(value == "first", "lpop should return first item")
		
		value, err = redis.rpop(client, "mylist")
		assert(err == nil, "rpop should not error: " .. tostring(err))
		assert(value == "last", "rpop should return last item")
		
		redis.close(client)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// pubsub tests
// =============================================================================

func TestPublish(t *testing.T) {
	host, port := skipIfNoRedis(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("redis_host", lua.LString(host))
	L.SetGlobal("redis_port", lua.LString(port))

	err := L.DoString(`
		local redis = require("integrations.redis")
		local client, _ = redis.connect({host = redis_host, port = tonumber(redis_port)})
		
		local err = redis.publish(client, "test_channel", "test message")
		assert(err == nil, "publish should not error: " .. tostring(err))
		
		redis.close(client)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// close tests
// =============================================================================

func TestCloseNil(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local redis = require("integrations.redis")
		local err = redis.close(nil)
		-- Should handle nil gracefully
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
