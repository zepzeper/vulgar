package gdrive

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules/util"
	"google.golang.org/api/drive/v3"
)

// luaCopy copies a file
// Usage: local new_file, err = gdrive.copy(client, file_id, {name = "Copy of file", folder_id = "..."})
func luaCopy(L *lua.LState) int {
	client := getClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid gdrive client")
	}

	fileID := L.CheckString(2)
	opts := L.OptTable(3, L.NewTable())

	copyMetadata := &drive.File{}

	if v := opts.RawGetString("name"); v != lua.LNil {
		copyMetadata.Name = v.String()
	}

	if v := opts.RawGetString("folder_id"); v != lua.LNil {
		copyMetadata.Parents = []string{v.String()}
	}

	file, err := client.service.Files.Copy(fileID, copyMetadata).
		Fields("id, name, mimeType, size, webViewLink").
		Do()
	if err != nil {
		return util.PushError(L, "failed to copy file: %v", err)
	}

	result := L.NewTable()
	L.SetField(result, "id", lua.LString(file.Id))
	L.SetField(result, "name", lua.LString(file.Name))
	L.SetField(result, "mime_type", lua.LString(file.MimeType))
	L.SetField(result, "size", lua.LNumber(file.Size))
	L.SetField(result, "web_view_link", lua.LString(file.WebViewLink))

	L.Push(result)
	L.Push(lua.LNil)
	return 2
}

// luaShare shares a file with a user
// Usage: local err = gdrive.share(client, file_id, {email = "user@example.com", role = "reader"})
func luaShare(L *lua.LState) int {
	client := getClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid gdrive client")
	}

	fileID := L.CheckString(2)
	opts := L.CheckTable(3)

	email := ""
	if v := opts.RawGetString("email"); v != lua.LNil {
		email = v.String()
	} else {
		return util.PushError(L, "email is required")
	}

	role := "reader"
	if v := opts.RawGetString("role"); v != lua.LNil {
		role = v.String()
	}

	permType := "user"
	if v := opts.RawGetString("type"); v != lua.LNil {
		permType = v.String()
	}

	permission := &drive.Permission{
		Type:         permType,
		Role:         role,
		EmailAddress: email,
	}

	_, err := client.service.Permissions.Create(fileID, permission).Do()
	if err != nil {
		return util.PushError(L, "failed to share file: %v", err)
	}

	L.Push(lua.LNil)
	return 1
}
