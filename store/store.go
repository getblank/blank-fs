package store

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"io"

	"github.com/spf13/afero"
)

var (
	appFs         afero.Fs
	rootDir       = "files"
	tmpFiles      = map[string]*Item{}
	tmpFilesMutex = &sync.Mutex{}
	tmpFileTTL    = time.Hour * 2
)

// Errors
var (
	ErrAccessDenied = errors.New("Access denied!")
	ErrServerError  = errors.New("Server error")
	ErrFileExists   = errors.New("File exists")
	ErrNotFound     = errors.New("File not found")
)

type Item struct {
	ID         string     `json:"_id"`
	Name       string     `json:"name"`
	Store      string     `json:"store"`
	Size       int        `json:"size,omitempty"`
	UploadedAt *time.Time `json:"uploadedAt,omitempty"`
}

func init() {
	appFs = afero.NewOsFs()
}

// Del removes file from fs
func Del(store, id string) error {
	path, err := getFilePath(store, "", id)
	if err != nil {
		return err
	}

	err = appFs.Remove(path)
	if err != nil {
		if os.IsNotExist(err) {
			err = ErrNotFound
		}
	} else {
		appFs.Remove(filepath.Dir(path))
	}
	return err
}

// Exists returns true if file already stored
func Exists(store, id string) bool {
	filePath, err := getFilePath(store, "", id)
	if err != nil {
		return false
	}
	_, err = appFs.Stat(filePath)
	if err != nil {
		// may be server doesn't have permissions to read
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

// File stores file to fs
func File(_store, fileID, fileName string, _file []byte) error {
	path, err := getFilePath(_store, fileName, fileID)
	if err != nil {
		return err
	}

	return saveFile(path, _file)
}

// Get returns fileName and content from appFs or error
func Get(store, id string) (string, []byte, error) {
	path, err := getFilePath(store, "", id)
	if err != nil {
		return "", nil, err
	}

	content, err := afero.ReadFile(appFs, path)
	if err != nil && os.IsNotExist(err) {
		err = ErrNotFound
	}
	return filepath.Base(path), content, err
}

// List returs list of files in store. Result can by limited by skip and take params
func List(store string, skip, take int) ([]*Item, error) {
	fileDir := rootDir + "/" + store + "/"
	res := []*Item{}
	dirExists, err := afero.DirExists(appFs, fileDir)
	if err != nil {
		return res, err
	}
	if !dirExists {
		return res, nil
	}

	cur := 0
	err = afero.Walk(appFs, fileDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			cur++
			if skip >= cur {
				return nil
			}

			if take > 0 && len(res) == take {
				return io.EOF
			}

			item := Item{
				ID:    filepath.Base(filepath.Dir(path)),
				Name:  info.Name(),
				Store: store,
				Size:  int(info.Size()),
			}
			res = append(res, &item)
		}
		return nil
	})

	if err == io.EOF {
		err = nil
	}

	return res, err
}

func saveFile(path string, content []byte) error {
	err := appFs.MkdirAll(filepath.Dir(path), 0744)
	if err != nil {
		return err
	}
	file, err := appFs.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(content)
	return err
}

func getFilePath(store, name, id string) (string, error) {
	if len(id) < 3 {
		return "", ErrNotFound
	}
	fileDir := rootDir + "/" + store + "/" + id[:2] + "/" + id
	if name != "" {
		return fileDir + "/" + name, nil
	}
	files, err := afero.ReadDir(appFs, fileDir)
	if err != nil {
		if os.IsNotExist(err) {
			err = ErrNotFound
		}
		return "", err
	}
	if len(files) == 0 {
		return "", ErrNotFound
	}

	return fileDir + "/" + files[0].Name(), nil
}
