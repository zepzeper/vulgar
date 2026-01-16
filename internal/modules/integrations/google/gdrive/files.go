package gdrive

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules/util"
	"google.golang.org/api/drive/v3"
)

// luaListFiles lists files in Drive
// Usage: local files, err = gdrive.list_files(client, {query = "name contains 'report'", folder_id = "..."})
func luaListFiles(L *lua.LState) int {
	client := getClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid gdrive client")
	}

	opts := L.OptTable(2, L.NewTable())

	// Build query
	call := client.service.Files.List().
		Fields("files(id, name, mimeType, size, createdTime, modifiedTime, parents, webViewLink)")

	// Apply filters
	if v := opts.RawGetString("query"); v != lua.LNil {
		call = call.Q(v.String())
	}

	if v := opts.RawGetString("folder_id"); v != lua.LNil {
		folderID := v.String()
		q := fmt.Sprintf("'%s' in parents", folderID)
		if existing := opts.RawGetString("query"); existing != lua.LNil {
			q = existing.String() + " and " + q
		}
		call = call.Q(q)
	}

	if v := opts.RawGetString("page_size"); v != lua.LNil {
		if num, ok := v.(lua.LNumber); ok {
			call = call.PageSize(int64(num))
		}
	} else {
		call = call.PageSize(100)
	}

	if v := opts.RawGetString("order_by"); v != lua.LNil {
		call = call.OrderBy(v.String())
	}

	// Execute request
	resp, err := call.Do()
	if err != nil {
		return util.PushError(L, "failed to list files: %v", err)
	}

	// Convert response to Lua table
	result := L.NewTable()
	for i, file := range resp.Files {
		fileTable := L.NewTable()
		L.SetField(fileTable, "id", lua.LString(file.Id))
		L.SetField(fileTable, "name", lua.LString(file.Name))
		L.SetField(fileTable, "mime_type", lua.LString(file.MimeType))
		L.SetField(fileTable, "size", lua.LNumber(file.Size))
		L.SetField(fileTable, "created_time", lua.LString(file.CreatedTime))
		L.SetField(fileTable, "modified_time", lua.LString(file.ModifiedTime))
		L.SetField(fileTable, "web_view_link", lua.LString(file.WebViewLink))

		if len(file.Parents) > 0 {
			parents := L.NewTable()
			for j, parent := range file.Parents {
				parents.RawSetInt(j+1, lua.LString(parent))
			}
			L.SetField(fileTable, "parents", parents)
		}

		result.RawSetInt(i+1, fileTable)
	}

	L.Push(result)
	L.Push(lua.LNil)
	return 2
}

// luaGetFile gets file metadata
// Usage: local file, err = gdrive.get_file(client, file_id)
func luaGetFile(L *lua.LState) int {
	client := getClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid gdrive client")
	}

	fileID := L.CheckString(2)

	file, err := client.service.Files.Get(fileID).
		Fields("id, name, mimeType, size, createdTime, modifiedTime, parents, webViewLink, description").
		Do()
	if err != nil {
		return util.PushError(L, "failed to get file: %v", err)
	}

	result := L.NewTable()
	L.SetField(result, "id", lua.LString(file.Id))
	L.SetField(result, "name", lua.LString(file.Name))
	L.SetField(result, "mime_type", lua.LString(file.MimeType))
	L.SetField(result, "size", lua.LNumber(file.Size))
	L.SetField(result, "created_time", lua.LString(file.CreatedTime))
	L.SetField(result, "modified_time", lua.LString(file.ModifiedTime))
	L.SetField(result, "web_view_link", lua.LString(file.WebViewLink))
	L.SetField(result, "description", lua.LString(file.Description))

	L.Push(result)
	L.Push(lua.LNil)
	return 2
}

