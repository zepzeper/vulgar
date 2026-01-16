package xml

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
// parse tests
// =============================================================================

func TestParseSimpleXML(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local xml = require("stdlib.xml")
		local doc, err = xml.parse("<root><item>value</item></root>")
		assert(err == nil, "parse should not error: " .. tostring(err))
		assert(doc ~= nil, "doc should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestParseWithAttributes(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local xml = require("stdlib.xml")
		local doc, err = xml.parse('<item id="1" name="test">content</item>')
		assert(err == nil, "parse should not error: " .. tostring(err))
		assert(doc ~= nil, "doc should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestParseNestedElements(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local xml = require("stdlib.xml")
		local doc, err = xml.parse([[
<root>
  <parent>
    <child>value1</child>
    <child>value2</child>
  </parent>
</root>
]])
		assert(err == nil, "parse should not error: " .. tostring(err))
		assert(doc ~= nil, "doc should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestParseInvalidXML(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local xml = require("stdlib.xml")
		local doc, err = xml.parse("<unclosed>")
		assert(doc == nil, "doc should be nil for invalid xml")
		assert(err ~= nil, "should return error for invalid xml")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestParseEmptyString(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local xml = require("stdlib.xml")
		local doc, err = xml.parse("")
		-- Empty string should either error or return nil
		assert(doc == nil or err ~= nil, "empty string should not produce valid doc")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// decode tests
// =============================================================================

func TestDecodeSimpleXML(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local xml = require("stdlib.xml")
		local tbl, err = xml.decode("<root><name>John</name><age>30</age></root>")
		assert(err == nil, "decode should not error: " .. tostring(err))
		assert(tbl ~= nil, "result should not be nil")
		assert(tbl.root.name == "John", "name should be John")
		assert(tbl.root.age == "30", "age should be 30")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestDecodeWithAttributes(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local xml = require("stdlib.xml")
		local tbl, err = xml.decode('<item id="123">content</item>')
		assert(err == nil, "decode should not error: " .. tostring(err))
		assert(tbl ~= nil, "result should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestDecodeInvalidXML(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local xml = require("stdlib.xml")
		local tbl, err = xml.decode("<invalid>")
		assert(tbl == nil, "result should be nil for invalid xml")
		assert(err ~= nil, "should return error")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// encode tests
// =============================================================================

func TestEncodeSimpleTable(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local xml = require("stdlib.xml")
		local result, err = xml.encode({root = {name = "John", age = "30"}})
		assert(err == nil, "encode should not error: " .. tostring(err))
		assert(result ~= nil, "result should not be nil")
		assert(string.find(result, "root") ~= nil, "should contain root element")
		assert(string.find(result, "name") ~= nil, "should contain name element")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestEncodeEmptyTable(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local xml = require("stdlib.xml")
		local result, err = xml.encode({})
		assert(err == nil, "encode empty should not error")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// xpath tests
// =============================================================================

func TestXPathFindElements(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local xml = require("stdlib.xml")
		local doc, _ = xml.parse("<root><item>1</item><item>2</item></root>")
		local results, err = xml.xpath(doc, "//item")
		assert(err == nil, "xpath should not error: " .. tostring(err))
		assert(#results == 2, "should find 2 items")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestXPathWithPredicate(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local xml = require("stdlib.xml")
		local doc, _ = xml.parse('<root><item id="1">a</item><item id="2">b</item></root>')
		local results, err = xml.xpath(doc, "//item[@id='1']")
		assert(err == nil, "xpath should not error: " .. tostring(err))
		assert(#results == 1, "should find 1 item")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestXPathNoMatch(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local xml = require("stdlib.xml")
		local doc, _ = xml.parse("<root><item>value</item></root>")
		local results, err = xml.xpath(doc, "//nonexistent")
		assert(err == nil, "xpath should not error")
		assert(#results == 0, "should find 0 items")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// xpath_one tests
// =============================================================================

func TestXPathOneSuccess(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local xml = require("stdlib.xml")
		local doc, _ = xml.parse("<root><item>value</item></root>")
		local result, err = xml.xpath_one(doc, "//item")
		assert(err == nil, "xpath_one should not error: " .. tostring(err))
		assert(result ~= nil, "should find element")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestXPathOneNoMatch(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local xml = require("stdlib.xml")
		local doc, _ = xml.parse("<root><item>value</item></root>")
		local result, err = xml.xpath_one(doc, "//nonexistent")
		assert(result == nil, "should return nil for no match")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// attr tests
// =============================================================================

func TestAttrGetValue(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local xml = require("stdlib.xml")
		local doc, _ = xml.parse('<item id="123" name="test">content</item>')
		local elem, _ = xml.xpath_one(doc, "//item")
		local id = xml.attr(elem, "id")
		assert(id == "123", "id should be '123'")
		local name = xml.attr(elem, "name")
		assert(name == "test", "name should be 'test'")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestAttrMissing(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local xml = require("stdlib.xml")
		local doc, _ = xml.parse('<item id="123">content</item>')
		local elem, _ = xml.xpath_one(doc, "//item")
		local value = xml.attr(elem, "nonexistent")
		assert(value == nil, "missing attribute should return nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// text tests
// =============================================================================

func TestTextGetContent(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local xml = require("stdlib.xml")
		local doc, _ = xml.parse("<item>Hello World</item>")
		local elem, _ = xml.xpath_one(doc, "//item")
		local text = xml.text(elem)
		assert(text == "Hello World", "text should be 'Hello World'")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestTextEmpty(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local xml = require("stdlib.xml")
		local doc, _ = xml.parse("<item></item>")
		local elem, _ = xml.xpath_one(doc, "//item")
		local text = xml.text(elem)
		assert(text == "", "empty element should return empty string")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// escape tests
// =============================================================================

func TestEscapeSpecialChars(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local xml = require("stdlib.xml")
		local result = xml.escape("<tag>&value</tag>")
		assert(string.find(result, "&lt;") ~= nil, "should escape <")
		assert(string.find(result, "&gt;") ~= nil, "should escape >")
		assert(string.find(result, "&amp;") ~= nil, "should escape &")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestEscapeQuotes(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local xml = require("stdlib.xml")
		local result = xml.escape('value="test"')
		assert(string.find(result, "&quot;") ~= nil, "should escape quotes")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestEscapeNoSpecialChars(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local xml = require("stdlib.xml")
		local result = xml.escape("plain text")
		assert(result == "plain text", "plain text should remain unchanged")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// format tests
// =============================================================================

func TestFormatPrettyPrint(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local xml = require("stdlib.xml")
		local result = xml.format("<root><item>value</item></root>")
		assert(result ~= nil, "format should return string")
		assert(#result > 0, "formatted result should not be empty")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestFormatAlreadyFormatted(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local xml = require("stdlib.xml")
		local formatted = [[
<root>
  <item>value</item>
</root>
]]
		local result = xml.format(formatted)
		assert(result ~= nil, "format should return string")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
