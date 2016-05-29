package store

import (
	"path/filepath"
	"strings"
	"testing"

	. "github.com/franela/goblin"
	"github.com/spf13/afero"
)

func TestFileStore(t *testing.T) {
	g := Goblin(t)
	g.Describe("#getFilePath", func() {
		g.Before(func() {
			// Using memory stored backend for testing
			appFs = afero.NewMemMapFs()
		})
		store := "fileStore"
		fileID := "c669166e-18d3-46ae-9b35-617ea6d5a27a"
		fileName := "img.gif"
		fileContent := "Blank is the the next generation of web applications"
		g.It("should return not empty file path if fileName provided", func() {
			path, err := getFilePath(store, fileName, fileID)
			g.Assert(err == nil).IsTrue()
			g.Assert(path != "").IsTrue()
		})
		g.It("should return error and empty path if length of filID < 3", func() {
			path, err := getFilePath(store, fileName, "aa")
			g.Assert(err == nil).IsFalse()
			g.Assert(path == "").IsTrue()
		})
		g.It("should return error and empty path if file not found", func() {
			path, err := getFilePath(store, "", "aaaaaaaaaaa")
			g.Assert(err != nil).IsTrue()
			g.Assert(path == "").IsTrue()
		})
		g.It("should return full path if fileName is not provided and file was stored", func() {
			path, err := getFilePath(store, fileName, fileID)
			err = appFs.MkdirAll(filepath.Dir(path), 0644)
			g.Assert(err == nil).IsTrue()
			file, err := appFs.Create(path)
			g.Assert(err == nil).IsTrue()
			_, err = file.WriteString(fileContent)
			g.Assert(err == nil).IsTrue()
			file.Close()

			storedPath, err := getFilePath(store, "", fileID)
			g.Assert(err == nil).IsTrue()
			g.Assert(strings.HasSuffix(storedPath, fileName)).IsTrue()
		})
	})
}
