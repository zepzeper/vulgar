package xml

import (
	"encoding/xml"
	"fmt"
	"strings"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "stdlib.xml"
const xmlDocType = "xml_doc"
const xmlElementType = "xml_element"

// XMLDoc represents a parsed XML document
type XMLDoc struct {
	Root *XMLElement
}

// XMLElement represents an XML element
type XMLElement struct {
	Name       string
	Attributes map[string]string
	Children   []*XMLElement
	Text       string
	Parent     *XMLElement
}

// parseXMLString parses XML string into XMLElement tree
func parseXMLString(xmlStr string) (*XMLElement, error) {
	decoder := xml.NewDecoder(strings.NewReader(xmlStr))

	var root *XMLElement
	var stack []*XMLElement

	for {
		token, err := decoder.Token()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, err
		}

		switch t := token.(type) {
		case xml.StartElement:
			elem := &XMLElement{
				Name:       t.Name.Local,
				Attributes: make(map[string]string),
				Children:   []*XMLElement{},
			}
			for _, attr := range t.Attr {
				elem.Attributes[attr.Name.Local] = attr.Value
			}
			if root == nil {
				root = elem
			} else if len(stack) > 0 {
				parent := stack[len(stack)-1]
				elem.Parent = parent
				parent.Children = append(parent.Children, elem)
			}
			stack = append(stack, elem)
		case xml.EndElement:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		case xml.CharData:
			if len(stack) > 0 {
				text := strings.TrimSpace(string(t))
				if text != "" {
					if stack[len(stack)-1].Text == "" {
						stack[len(stack)-1].Text = text
					} else {
						stack[len(stack)-1].Text += " " + text
					}
				}
			}
		}
	}

	return root, nil
}

// elementToLua converts XMLElement to Lua table
func elementToLua(L *lua.LState, elem *XMLElement) lua.LValue {
	if elem == nil {
		return L.NewTable()
	}

	// If element has only text and no children, return text directly
	if len(elem.Children) == 0 && elem.Text != "" {
		return lua.LString(elem.Text)
	}

	tbl := L.NewTable()

	// Group children by name
	childrenMap := make(map[string][]*XMLElement)
	for _, child := range elem.Children {
		childrenMap[child.Name] = append(childrenMap[child.Name], child)
	}

	// Add children
	for name, children := range childrenMap {
		if len(children) == 1 {
			// Single child - use name directly
			tbl.RawSetString(name, elementToLua(L, children[0]))
		} else {
			// Multiple children - use array
			arr := L.NewTable()
			for i, child := range children {
				arr.RawSetInt(i+1, elementToLua(L, child))
			}
			tbl.RawSetString(name, arr)
		}
	}

	// Add text if present (for mixed content)
	if elem.Text != "" && len(elem.Children) > 0 {
		tbl.RawSetString("_text", lua.LString(elem.Text))
	}

	// Add attributes if present
	if len(elem.Attributes) > 0 {
		attrs := L.NewTable()
		for k, v := range elem.Attributes {
			attrs.RawSetString(k, lua.LString(v))
		}
		tbl.RawSetString("_attrs", attrs)
	}

	return tbl
}