// luaDownload downloads a file's content
// Usage: local content, err = gdrive.download(client, file_id)
// Usage: local err = gdrive.download(client, file_id, {path = "/tmp/file.txt"})
func luaDownload(L *lua.LState) int {
	client := getClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid gdrive client")
	}

	fileID := L.CheckString(2)
	opts := L.OptTable(3, nil)

	// Download the file
	resp, err := client.service.Files.Get(fileID).Download()
	if err != nil {
		return util.PushError(L, "failed to download file: %v", err)
	}
	defer resp.Body.Close()

	// Read content
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return util.PushError(L, "failed to read file content: %v", err)
	}

	// If path is provided, write to file
	if opts != nil {
		if v := opts.RawGetString("path"); v != lua.LNil {
			destPath := v.String()
			if err := os.WriteFile(destPath, content, 0644); err != nil {
				return util.PushError(L, "failed to write file: %v", err)
			}
			L.Push(lua.LNil)
			return 1
		}
	}

	// Return content as string
	L.Push(lua.LString(content))
	L.Push(lua.LNil)
	return 2
}

// luaUpload uploads a file to Drive
// Usage: local file, err = gdrive.upload(client, {name = "file.txt", path = "/tmp/file.txt", folder_id = "..."})
// Usage: local file, err = gdrive.upload(client, {name = "file.txt", content = "...", folder_id = "..."})
func luaUpload(L *lua.LState) int {
	client := getClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid gdrive client")
	}

	opts := L.CheckTable(2)

	// Get file name
	name := ""
	if v := opts.RawGetString("name"); v != lua.LNil {
		name = v.String()
	} else {
		return util.PushError(L, "name is required")
	}

	// Get content (from path or inline)
	var reader io.Reader
	var err error

	if v := opts.RawGetString("path"); v != lua.LNil {
		filePath := v.String()
		file, err := os.Open(filePath)
		if err != nil {
			return util.PushError(L, "failed to open file: %v", err)
		}
		defer file.Close()
		reader = file

		// Use filename from path if name not specified
		if name == "" {
			name = filepath.Base(filePath)
		}
	} else if v := opts.RawGetString("content"); v != lua.LNil {
		content := v.String()
		// Create a simple reader from string
		reader = &stringReader{content: content}
	} else {
		return util.PushError(L, "path or content is required")
	}

	// Build file metadata
	fileMetadata := &drive.File{
		Name: name,
	}

	// Set folder
	if v := opts.RawGetString("folder_id"); v != lua.LNil {
		fileMetadata.Parents = []string{v.String()}
	}

	// Set mime type
	if v := opts.RawGetString("mime_type"); v != lua.LNil {
		fileMetadata.MimeType = v.String()
	}

	// Set description
	if v := opts.RawGetString("description"); v != lua.LNil {
		fileMetadata.Description = v.String()
	}

	// Upload
	file, err := client.service.Files.Create(fileMetadata).
		Media(reader).
		Fields("id, name, mimeType, size, webViewLink").
		Do()
	if err != nil {
		return util.PushError(L, "failed to upload file: %v", err)
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

// stringReader implements io.Reader for a string
type stringReader struct {
	content string
	pos     int
}

func (r *stringReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.content) {
		return 0, io.EOF
	}
	n = copy(p, r.content[r.pos:])
	r.pos += n
	return n, nil
}

// luaDelete deletes a file
// Usage: local err = gdrive.delete(client, file_id)
func luaDelete(L *lua.LState) int {
	client := getClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid gdrive client")
	}

	fileID := L.CheckString(2)

	err := client.service.Files.Delete(fileID).Do()
	if err != nil {
		return util.PushError(L, "failed to delete file: %v", err)
	}

	L.Push(lua.LNil)
	return 1
}

// luaRename renames a file
// Usage: local file, err = gdrive.rename(client, file_id, "new_name.txt")
func luaRename(L *lua.LState) int {
	client := getClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid gdrive client")
	}

	fileID := L.CheckString(2)
	newName := L.CheckString(3)

	file, err := client.service.Files.Update(fileID, &drive.File{Name: newName}).
		Fields("id, name, mimeType, webViewLink").
		Do()
	if err != nil {
		return util.PushError(L, "failed to rename file: %v", err)
	}

	result := L.NewTable()
	L.SetField(result, "id", lua.LString(file.Id))
	L.SetField(result, "name", lua.LString(file.Name))
	L.SetField(result, "mime_type", lua.LString(file.MimeType))
	L.SetField(result, "web_view_link", lua.LString(file.WebViewLink))

	L.Push(result)
	L.Push(lua.LNil)
	return 2
}

