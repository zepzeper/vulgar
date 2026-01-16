package s3

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

func skipIfNoS3(t *testing.T) {
	if os.Getenv("S3_TEST_ACCESS_KEY") == "" {
		t.Skip("S3_TEST_ACCESS_KEY not set, skipping integration test")
	}
}

// =============================================================================
// configure tests
// =============================================================================

func TestConfigureMissingCredentials(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local s3 = require("integrations.s3")
		local client, err = s3.configure({})
		assert(client == nil or err ~= nil, "should error with missing credentials")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestConfigureValid(t *testing.T) {
	skipIfNoS3(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("access_key", lua.LString(os.Getenv("S3_TEST_ACCESS_KEY")))
	L.SetGlobal("secret_key", lua.LString(os.Getenv("S3_TEST_SECRET_KEY")))
	L.SetGlobal("region", lua.LString(os.Getenv("S3_TEST_REGION")))

	err := L.DoString(`
		local s3 = require("integrations.s3")
		local client, err = s3.configure({
			access_key = access_key,
			secret_key = secret_key,
			region = region
		})
		assert(err == nil, "configure should not error: " .. tostring(err))
		assert(client ~= nil, "client should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// upload tests
// =============================================================================

func TestUploadNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local s3 = require("integrations.s3")
		local err = s3.upload(nil, "bucket", "key", "content")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestUpload(t *testing.T) {
	skipIfNoS3(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("access_key", lua.LString(os.Getenv("S3_TEST_ACCESS_KEY")))
	L.SetGlobal("secret_key", lua.LString(os.Getenv("S3_TEST_SECRET_KEY")))
	L.SetGlobal("region", lua.LString(os.Getenv("S3_TEST_REGION")))
	L.SetGlobal("bucket", lua.LString(os.Getenv("S3_TEST_BUCKET")))

	err := L.DoString(`
		local s3 = require("integrations.s3")
		local client, _ = s3.configure({
			access_key = access_key,
			secret_key = secret_key,
			region = region
		})
		
		local err = s3.upload(client, bucket, "test/upload.txt", "Hello World!", {
			content_type = "text/plain"
		})
		assert(err == nil, "upload should not error: " .. tostring(err))
		
		-- Cleanup
		s3.delete(client, bucket, "test/upload.txt")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// download tests
// =============================================================================

func TestDownloadNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local s3 = require("integrations.s3")
		local content, err = s3.download(nil, "bucket", "key")
		assert(content == nil, "content should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestDownloadNotFound(t *testing.T) {
	skipIfNoS3(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("access_key", lua.LString(os.Getenv("S3_TEST_ACCESS_KEY")))
	L.SetGlobal("secret_key", lua.LString(os.Getenv("S3_TEST_SECRET_KEY")))
	L.SetGlobal("region", lua.LString(os.Getenv("S3_TEST_REGION")))
	L.SetGlobal("bucket", lua.LString(os.Getenv("S3_TEST_BUCKET")))

	err := L.DoString(`
		local s3 = require("integrations.s3")
		local client, _ = s3.configure({
			access_key = access_key,
			secret_key = secret_key,
			region = region
		})
		
		local content, err = s3.download(client, bucket, "nonexistent/key/xyz123")
		assert(content == nil, "content should be nil for missing key")
		assert(err ~= nil, "should error for missing key")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// delete tests
// =============================================================================

func TestDeleteNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local s3 = require("integrations.s3")
		local err = s3.delete(nil, "bucket", "key")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// list tests
// =============================================================================

func TestListNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local s3 = require("integrations.s3")
		local objects, err = s3.list(nil, "bucket")
		assert(objects == nil, "objects should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestListWithPrefix(t *testing.T) {
	skipIfNoS3(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("access_key", lua.LString(os.Getenv("S3_TEST_ACCESS_KEY")))
	L.SetGlobal("secret_key", lua.LString(os.Getenv("S3_TEST_SECRET_KEY")))
	L.SetGlobal("region", lua.LString(os.Getenv("S3_TEST_REGION")))
	L.SetGlobal("bucket", lua.LString(os.Getenv("S3_TEST_BUCKET")))

	err := L.DoString(`
		local s3 = require("integrations.s3")
		local client, _ = s3.configure({
			access_key = access_key,
			secret_key = secret_key,
			region = region
		})
		
		local objects, err = s3.list(client, bucket, {prefix = "test/"})
		assert(err == nil, "list should not error: " .. tostring(err))
		assert(objects ~= nil, "objects should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// exists tests
// =============================================================================

func TestExistsNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local s3 = require("integrations.s3")
		local exists, err = s3.exists(nil, "bucket", "key")
		assert(exists == false, "exists should be false")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// presigned_url tests
// =============================================================================

func TestPresignedUrlNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local s3 = require("integrations.s3")
		local url, err = s3.presigned_url(nil, "bucket", "key")
		assert(url == nil, "url should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestPresignedUrl(t *testing.T) {
	skipIfNoS3(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("access_key", lua.LString(os.Getenv("S3_TEST_ACCESS_KEY")))
	L.SetGlobal("secret_key", lua.LString(os.Getenv("S3_TEST_SECRET_KEY")))
	L.SetGlobal("region", lua.LString(os.Getenv("S3_TEST_REGION")))
	L.SetGlobal("bucket", lua.LString(os.Getenv("S3_TEST_BUCKET")))

	err := L.DoString(`
		local s3 = require("integrations.s3")
		local client, _ = s3.configure({
			access_key = access_key,
			secret_key = secret_key,
			region = region
		})
		
		local url, err = s3.presigned_url(client, bucket, "some/key", {expires = 3600})
		assert(err == nil, "presigned_url should not error: " .. tostring(err))
		assert(url ~= nil, "url should not be nil")
		assert(string.find(url, "http") ~= nil, "url should start with http")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// copy tests
// =============================================================================

func TestCopyNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local s3 = require("integrations.s3")
		local err = s3.copy(nil, "bucket", "src", "bucket", "dst")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