// luaToElement converts Lua table to XMLElement
func luaToElement(L *lua.LState, name string, tbl *lua.LTable) *XMLElement {
	elem := &XMLElement{
		Name:       name,
		Attributes: make(map[string]string),
		Children:   []*XMLElement{},
	}

	// Get text content
	if textVal := tbl.RawGetString("_text"); textVal != lua.LNil {
		elem.Text = lua.LVAsString(textVal)
	}

	// Get attributes
	if attrsVal := tbl.RawGetString("_attrs"); attrsVal != lua.LNil {
		if attrsTbl, ok := attrsVal.(*lua.LTable); ok {
			attrsTbl.ForEach(func(k, v lua.LValue) {
				elem.Attributes[lua.LVAsString(k)] = lua.LVAsString(v)
			})
		}
	}

	// Get children
	tbl.ForEach(func(k, v lua.LValue) {
		keyStr := lua.LVAsString(k)
		if keyStr == "_text" || keyStr == "_attrs" {
			return
		}
		if childTbl, ok := v.(*lua.LTable); ok {
			// Check if it's an array
			isArray := true
			childTbl.ForEach(func(ck, _ lua.LValue) {
				if _, ok := ck.(lua.LNumber); !ok {
					isArray = false
				}
			})
			if isArray {
				// Multiple children with same name
				childTbl.ForEach(func(_, cv lua.LValue) {
					if ctbl, ok := cv.(*lua.LTable); ok {
						child := luaToElement(L, keyStr, ctbl)
						child.Parent = elem
						elem.Children = append(elem.Children, child)
					} else {
						// String value in array
						child := &XMLElement{
							Name:       keyStr,
							Attributes: make(map[string]string),
							Children:   []*XMLElement{},
							Text:       lua.LVAsString(cv),
							Parent:     elem,
						}
						elem.Children = append(elem.Children, child)
					}
				})
			} else {
				// Single child
				child := luaToElement(L, keyStr, childTbl)
				child.Parent = elem
				elem.Children = append(elem.Children, child)
			}
		} else {
			// String value - create simple element with text
			child := &XMLElement{
				Name:       keyStr,
				Attributes: make(map[string]string),
				Children:   []*XMLElement{},
				Text:       lua.LVAsString(v),
				Parent:     elem,
			}
			elem.Children = append(elem.Children, child)
		}
	})

	return elem
}

// elementToXMLString converts XMLElement to XML string
func elementToXMLString(elem *XMLElement, indent string) string {
	if elem == nil {
		return ""
	}

	var buf strings.Builder
	buf.WriteString("<" + elem.Name)

	// Write attributes
	for k, v := range elem.Attributes {
		// Escape XML special chars in attribute values
		escaped := strings.ReplaceAll(v, "&", "&amp;")
		escaped = strings.ReplaceAll(escaped, "<", "&lt;")
		escaped = strings.ReplaceAll(escaped, ">", "&gt;")
		escaped = strings.ReplaceAll(escaped, "\"", "&quot;")
		escaped = strings.ReplaceAll(escaped, "'", "&apos;")
		buf.WriteString(fmt.Sprintf(` %s="%s"`, k, escaped))
	}

	if len(elem.Children) == 0 && elem.Text == "" {
		buf.WriteString("/>")
		return buf.String()
	}

	buf.WriteString(">")

	// Write text
	if elem.Text != "" {
		// Escape XML special chars in text
		escaped := strings.ReplaceAll(elem.Text, "&", "&amp;")
		escaped = strings.ReplaceAll(escaped, "<", "&lt;")
		escaped = strings.ReplaceAll(escaped, ">", "&gt;")
		escaped = strings.ReplaceAll(escaped, "\"", "&quot;")
		escaped = strings.ReplaceAll(escaped, "'", "&apos;")
		buf.WriteString(escaped)
	}

	// Write children
	for _, child := range elem.Children {
		buf.WriteString("\n" + indent + "  ")
		buf.WriteString(elementToXMLString(child, indent+"  "))
	}

	if len(elem.Children) > 0 {
		buf.WriteString("\n" + indent)
	}

	buf.WriteString("</" + elem.Name + ">")
	return buf.String()
}

// simpleXPathQuery performs a simplified XPath query
func simpleXPathQuery(elem *XMLElement, query string) []*XMLElement {
	var results []*XMLElement

	if elem == nil {
		return results
	}

	// Simple //tag query
	if after, ok := strings.CutPrefix(query, "//"); ok {
		tagName := after
		// Remove predicate if present
		if idx := strings.Index(tagName, "["); idx != -1 {
			tagName = tagName[:idx]
		}
		collectElements(elem, tagName, &results)
		return results
	}

	// Simple /tag query
	if after, ok := strings.CutPrefix(query, "/"); ok {
		tagName := after
		if idx := strings.Index(tagName, "["); idx != -1 {
			tagName = tagName[:idx]
		}
		if elem.Name == tagName {
			results = append(results, elem)
		}
		return results
	}

	// Simple tag name
	if elem.Name == query {
		results = append(results, elem)
	}

	return results
}

