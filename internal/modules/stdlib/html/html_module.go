package html

import (
	ht "html"
	"strings"

	"golang.org/x/net/html"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "stdlib.html"
const htmlDocTypeName = "html_document"
const htmlElementTypeName = "html_element"

// Helper types for userdata
type htmlDocument struct {
	root *html.Node
}

type htmlElement struct {
	node *html.Node
}

// registerTypes registers userdata types
func registerTypes(L *lua.LState) {
	// Document type
	docMT := L.NewTypeMetatable(htmlDocTypeName)
	L.SetField(docMT, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{}))

	// Element type
	elemMT := L.NewTypeMetatable(htmlElementTypeName)
	L.SetField(elemMT, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{}))
}

// checkDocument checks if userdata is a document
func checkDocument(L *lua.LState) *htmlDocument {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*htmlDocument); ok {
		return v
	}
	L.ArgError(1, "html document expected")
	return nil
}

// checkElement checks if userdata is an element
func checkElement(L *lua.LState) *htmlElement {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*htmlElement); ok {
		return v
	}
	L.ArgError(1, "html element expected")
	return nil
}

// extractText extracts all text from a node and its children
func extractText(n *html.Node) string {
	var text strings.Builder
	var f func(*html.Node)
	f = func(node *html.Node) {
		if node.Type == html.TextNode {
			text.WriteString(node.Data)
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	return strings.TrimSpace(text.String())
}

// getAttribute gets an attribute value from a node
func getAttribute(n *html.Node, name string) string {
	for _, attr := range n.Attr {
		if attr.Key == name {
			return attr.Val
		}
	}
	return ""
}

// hasClass checks if a node has a specific class
func hasClass(n *html.Node, className string) bool {
	classAttr := getAttribute(n, "class")
	classes := strings.Fields(classAttr)
	for _, cls := range classes {
		if cls == className {
			return true
		}
	}
	return false
}

// matchesSelector checks if a node matches a simple CSS selector
func matchesSelector(n *html.Node, selector string) bool {
	// Simple selector parsing: #id, .class, tag, tag.class, tag#id
	if strings.HasPrefix(selector, "#") {
		// ID selector
		id := selector[1:]
		return getAttribute(n, "id") == id
	} else if strings.HasPrefix(selector, ".") {
		// Class selector
		className := selector[1:]
		return hasClass(n, className)
	} else if strings.Contains(selector, ".") {
		// Tag.class
		parts := strings.SplitN(selector, ".", 2)
		tagName := parts[0]
		className := parts[1]
		return n.Data == tagName && hasClass(n, className)
	} else if strings.Contains(selector, "#") {
		// Tag#id
		parts := strings.SplitN(selector, "#", 2)
		tagName := parts[0]
		id := parts[1]
		return n.Data == tagName && getAttribute(n, "id") == id
	} else {
		// Tag name
		return n.Data == selector
	}
}

// selectNodes finds all nodes matching a selector
func selectNodes(root *html.Node, selector string) []*html.Node {
	// Handle combinators like "parent > child"
	if strings.Contains(selector, " > ") {
		parts := strings.SplitN(selector, " > ", 2)
		if len(parts) == 2 {
			parentSel := strings.TrimSpace(parts[0])
			childSel := strings.TrimSpace(parts[1])

			// Find all parent nodes
			parents := selectNodes(root, parentSel)
			var results []*html.Node

			// For each parent, find direct children matching child selector
			for _, parent := range parents {
				for c := parent.FirstChild; c != nil; c = c.NextSibling {
					if c.Type == html.ElementNode && matchesSelector(c, childSel) {
						results = append(results, c)
					}
				}
			}
			return results
		}
	}

	// Simple selector - search all nodes
	var results []*html.Node
	var f func(*html.Node)
	f = func(node *html.Node) {
		if node.Type == html.ElementNode && matchesSelector(node, selector) {
			results = append(results, node)
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(root)
	return results
}

// nodeToHTML converts a node to HTML string
func nodeToHTML(n *html.Node) string {
	var buf strings.Builder
	html.Render(&buf, n)
	return buf.String()
}

// innerHTML gets inner HTML of a node
func innerHTML(n *html.Node) string {
	var buf strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		html.Render(&buf, c)
	}
	return buf.String()
}

// luaParse parses HTML and returns a document
// Usage: local doc, err = html.parse("<html><body>Hello</body></html>")
func luaParse(L *lua.LState) int {
	htmlStr := L.CheckString(1)

	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		return util.PushError(L, "parse error: %v", err)
	}

	htmlDoc := &htmlDocument{root: doc}
	ud := util.NewUserData(L, htmlDoc, htmlDocTypeName)

	return util.PushSuccess(L, ud)
}

// luaSelect selects elements using CSS selector
// Usage: local elements, err = html.select(doc, "div.class > a")
func luaSelect(L *lua.LState) int {
	doc := checkDocument(L)
	selector := L.CheckString(2)

	// Simple selector - for now just handle basic selectors
	// Complex selectors like "div.class > a" would need a proper CSS parser
	// For now, handle the most common case: simple selector
	nodes := selectNodes(doc.root, selector)

	tbl := L.NewTable()
	for i, node := range nodes {
		elem := &htmlElement{node: node}
		ud := L.NewUserData()
		ud.Value = elem
		L.SetMetatable(ud, L.GetTypeMetatable(htmlElementTypeName))
		tbl.RawSetInt(i+1, ud)
	}

	return util.PushSuccess(L, tbl)
}

// luaSelectOne selects first matching element
// Usage: local element, err = html.select_one(doc, "#main")
func luaSelectOne(L *lua.LState) int {
	doc := checkDocument(L)
	selector := L.CheckString(2)

	nodes := selectNodes(doc.root, selector)
	if len(nodes) == 0 {
		return util.PushSuccess(L, lua.LNil)
	}

	elem := &htmlElement{node: nodes[0]}
	ud := util.NewUserData(L, elem, htmlElementTypeName)

	return util.PushSuccess(L, ud)
}

// luaText extracts text content from element
// Usage: local text = html.text(element)
func luaText(L *lua.LState) int {
	elem := checkElement(L)
	text := extractText(elem.node)
	L.Push(lua.LString(text))
	return 1
}

// luaAttr gets an attribute from element
// Usage: local href = html.attr(element, "href")
func luaAttr(L *lua.LState) int {
	elem := checkElement(L)
	attrName := L.CheckString(2)

	value := getAttribute(elem.node, attrName)
	if value == "" {
		L.Push(lua.LNil)
	} else {
		L.Push(lua.LString(value))
	}
	return 1
}

// luaInnerHTML gets inner HTML of element
// Usage: local inner = html.inner_html(element)
func luaInnerHTML(L *lua.LState) int {
	elem := checkElement(L)
	inner := innerHTML(elem.node)
	L.Push(lua.LString(inner))
	return 1
}

// luaOuterHTML gets outer HTML of element
// Usage: local outer = html.outer_html(element)
func luaOuterHTML(L *lua.LState) int {
	elem := checkElement(L)
	outer := nodeToHTML(elem.node)
	L.Push(lua.LString(outer))
	return 1
}

// luaEscape escapes HTML special characters
// Usage: local safe = html.escape("<script>alert('xss')</script>")
func luaEscape(L *lua.LState) int {
	text := L.CheckString(1)
	escaped := ht.EscapeString(text)
	L.Push(lua.LString(escaped))
	return 1
}

// luaUnescape unescapes HTML entities
// Usage: local text = html.unescape("&lt;div&gt;")
func luaUnescape(L *lua.LState) int {
	text := L.CheckString(1)
	unescaped := ht.UnescapeString(text)
	L.Push(lua.LString(unescaped))
	return 1
}

// luaStripTags removes HTML tags
// Usage: local text = html.strip_tags("<p>Hello <b>world</b></p>")
func luaStripTags(L *lua.LState) int {
	htmlStr := L.CheckString(1)

	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		L.Push(lua.LString(htmlStr)) // Return original on parse error
		return 1
	}

	text := extractText(doc)
	L.Push(lua.LString(text))
	return 1
}

// luaSanitize sanitizes HTML allowing only safe tags
// Usage: local safe = html.sanitize(html_string, {allow = {"p", "b", "i", "a"}})
func luaSanitize(L *lua.LState) int {
	htmlStr := L.CheckString(1)
	opts := L.CheckTable(2)

	// Get allowed tags
	allowedTags := make(map[string]bool)
	if allowVal := L.GetField(opts, "allow"); allowVal != lua.LNil {
		if allowTbl, ok := allowVal.(*lua.LTable); ok {
			allowTbl.ForEach(func(_, v lua.LValue) {
				if tag, ok := v.(lua.LString); ok {
					allowedTags[strings.ToLower(string(tag))] = true
				}
			})
		}
	}

	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		L.Push(lua.LString("")) // Return empty on parse error
		return 1
	}

	// Build a new sanitized tree
	var sanitize func(*html.Node, *html.Node)
	sanitize = func(src, dst *html.Node) {
		for c := src.FirstChild; c != nil; c = c.NextSibling {
			switch c.Type {
			case html.ElementNode:
				tagName := strings.ToLower(c.Data)
				if allowedTags[tagName] {
					// Clone allowed element
					newNode := &html.Node{
						Type:     c.Type,
						DataAtom: c.DataAtom,
						Data:     c.Data,
						Attr:     make([]html.Attribute, len(c.Attr)),
					}
					copy(newNode.Attr, c.Attr)
					dst.AppendChild(newNode)
					sanitize(c, newNode)
				} else {
					// Skip disallowed element but keep its children
					sanitize(c, dst)
				}
			case html.TextNode:
				// Keep text nodes
				newNode := &html.Node{
					Type: html.TextNode,
					Data: c.Data,
				}
				dst.AppendChild(newNode)
			}
		}
	}

	// Create new document
	newDoc := &html.Node{Type: html.DocumentNode}
	sanitize(doc, newDoc)

	result := nodeToHTML(newDoc)
	L.Push(lua.LString(result))
	return 1
}

var exports = map[string]lua.LGFunction{
	"parse":      luaParse,
	"select":     luaSelect,
	"select_one": luaSelectOne,
	"text":       luaText,
	"attr":       luaAttr,
	"inner_html": luaInnerHTML,
	"outer_html": luaOuterHTML,
	"escape":     luaEscape,
	"unescape":   luaUnescape,
	"strip_tags": luaStripTags,
	"sanitize":   luaSanitize,
}

// Loader is called when the module is required via require("html")
func Loader(L *lua.LState) int {
	registerTypes(L)
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
