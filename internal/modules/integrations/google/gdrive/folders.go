package gdrive

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules/util"
	"google.golang.org/api/drive/v3"
)

// luaCreateFolder creates a folder
// Usage: local folder, err = gdrive.create_folder(client, {name = "My Folder", parent_id = "..."})
func luaCreateFolder(L *lua.LState) int {
	client := getClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid gdrive client")
	}

	opts := L.CheckTable(2)

	name := ""
	if v := opts.RawGetString("name"); v != lua.LNil {
		name = v.String()
	} else {
		return util.PushError(L, "name is required")
	}

	folderMetadata := &drive.File{
		Name:     name,
		MimeType: "application/vnd.google-apps.folder",
	}

	if v := opts.RawGetString("parent_id"); v != lua.LNil {
		folderMetadata.Parents = []string{v.String()}
	}

	folder, err := client.service.Files.Create(folderMetadata).
		Fields("id, name, webViewLink").
		Do()
	if err != nil {
		return util.PushError(L, "failed to create folder: %v", err)
	}

	result := L.NewTable()
	L.SetField(result, "id", lua.LString(folder.Id))
	L.SetField(result, "name", lua.LString(folder.Name))
	L.SetField(result, "web_view_link", lua.LString(folder.WebViewLink))

	L.Push(result)
	L.Push(lua.LNil)
	return 2
}

// luaMove moves a file to a different folder
// Usage: local err = gdrive.move(client, file_id, new_folder_id)
func luaMove(L *lua.LState) int {
	client := getClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid gdrive client")
	}

	fileID := L.CheckString(2)
	newFolderID := L.CheckString(3)

	// Get current parents
	file, err := client.service.Files.Get(fileID).Fields("parents").Do()
	if err != nil {
		return util.PushError(L, "failed to get file: %v", err)
	}

	// Build remove parents string
	var previousParents string
	for i, parent := range file.Parents {
		if i > 0 {
			previousParents += ","
		}
		previousParents += parent
	}

	// Move file
	_, err = client.service.Files.Update(fileID, nil).
		AddParents(newFolderID).
		RemoveParents(previousParents).
		Do()
	if err != nil {
		return util.PushError(L, "failed to move file: %v", err)
	}

	L.Push(lua.LNil)
	return 1
}

// luaFindFolder finds a folder by name
// Usage: local folder, err = gdrive.find_folder(client, "Reports", {parent_id = "..."})
func luaFindFolder(L *lua.LState) int {
	client := getClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid gdrive client")
	}

	name := L.CheckString(2)
	opts := L.OptTable(3, L.NewTable())

	query := fmt.Sprintf("name = '%s' and mimeType = 'application/vnd.google-apps.folder' and trashed = false", name)

	if v := opts.RawGetString("parent_id"); v != lua.LNil {
		query += fmt.Sprintf(" and '%s' in parents", v.String())
	}

	resp, err := client.service.Files.List().
		Q(query).
		Fields("files(id, name, webViewLink)").
		PageSize(1).
		Do()
	if err != nil {
		return util.PushError(L, "failed to find folder: %v", err)
	}

	if len(resp.Files) == 0 {
		L.Push(lua.LNil)
		L.Push(lua.LString("folder not found: " + name))
		return 2
	}

	folder := resp.Files[0]
	result := L.NewTable()
	L.SetField(result, "id", lua.LString(folder.Id))
	L.SetField(result, "name", lua.LString(folder.Name))
	L.SetField(result, "web_view_link", lua.LString(folder.WebViewLink))

	L.Push(result)
	L.Push(lua.LNil)
	return 2
}

// splitPath splits a path into components, handling leading/trailing slashes
func splitPath(path string) []string {
	var parts []string
	current := ""
	for _, c := range path {
		if c == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}
