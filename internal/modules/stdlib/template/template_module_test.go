package template

import (
	"os"
	"path/filepath"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func newTestState() *lua.LState {
	L := lua.NewState()
	L.PreloadModule(ModuleName, Loader)
	return L
}

// =============================================================================
// render tests
// =============================================================================

func TestRenderSimpleVariable(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local template = require("stdlib.template")
		local result, err = template.render("Hello {{.name}}!", {name = "World"})
		assert(err == nil, "render should not error: " .. tostring(err))
		assert(result == "Hello World!", "should render variable")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestRenderMultipleVariables(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local template = require("stdlib.template")
		local result, err = template.render("{{.greeting}} {{.name}}!", {greeting = "Hello", name = "World"})
		assert(err == nil, "render should not error: " .. tostring(err))
		assert(result == "Hello World!", "should render multiple variables")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestRenderNestedData(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local template = require("stdlib.template")
		local result, err = template.render("{{.user.name}} ({{.user.age}})", {
			user = {name = "John", age = 30}
		})
		assert(err == nil, "render should not error: " .. tostring(err))
		assert(result == "John (30)", "should render nested data")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestRenderWithCondition(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local template = require("stdlib.template")
		local result, err = template.render("{{if .active}}Active{{else}}Inactive{{end}}", {active = true})
		assert(err == nil, "render should not error: " .. tostring(err))
		assert(result == "Active", "should render conditional")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestRenderWithRange(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local template = require("stdlib.template")
		local result, err = template.render("{{range .items}}{{.}} {{end}}", {items = {"a", "b", "c"}})
		assert(err == nil, "render should not error: " .. tostring(err))
		assert(string.find(result, "a") ~= nil, "should contain 'a'")
		assert(string.find(result, "b") ~= nil, "should contain 'b'")
		assert(string.find(result, "c") ~= nil, "should contain 'c'")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestRenderEmptyData(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local template = require("stdlib.template")
		local result, err = template.render("Hello World!", {})
		assert(err == nil, "render should not error")
		assert(result == "Hello World!", "should render static content")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestRenderMissingVariable(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local template = require("stdlib.template")
		local result, err = template.render("Hello {{.missing}}!", {name = "World"})
		-- Should either error or render empty for missing variable
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestRenderInvalidTemplate(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local template = require("stdlib.template")
		local result, err = template.render("Hello {{.name", {name = "World"})
		assert(result == nil or err ~= nil, "should error on invalid template")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// render_file tests
// =============================================================================

func TestRenderFileSuccess(t *testing.T) {
	L := newTestState()
	defer L.Close()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "template.tmpl")
	content := []byte("Hello {{.name}}!")
	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	L.SetGlobal("test_file", lua.LString(tmpFile))

	err := L.DoString(`
		local template = require("stdlib.template")
		local result, err = template.render_file(test_file, {name = "World"})
		assert(err == nil, "render_file should not error: " .. tostring(err))
		assert(result == "Hello World!", "should render file template")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestRenderFileNotFound(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local template = require("stdlib.template")
		local result, err = template.render_file("/nonexistent/template.tmpl", {name = "World"})
		assert(result == nil, "result should be nil when file not found")
		assert(err ~= nil, "should return error for missing file")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestRenderFileComplexTemplate(t *testing.T) {
	L := newTestState()
	defer L.Close()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "complex.tmpl")
	content := []byte(`
User: {{.user.name}}
Age: {{.user.age}}
{{if .user.active}}Status: Active{{else}}Status: Inactive{{end}}
Items:
{{range .items}}- {{.}}
{{end}}`)
	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	L.SetGlobal("test_file", lua.LString(tmpFile))

	err := L.DoString(`
		local template = require("stdlib.template")
		local result, err = template.render_file(test_file, {
			user = {name = "John", age = 30, active = true},
			items = {"apple", "banana"}
		})
		assert(err == nil, "render_file should not error: " .. tostring(err))
		assert(string.find(result, "John") ~= nil, "should contain user name")
		assert(string.find(result, "Active") ~= nil, "should show active status")
		assert(string.find(result, "apple") ~= nil, "should contain items")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// parse tests
// =============================================================================

func TestParseSuccess(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local template = require("stdlib.template")
		local tmpl, err = template.parse("Hello {{.name}}!")
		assert(err == nil, "parse should not error: " .. tostring(err))
		assert(tmpl ~= nil, "tmpl should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestParseInvalid(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local template = require("stdlib.template")
		local tmpl, err = template.parse("Hello {{.name")
		assert(tmpl == nil, "tmpl should be nil for invalid template")
		assert(err ~= nil, "should return error for invalid template")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestParseComplex(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local template = require("stdlib.template")
		local tmpl, err = template.parse([[
{{if .condition}}
  {{range .items}}
    Item: {{.}}
  {{end}}
{{end}}
]])
		assert(err == nil, "parse should not error: " .. tostring(err))
		assert(tmpl ~= nil, "tmpl should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// execute tests
// =============================================================================

func TestExecuteParsedTemplate(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local template = require("stdlib.template")
		local tmpl, err = template.parse("Hello {{.name}}!")
		assert(err == nil, "parse should not error")
		
		local result, err = template.execute(tmpl, {name = "World"})
		assert(err == nil, "execute should not error: " .. tostring(err))
		assert(result == "Hello World!", "should execute template")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestExecuteMultipleTimes(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local template = require("stdlib.template")
		local tmpl, err = template.parse("Hello {{.name}}!")
		assert(err == nil, "parse should not error")
		
		local result1, err = template.execute(tmpl, {name = "Alice"})
		assert(result1 == "Hello Alice!", "first execution should work")
		
		local result2, err = template.execute(tmpl, {name = "Bob"})
		assert(result2 == "Hello Bob!", "second execution should work")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestExecuteWithNilTemplate(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local template = require("stdlib.template")
		local result, err = template.execute(nil, {name = "World"})
		assert(result == nil or err ~= nil, "should error with nil template")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestExecuteWithEmptyData(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local template = require("stdlib.template")
		local tmpl, _ = template.parse("Static content")
		local result, err = template.execute(tmpl, {})
		assert(err == nil, "execute with empty data should not error")
		assert(result == "Static content", "should render static content")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
