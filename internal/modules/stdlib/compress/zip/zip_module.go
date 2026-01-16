package zip

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "stdlib.zip"

// Usage: local err = zip.create(output_path, {source_paths...})
func luaCreate(L *lua.LState) int {
	outputPath := L.CheckString(1)
	sourcesTable := L.CheckTable(2)

	var sources []string
	sourcesTable.ForEach(func(_, v lua.LValue) {
		if str, ok := v.(lua.LString); ok {
			sources = append(sources, string(str))
		}
	})

	if len(sources) == 0 {
		L.Push(lua.LString("no source files provided"))
		return 1
	}

	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}
	defer outFile.Close()

	writer := zip.NewWriter(outFile)
	defer writer.Close()

	for _, source := range sources {
		if err := addToZip(writer, source, ""); err != nil {
			L.Push(lua.LString(err.Error()))
			return 1
		}
	}

	L.Push(lua.LNil)
	return 1
}

func addToZip(writer *zip.Writer, source, baseInZip string) error {
	info, err := os.Stat(source)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return addDirToZip(writer, source, baseInZip)
	}
	return addFileToZip(writer, source, baseInZip)
}

func addFileToZip(writer *zip.Writer, filePath, baseInZip string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	if baseInZip != "" {
		header.Name = filepath.Join(baseInZip, filepath.Base(filePath))
	} else {
		header.Name = filepath.Base(filePath)
	}

	header.Method = zip.Deflate

	w, err := writer.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(w, file)
	return err
}

func addDirToZip(writer *zip.Writer, dirPath, baseInZip string) error {
	dirName := filepath.Base(dirPath)
	if baseInZip != "" {
		baseInZip = filepath.Join(baseInZip, dirName)
	} else {
		baseInZip = dirName
	}

	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = filepath.Join(baseInZip, relPath)
		header.Method = zip.Deflate

		w, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}

		_, err = io.Copy(w, file)
		return err
	})
}

// Usage: local err = zip.extract(archive_path, dest_dir)
func luaExtract(L *lua.LState) int {
	archivePath := L.CheckString(1)
	destDir := L.CheckString(2)

	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}
	defer reader.Close()

	// Create destination directory
	if err := os.MkdirAll(destDir, 0755); err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	for _, file := range reader.File {
		destPath := filepath.Join(destDir, file.Name)

		// Check for zip slip vulnerability
		if !isValidPath(destDir, destPath) {
			L.Push(lua.LString("invalid file path in archive"))
			return 1
		}

		if file.FileInfo().IsDir() {
			os.MkdirAll(destPath, file.Mode())
			continue
		}

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			L.Push(lua.LString(err.Error()))
			return 1
		}

		if err := extractFile(file, destPath); err != nil {
			L.Push(lua.LString(err.Error()))
			return 1
		}
	}

	L.Push(lua.LNil)
	return 1
}

func extractFile(file *zip.File, destPath string) error {
	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	outFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, rc)
	return err
}

func isValidPath(base, target string) bool {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return false
	}
	return rel != ".." && !filepath.IsAbs(rel) && rel[:2] != ".."
}

// Usage: local files, err = zip.list(archive_path)
func luaList(L *lua.LState) int {
	archivePath := L.CheckString(1)

	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return util.PushError(L, "failed to open archive: %v", err)
	}
	defer reader.Close()

	files := L.NewTable()
	for i, file := range reader.File {
		fileInfo := L.NewTable()
		fileInfo.RawSetString("name", lua.LString(file.Name))
		fileInfo.RawSetString("size", lua.LNumber(file.UncompressedSize64))
		fileInfo.RawSetString("compressed_size", lua.LNumber(file.CompressedSize64))
		fileInfo.RawSetString("is_dir", lua.LBool(file.FileInfo().IsDir()))

		files.RawSetInt(i+1, fileInfo)
	}

	return util.PushSuccess(L, files)
}

// Usage: local err = zip.add_file(archive_path, file_path, archive_name)
func luaAddFile(L *lua.LState) int {
	archivePath := L.CheckString(1)
	filePath := L.CheckString(2)
	archiveName := L.CheckString(3)

	// Read existing archive
	existingReader, err := zip.OpenReader(archivePath)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	// Create temp file for new archive
	tmpFile, err := os.CreateTemp("", "zip-*.zip")
	if err != nil {
		existingReader.Close()
		L.Push(lua.LString(err.Error()))
		return 1
	}
	tmpPath := tmpFile.Name()

	writer := zip.NewWriter(tmpFile)

	// Copy existing files
	for _, file := range existingReader.File {
		if err := copyZipFile(writer, file); err != nil {
			writer.Close()
			tmpFile.Close()
			existingReader.Close()
			os.Remove(tmpPath)
			L.Push(lua.LString(err.Error()))
			return 1
		}
	}

	existingReader.Close()

	// Add new file
	newFile, err := os.Open(filePath)
	if err != nil {
		writer.Close()
		tmpFile.Close()
		os.Remove(tmpPath)
		L.Push(lua.LString(err.Error()))
		return 1
	}

	info, err := newFile.Stat()
	if err != nil {
		newFile.Close()
		writer.Close()
		tmpFile.Close()
		os.Remove(tmpPath)
		L.Push(lua.LString(err.Error()))
		return 1
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		newFile.Close()
		writer.Close()
		tmpFile.Close()
		os.Remove(tmpPath)
		L.Push(lua.LString(err.Error()))
		return 1
	}

	header.Name = archiveName
	header.Method = zip.Deflate

	w, err := writer.CreateHeader(header)
	if err != nil {
		newFile.Close()
		writer.Close()
		tmpFile.Close()
		os.Remove(tmpPath)
		L.Push(lua.LString(err.Error()))
		return 1
	}

	_, err = io.Copy(w, newFile)
	newFile.Close()

	if err != nil {
		writer.Close()
		tmpFile.Close()
		os.Remove(tmpPath)
		L.Push(lua.LString(err.Error()))
		return 1
	}

	writer.Close()
	tmpFile.Close()

	// Replace original with temp
	if err := os.Rename(tmpPath, archivePath); err != nil {
		os.Remove(tmpPath)
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

func copyZipFile(writer *zip.Writer, file *zip.File) error {
	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	w, err := writer.CreateHeader(&file.FileHeader)
	if err != nil {
		return err
	}

	_, err = io.Copy(w, rc)
	return err
}

var exports = map[string]lua.LGFunction{
	"create":   luaCreate,
	"extract":  luaExtract,
	"list":     luaList,
	"add_file": luaAddFile,
}

func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

func init() {
	modules.Register(ModuleName, Loader)
}
