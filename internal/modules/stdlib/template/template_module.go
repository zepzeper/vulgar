package template

import (
	"bytes"
	"os"
	"text/template"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "stdlib.template"
const templateTypeName = "template_parsed"

// registerTemplateType registers the parsed template userdata type
func registerTemplateType(L *lua.LState) {
	util.RegisterUserDataType(L, templateTypeName, nil)
}

// checkTemplate checks if the userdata is a parsed template
func checkTemplate(L *lua.LState) *template.Template {
	val := L.Get(1)
	if val == lua.LNil {
		return nil
	}
	ud, ok := val.(*lua.LUserData)
	if !ok {
		return nil
	}
	if v, ok := ud.Value.(*template.Template); ok {
		return v
	}
	return nil
}

// luaRender renders a template string with data
// Usage: local result, err = template.render("Hello {{.name}}!", {name = "World"})
func luaRender(L *lua.LState) int {
	tmplStr := L.CheckString(1)
	dataTable := L.CheckTable(2)

	// Convert Lua table to Go map
	data := util.LuaToGo(dataTable).(map[string]interface{})

	// Parse and execute template
	tmpl, err := template.New("").Parse(tmplStr)
	if err != nil {
		return util.PushError(L, "template parse error: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return util.PushError(L, "template execute error: %v", err)
	}

	return util.PushSuccess(L, lua.LString(buf.String()))
}

// luaRenderFile renders a template file with data
// Usage: local result, err = template.render_file("template.tmpl", {name = "World"})
func luaRenderFile(L *lua.LState) int {
	filePath := L.CheckString(1)
	dataTable := L.CheckTable(2)

	// Read file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return util.PushError(L, "file read error: %v", err)
	}

	// Convert Lua table to Go map
	data := util.LuaToGo(dataTable).(map[string]interface{})

	// Parse and execute template
	tmpl, err := template.New("").Parse(string(content))
	if err != nil {
		return util.PushError(L, "template parse error: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return util.PushError(L, "template execute error: %v", err)
	}

	return util.PushSuccess(L, lua.LString(buf.String()))
}

// luaParse parses a template string for reuse
// Usage: local tmpl, err = template.parse("Hello {{.name}}!")
func luaParse(L *lua.LState) int {
	tmplStr := L.CheckString(1)

	tmpl, err := template.New("").Parse(tmplStr)
	if err != nil {
		return util.PushError(L, "template parse error: %v", err)
	}

	// Store template as userdata
	ud := util.NewUserData(L, tmpl, templateTypeName)

	return util.PushSuccess(L, ud)
}

// luaExecute executes a parsed template with data
// Usage: local result, err = template.execute(tmpl, {name = "World"})
func luaExecute(L *lua.LState) int {
	tmpl := checkTemplate(L)
	if tmpl == nil {
		return util.PushError(L, "invalid or nil template")
	}

	dataTable := L.CheckTable(2)

	// Convert Lua table to Go map
	data := util.LuaToGo(dataTable).(map[string]interface{})

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return util.PushError(L, "template execute error: %v", err)
	}

	return util.PushSuccess(L, lua.LString(buf.String()))
}

var exports = map[string]lua.LGFunction{
	"render":      luaRender,
	"render_file": luaRenderFile,
	"parse":       luaParse,
	"execute":     luaExecute,
}

// Loader is called when the module is required via require("template")
func Loader(L *lua.LState) int {
	registerTemplateType(L)
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