// collectElements recursively collects elements with given name
func collectElements(elem *XMLElement, name string, results *[]*XMLElement) {
	if elem == nil {
		return
	}
	if elem.Name == name {
		*results = append(*results, elem)
	}
	for _, child := range elem.Children {
		collectElements(child, name, results)
	}
}

// luaParse parses XML string into a document
// Usage: local doc, err = xml.parse("<root><item>value</item></root>")
func luaParse(L *lua.LState) int {
	xmlString := L.CheckString(1)

	if strings.TrimSpace(xmlString) == "" {
		return util.PushError(L, "xml parse error: empty string")
	}

	root, err := parseXMLString(xmlString)
	if err != nil {
		return util.PushError(L, "xml parse error: %v", err)
	}

	if root == nil {
		return util.PushError(L, "xml parse error: no root element")
	}

	doc := &XMLDoc{Root: root}
	ud := util.NewUserData(L, doc, xmlDocType)
	return util.PushSuccess(L, ud)
}

// luaEncode converts a Lua table to XML string
// Usage: local xml_str, err = xml.encode({root = {item = "value"}})
func luaEncode(L *lua.LState) int {
	dataTable := L.CheckTable(1)

	// Find root element (should have one key)
	var rootName string
	var rootTable *lua.LTable
	dataTable.ForEach(func(k, v lua.LValue) {
		if rootName == "" {
			rootName = lua.LVAsString(k)
			if tbl, ok := v.(*lua.LTable); ok {
				rootTable = tbl
			} else {
				// If value is not a table, create a simple element
				rootTable = L.NewTable()
				rootTable.RawSetString("_text", v)
			}
		}
	})

	if rootName == "" {
		// Empty table - return empty XML
		return util.PushSuccess(L, lua.LString(""))
	}

	if rootTable == nil {
		rootTable = L.NewTable()
	}

	root := luaToElement(L, rootName, rootTable)
	xmlStr := elementToXMLString(root, "")

	return util.PushSuccess(L, lua.LString(xmlStr))
}

// luaDecode converts XML to Lua table
// Usage: local tbl, err = xml.decode("<root><item>value</item></root>")
func luaDecode(L *lua.LState) int {
	xmlString := L.CheckString(1)

	root, err := parseXMLString(xmlString)
	if err != nil {
		return util.PushError(L, "xml decode error: %v", err)
	}

	// Create a table with root element name
	result := L.NewTable()
	rootValue := elementToLua(L, root)
	result.RawSetString(root.Name, rootValue)

	return util.PushSuccess(L, result)
}

// luaXPath queries XML using XPath expression
// Usage: local results, err = xml.xpath(doc, "//item[@id='1']")
func luaXPath(L *lua.LState) int {
	docVal := L.Get(1)
	query := L.CheckString(2)

	var root *XMLElement
	if ud, ok := docVal.(*lua.LUserData); ok {
		if doc, ok := ud.Value.(*XMLDoc); ok {
			root = doc.Root
		} else if elem, ok := ud.Value.(*XMLElement); ok {
			root = elem
		}
	}

	if root == nil {
		return util.PushError(L, "xml xpath error: invalid document or element")
	}

	results := simpleXPathQuery(root, query)

	// Filter by predicate if present
	if strings.Contains(query, "[@") {
		// Simple attribute predicate
		parts := strings.Split(query, "[@")
		if len(parts) > 1 {
			pred := strings.TrimSuffix(parts[1], "]")
			if idx := strings.Index(pred, "='"); idx != -1 {
				attrName := pred[:idx]
				attrValue := strings.Trim(pred[idx+2:], "'\"")
				filtered := []*XMLElement{}
				for _, elem := range results {
					if val, ok := elem.Attributes[attrName]; ok && val == attrValue {
						filtered = append(filtered, elem)
					}
				}
				results = filtered
			}
		}
	}

	resultTable := L.NewTable()
	for i, elem := range results {
		ud := util.NewUserData(L, elem, xmlElementType)
		resultTable.RawSetInt(i+1, ud)
	}

	return util.PushSuccess(L, resultTable)
}

