package html

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

func TestParseSimpleHTML(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local html = require("stdlib.html")
		local doc, err = html.parse("<html><body>Hello</body></html>")
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
		local html = require("stdlib.html")
		local doc, err = html.parse('<div id="main" class="container"><p>Text</p></div>')
		assert(err == nil, "parse should not error: " .. tostring(err))
		assert(doc ~= nil, "doc should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestParseComplexHTML(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local html = require("stdlib.html")
		local doc, err = html.parse([[
<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body>
<div id="content">
  <ul class="items">
    <li>Item 1</li>
    <li>Item 2</li>
  </ul>
</div>
</body>
</html>
]])
		assert(err == nil, "parse should not error: " .. tostring(err))
		assert(doc ~= nil, "doc should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestParseInvalidHTML(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local html = require("stdlib.html")
		-- HTML parsers are usually lenient, so this might still parse
		local doc, err = html.parse("<div>unclosed")
		-- Check that it doesn't crash at least
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// select tests
// =============================================================================

func TestSelectById(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local html = require("stdlib.html")
		local doc, _ = html.parse('<div id="main"><p>Hello</p></div>')
		local elements, err = html.select(doc, "#main")
		assert(err == nil, "select should not error: " .. tostring(err))
		assert(#elements == 1, "should find 1 element with id main")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestSelectByClass(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local html = require("stdlib.html")
		local doc, _ = html.parse('<div class="item">1</div><div class="item">2</div><div>3</div>')
		local elements, err = html.select(doc, ".item")
		assert(err == nil, "select should not error: " .. tostring(err))
		assert(#elements == 2, "should find 2 elements with class item")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestSelectByTag(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local html = require("stdlib.html")
		local doc, _ = html.parse('<ul><li>1</li><li>2</li><li>3</li></ul>')
		local elements, err = html.select(doc, "li")
		assert(err == nil, "select should not error: " .. tostring(err))
		assert(#elements == 3, "should find 3 li elements")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestSelectComplex(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local html = require("stdlib.html")
		local doc, _ = html.parse('<div class="container"><a href="#">Link</a></div>')
		local elements, err = html.select(doc, "div.container > a")
		assert(err == nil, "select should not error: " .. tostring(err))
		assert(#elements == 1, "should find 1 direct child link")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestSelectNoMatch(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local html = require("stdlib.html")
		local doc, _ = html.parse('<div>content</div>')
		local elements, err = html.select(doc, ".nonexistent")
		assert(err == nil, "select should not error")
		assert(#elements == 0, "should find 0 elements")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// select_one tests
// =============================================================================

func TestSelectOneFound(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local html = require("stdlib.html")
		local doc, _ = html.parse('<div id="main">content</div>')
		local elem, err = html.select_one(doc, "#main")
		assert(err == nil, "select_one should not error: " .. tostring(err))
		assert(elem ~= nil, "should find element")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestSelectOneNotFound(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local html = require("stdlib.html")
		local doc, _ = html.parse('<div>content</div>')
		local elem, err = html.select_one(doc, "#nonexistent")
		assert(elem == nil, "should return nil for no match")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// text tests
// =============================================================================

func TestTextSimple(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local html = require("stdlib.html")
		local doc, _ = html.parse('<p>Hello World</p>')
		local elem, _ = html.select_one(doc, "p")
		local text = html.text(elem)
		assert(text == "Hello World", "should extract text content")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestTextNested(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local html = require("stdlib.html")
		local doc, _ = html.parse('<div><span>Hello</span> <span>World</span></div>')
		local elem, _ = html.select_one(doc, "div")
		local text = html.text(elem)
		-- Should contain both "Hello" and "World"
		assert(string.find(text, "Hello") ~= nil, "should contain Hello")
		assert(string.find(text, "World") ~= nil, "should contain World")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// attr tests
// =============================================================================

func TestAttrExists(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local html = require("stdlib.html")
		local doc, _ = html.parse('<a href="https://example.com" class="link">Click</a>')
		local elem, _ = html.select_one(doc, "a")
		local href = html.attr(elem, "href")
		assert(href == "https://example.com", "should get href attribute")
		local class = html.attr(elem, "class")
		assert(class == "link", "should get class attribute")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestAttrMissing(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local html = require("stdlib.html")
		local doc, _ = html.parse('<div>content</div>')
		local elem, _ = html.select_one(doc, "div")
		local attr = html.attr(elem, "nonexistent")
		assert(attr == nil, "should return nil for missing attribute")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// inner_html / outer_html tests
// =============================================================================

func TestInnerHTML(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local html = require("stdlib.html")
		local doc, _ = html.parse('<div><span>Hello</span></div>')
		local elem, _ = html.select_one(doc, "div")
		local inner = html.inner_html(elem)
		assert(string.find(inner, "span") ~= nil, "inner HTML should contain span")
		assert(string.find(inner, "Hello") ~= nil, "inner HTML should contain Hello")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestOuterHTML(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local html = require("stdlib.html")
		local doc, _ = html.parse('<div id="test"><span>Hello</span></div>')
		local elem, _ = html.select_one(doc, "div")
		local outer = html.outer_html(elem)
		assert(string.find(outer, "div") ~= nil, "outer HTML should contain div")
		assert(string.find(outer, "id") ~= nil, "outer HTML should contain id attribute")
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
		local html = require("stdlib.html")
		local result = html.escape("<script>alert('xss')</script>")
		assert(string.find(result, "&lt;") ~= nil, "should escape <")
		assert(string.find(result, "&gt;") ~= nil, "should escape >")
		assert(string.find(result, "<script>") == nil, "should not contain raw script tag")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestEscapeAmpersand(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local html = require("stdlib.html")
		local result = html.escape("foo & bar")
		assert(string.find(result, "&amp;") ~= nil, "should escape &")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestEscapeSafeText(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local html = require("stdlib.html")
		local result = html.escape("Hello World")
		assert(result == "Hello World", "safe text should be unchanged")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// unescape tests
// =============================================================================

func TestUnescapeEntities(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local html = require("stdlib.html")
		local result = html.unescape("&lt;div&gt;")
		assert(result == "<div>", "should unescape entities")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestUnescapeAmpersand(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local html = require("stdlib.html")
		local result = html.unescape("foo &amp; bar")
		assert(result == "foo & bar", "should unescape &amp;")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// strip_tags tests
// =============================================================================

func TestStripTags(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local html = require("stdlib.html")
		local result = html.strip_tags("<p>Hello <b>world</b></p>")
		assert(result == "Hello world", "should strip all HTML tags")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestStripTagsNested(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local html = require("stdlib.html")
		local result = html.strip_tags("<div><p><span>Text</span></p></div>")
		assert(result == "Text", "should strip nested tags")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestStripTagsNoTags(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local html = require("stdlib.html")
		local result = html.strip_tags("Plain text")
		assert(result == "Plain text", "plain text should be unchanged")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// sanitize tests
// =============================================================================

func TestSanitizeAllowedTags(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local html = require("stdlib.html")
		local result = html.sanitize("<p>Hello <b>world</b></p><script>alert('xss')</script>", {allow = {"p", "b"}})
		assert(string.find(result, "<p>") ~= nil, "should keep allowed p tag")
		assert(string.find(result, "<b>") ~= nil, "should keep allowed b tag")
		assert(string.find(result, "<script>") == nil, "should remove script tag")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestSanitizeNoAllowedTags(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local html = require("stdlib.html")
		local result = html.sanitize("<p>Hello</p>", {allow = {}})
		assert(string.find(result, "<p>") == nil, "should remove all tags")
		assert(string.find(result, "Hello") ~= nil, "should keep text content")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
