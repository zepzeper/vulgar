package tar

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "stdlib.tar"

// Usage: local err = tar.create(output_path, {source_paths...})
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

	writer := tar.NewWriter(outFile)
	defer writer.Close()

	for _, source := range sources {
		if err := addToTar(writer, source, ""); err != nil {
			L.Push(lua.LString(err.Error()))
			return 1
		}
	}

	L.Push(lua.LNil)
	return 1
}

func addToTar(writer *tar.Writer, source, baseInTar string) error {
	info, err := os.Stat(source)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return addDirToTar(writer, source, baseInTar)
	}
	return addFileToTar(writer, source, baseInTar)
}

func addFileToTar(writer *tar.Writer, filePath, baseInTar string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}

	if baseInTar != "" {
		header.Name = filepath.Join(baseInTar, filepath.Base(filePath))
	} else {
		header.Name = filepath.Base(filePath)
	}

	if err := writer.WriteHeader(header); err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}

func addDirToTar(writer *tar.Writer, dirPath, baseInTar string) error {
	dirName := filepath.Base(dirPath)
	if baseInTar != "" {
		baseInTar = filepath.Join(baseInTar, dirName)
	} else {
		baseInTar = dirName
	}

	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		header.Name = filepath.Join(baseInTar, relPath)

		if err := writer.WriteHeader(header); err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})
}

// Usage: local err = tar.extract(archive_path, dest_dir)
func luaExtract(L *lua.LState) int {
	archivePath := L.CheckString(1)
	destDir := L.CheckString(2)

	file, err := os.Open(archivePath)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}
	defer file.Close()

	reader := tar.NewReader(file)

	// Create destination directory
	if err := os.MkdirAll(destDir, 0755); err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			L.Push(lua.LString(err.Error()))
			return 1
		}

		destPath := filepath.Join(destDir, header.Name)

		// Check for path traversal vulnerability
		if !isValidPath(destDir, destPath) {
			L.Push(lua.LString("invalid file path in archive"))
			return 1
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(destPath, os.FileMode(header.Mode)); err != nil {
				L.Push(lua.LString(err.Error()))
				return 1
			}
		case tar.TypeReg:
			// Create parent directories
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				L.Push(lua.LString(err.Error()))
				return 1
			}

			if err := extractTarFile(reader, destPath, header); err != nil {
				L.Push(lua.LString(err.Error()))
				return 1
			}
		}
	}

	L.Push(lua.LNil)
	return 1
}

func extractTarFile(reader *tar.Reader, destPath string, header *tar.Header) error {
	outFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, reader)
	return err
}

func isValidPath(base, target string) bool {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return false
	}
	return rel != ".." && !filepath.IsAbs(rel) && (len(rel) < 2 || rel[:2] != "..")
}

// Usage: local files, err = tar.list(archive_path)
func luaList(L *lua.LState) int {
	archivePath := L.CheckString(1)

	file, err := os.Open(archivePath)
	if err != nil {
		return util.PushError(L, "failed to open archive: %v", err)
	}
	defer file.Close()

	reader := tar.NewReader(file)
	files := L.NewTable()
	idx := 1

	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return util.PushError(L, "failed to read archive: %v", err)
		}

		fileInfo := L.NewTable()
		fileInfo.RawSetString("name", lua.LString(header.Name))
		fileInfo.RawSetString("size", lua.LNumber(header.Size))
		fileInfo.RawSetString("mode", lua.LNumber(header.Mode))
		fileInfo.RawSetString("is_dir", lua.LBool(header.Typeflag == tar.TypeDir))

		files.RawSetInt(idx, fileInfo)
		idx++
	}

	return util.PushSuccess(L, files)
}

var exports = map[string]lua.LGFunction{
	"create":  luaCreate,
	"extract": luaExtract,
	"list":    luaList,
}

func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

func init() {
	modules.Register(ModuleName, Loader)
}