// luaXPathOne queries and returns first match
// Usage: local result, err = xml.xpath_one(doc, "//item[@id='1']")
func luaXPathOne(L *lua.LState) int {
	docVal := L.Get(1)
	query := L.CheckString(2)

	var root *XMLElement
	if ud, ok := docVal.(*lua.LUserData); ok {
		if doc, ok := ud.Value.(*XMLDoc); ok {
			root = doc.Root
		} else if elem, ok := ud.Value.(*XMLElement); ok {
			root = elem
		}
	}

	if root == nil {
		return util.PushError(L, "xml xpath_one error: invalid document or element")
	}

	results := simpleXPathQuery(root, query)

	// Filter by predicate if present
	if strings.Contains(query, "[@") {
		parts := strings.Split(query, "[@")
		if len(parts) > 1 {
			pred := strings.TrimSuffix(parts[1], "]")
			if idx := strings.Index(pred, "='"); idx != -1 {
				attrName := pred[:idx]
				attrValue := strings.Trim(pred[idx+2:], "'\"")
				for _, elem := range results {
					if val, ok := elem.Attributes[attrName]; ok && val == attrValue {
						ud := util.NewUserData(L, elem, xmlElementType)
						return util.PushSuccess(L, ud)
					}
				}
				return util.PushSuccess(L, lua.LNil)
			}
		}
	}

	if len(results) > 0 {
		ud := util.NewUserData(L, results[0], xmlElementType)
		return util.PushSuccess(L, ud)
	}

	return util.PushSuccess(L, lua.LNil)
}

// luaGetAttr gets attribute from XML element
// Usage: local value = xml.attr(element, "id")
func luaGetAttr(L *lua.LState) int {
	elemVal := L.Get(1)
	attrName := L.CheckString(2)

	var elem *XMLElement
	if ud, ok := elemVal.(*lua.LUserData); ok {
		if e, ok := ud.Value.(*XMLElement); ok {
			elem = e
		}
	}

	if elem == nil {
		L.Push(lua.LNil)
		return 1
	}

	if val, ok := elem.Attributes[attrName]; ok {
		L.Push(lua.LString(val))
	} else {
		L.Push(lua.LNil)
	}
	return 1
}

// luaGetText gets text content from XML element
// Usage: local text = xml.text(element)
func luaGetText(L *lua.LState) int {
	elemVal := L.Get(1)

	var elem *XMLElement
	if ud, ok := elemVal.(*lua.LUserData); ok {
		if e, ok := ud.Value.(*XMLElement); ok {
			elem = e
		}
	}

	if elem == nil {
		L.Push(lua.LString(""))
		return 1
	}

	L.Push(lua.LString(elem.Text))
	return 1
}

// luaEscape escapes XML special characters
// Usage: local safe = xml.escape("<value>")
func luaEscape(L *lua.LState) int {
	str := L.CheckString(1)
	// XML escaping: & < > " '
	escaped := strings.ReplaceAll(str, "&", "&amp;")
	escaped = strings.ReplaceAll(escaped, "<", "&lt;")
	escaped = strings.ReplaceAll(escaped, ">", "&gt;")
	escaped = strings.ReplaceAll(escaped, "\"", "&quot;")
	escaped = strings.ReplaceAll(escaped, "'", "&apos;")
	L.Push(lua.LString(escaped))
	return 1
}

// luaFormat pretty-prints XML
// Usage: local formatted = xml.format(xml_string)
func luaFormat(L *lua.LState) int {
	xmlString := L.CheckString(1)

	root, err := parseXMLString(xmlString)
	if err != nil {
		return util.PushError(L, "xml format error: %v", err)
	}

	formatted := elementToXMLString(root, "")
	L.Push(lua.LString(formatted))
	return 1
}

var exports = map[string]lua.LGFunction{
	"parse":     luaParse,
	"encode":    luaEncode,
	"decode":    luaDecode,
	"xpath":     luaXPath,
	"xpath_one": luaXPathOne,
	"attr":      luaGetAttr,
	"text":      luaGetText,
	"escape":    luaEscape,
	"format":    luaFormat,
}

// Loader is called when the module is required via require("xml")
func Loader(L *lua.LState) int {
	// Register userdata types
	util.RegisterUserDataType(L, xmlDocType, nil)
	util.RegisterUserDataType(L, xmlElementType, nil)

	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
