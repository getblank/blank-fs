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

	g.Describe("#list", func() {
		g.Before(func() {
			// Using memory stored backend for testing
			appFs = afero.NewMemMapFs()
		})

		g.It("should return empty response when store is empty", func() {
			store := "emptyListStore"
			res, err := List(store, 0, 0)
			g.Assert(err).Equal(nil)
			g.Assert(res == nil).IsFalse()
			g.Assert(len(res)).Equal(0)
		})
		g.It("should return list of Item when files presents", func() {
			store := "filledListStore"
			files := []Item{
				{
					ID:   "e669166e-18d3-46ae-9b35-617ea6d5a27a",
					Name: "1.zip",
				},
				{
					ID:   "c669166e-18d3-46ae-9b35-617ea6d5a27b",
					Name: "2.zip",
				},
				{
					ID:   "e769166e-18d3-46ae-9b35-617ea6d5a27c",
					Name: "3.zip",
				},
				{
					ID:   "d669166e-18d3-46ae-9b35-617ea6d5a27a",
					Name: "4.zip",
				},
				{
					ID:   "e669166e-18d3-46ae-9b35-617ea6d5a27b",
					Name: "5.zip",
				},
				{
					ID:   "f669166e-18d3-46ae-9b35-617ea6d5a27c",
					Name: "6.zip",
				},
			}

			for _, v := range files {
				err := File(store, v.ID, v.Name, []byte{})
				g.Assert(err).Equal(nil)
			}

			res, err := List(store, 0, 0)
			g.Assert(err).Equal(nil)
			g.Assert(len(res)).Equal(len(files))

			res, err = List(store, 2, 0)
			g.Assert(err).Equal(nil)
			g.Assert(len(res)).Equal(4)

			res, err = List(store, 2, 3)
			g.Assert(err).Equal(nil)
			g.Assert(len(res)).Equal(3)
			expected := []Item{
				{
					ID:    ".",
					Name:  "1.zip",
					Store: "filledListStore",
					Size:  0,
				},
				{
					ID:    ".",
					Name:  "5.zip",
					Store: "filledListStore",
					Size:  0,
				},
				{
					ID:    ".",
					Name:  "3.zip",
					Store: "filledListStore",
					Size:  0,
				},
			}

			for i, v := range expected {
				got := res[i]
				g.Assert(got.ID).Equal(v.ID)
				g.Assert(got.Name).Equal(v.Name)
				g.Assert(got.Store).Equal(v.Store)
				g.Assert(got.Size).Equal(v.Size)
			}
		})
	})
}