// luaSearch searches for files
// Usage: local files, err = gdrive.search(client, "name contains 'report' and mimeType = 'application/pdf'")
func luaSearch(L *lua.LState) int {
	client := getClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid gdrive client")
	}

	query := L.CheckString(2)

	resp, err := client.service.Files.List().
		Q(query).
		Fields("files(id, name, mimeType, size, createdTime, modifiedTime, webViewLink)").
		PageSize(100).
		Do()
	if err != nil {
		return util.PushError(L, "failed to search: %v", err)
	}

	result := L.NewTable()
	for i, file := range resp.Files {
		fileTable := L.NewTable()
		L.SetField(fileTable, "id", lua.LString(file.Id))
		L.SetField(fileTable, "name", lua.LString(file.Name))
		L.SetField(fileTable, "mime_type", lua.LString(file.MimeType))
		L.SetField(fileTable, "size", lua.LNumber(file.Size))
		L.SetField(fileTable, "created_time", lua.LString(file.CreatedTime))
		L.SetField(fileTable, "modified_time", lua.LString(file.ModifiedTime))
		L.SetField(fileTable, "web_view_link", lua.LString(file.WebViewLink))
		result.RawSetInt(i+1, fileTable)
	}

	L.Push(result)
	L.Push(lua.LNil)
	return 2
}

// luaFindByName finds a file or folder by exact name
// Usage: local file, err = gdrive.find_by_name(client, "report.pdf", {folder_id = "..."})
func luaFindByName(L *lua.LState) int {
	client := getClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid gdrive client")
	}

	name := L.CheckString(2)
	opts := L.OptTable(3, L.NewTable())

	// Build query for exact name match
	query := fmt.Sprintf("name = '%s' and trashed = false", name)

	if v := opts.RawGetString("folder_id"); v != lua.LNil {
		query += fmt.Sprintf(" and '%s' in parents", v.String())
	}

	if v := opts.RawGetString("mime_type"); v != lua.LNil {
		query += fmt.Sprintf(" and mimeType = '%s'", v.String())
	}

	resp, err := client.service.Files.List().
		Q(query).
		Fields("files(id, name, mimeType, size, createdTime, modifiedTime, parents, webViewLink)").
		PageSize(1).
		Do()
	if err != nil {
		return util.PushError(L, "failed to find file: %v", err)
	}

	if len(resp.Files) == 0 {
		L.Push(lua.LNil)
		L.Push(lua.LString("file not found: " + name))
		return 2
	}

	file := resp.Files[0]
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

// luaFindByPath finds a file by path (e.g., "/folder/subfolder/file.txt")
// Usage: local file, err = gdrive.find_by_path(client, "/Reports/2025/January/report.pdf")
func luaFindByPath(L *lua.LState) int {
	client := getClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid gdrive client")
	}

	path := L.CheckString(2)

	// Split path into components
	parts := splitPath(path)
	if len(parts) == 0 {
		return util.PushError(L, "invalid path")
	}

	// Start from root
	parentID := "root"

	// Navigate through each folder
	for i, part := range parts {
		isLastPart := i == len(parts)-1

		query := fmt.Sprintf("name = '%s' and '%s' in parents and trashed = false", part, parentID)

		// If not the last part, it should be a folder
		if !isLastPart {
			query += " and mimeType = 'application/vnd.google-apps.folder'"
		}

		resp, err := client.service.Files.List().
			Q(query).
			Fields("files(id, name, mimeType, size, webViewLink)").
			PageSize(1).
			Do()
		if err != nil {
			return util.PushError(L, "failed to find '%s': %v", part, err)
		}

		if len(resp.Files) == 0 {
			L.Push(lua.LNil)
			L.Push(lua.LString(fmt.Sprintf("not found: %s (at '%s')", path, part)))
			return 2
		}

		if isLastPart {
			// Return the found file
			file := resp.Files[0]
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

		// Move to next folder
		parentID = resp.Files[0].Id
	}

	return util.PushError(L, "unexpected error navigating path")
}
