package store

import (
	"bytes"
	"errors"
	"mime/multipart"
	"os"
	"path/filepath"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/afero"
)

var (
	AppFs afero.Fs
	// ErrAccessDenied uses when access denied
	ErrAccessDenied = errors.New("Access denied!")
	ErrServerError  = errors.New("Server error")
	ErrFileExists   = errors.New("File exists")
	tmpFiles        = map[string]*tmpFile{}
	tmpFilesMutex   = &sync.Mutex{}
	tmpFileTTL      = time.Hour * 2
)

type tmpFile struct {
	ID         string     `json:"_id"`
	Name       string     `json:"name"`
	Store      string     `json:"store"`
	Size       int        `json:"size"`
	UploadedAt *time.Time `json:"uploadedAt"`
}

func init() {
	AppFs = afero.NewOsFs()

}

func StoreFile(_store string, _file *multipart.FileHeader) error {
	fileID := "1"
	filePath := getFilePath(_store, fileID)
	if filePath == "" {
		return ErrServerError
	}

	uploadedFile, err := _file.Open()
	if err != nil {
		return err
	}
	defer uploadedFile.Close()

	_, err = os.Stat(filePath)
	if err == nil || os.IsExist(err) {
		return ErrFileExists
	}

	buf := new(bytes.Buffer)

	n, err := buf.ReadFrom(uploadedFile)
	if err != nil {
		return err
	}
	if n == 0 {
		log.WithField("filename", _file.Filename).Warn("Uploaded file is empty")
	}
	err = os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
	if err != nil {
		return err
	}
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	nn, err := file.Write(buf.Bytes())
	if err != nil {
		return err
	}
	if nn == 0 {
		return errors.New("Can't write file")
	}
	return nil
}

func getFilePath(_ string, _ string) string {
	return ""
}

func fileTerminator() {
	ch := time.Tick(time.Minute)
	for {
		<-ch
		// now := time.Now()
		// tmpFilesMutex.Lock()
		// for i, tfile := range tmpFiles {
		// 	if now.Sub(*tfile.UploadedAt) > tmpFileTTL {
		// 		filePath := getFilePath(tfile.Store, tfile.Id)
		// 		err := os.Remove(filePath)
		// 		if err != nil {
		// 			if os.IsExist(err) {
		// 				logger.Error("Can't remove temp file", err.Error())
		// 				continue
		// 			}
		// 		}
		// 		delete(tmpFiles, i)
		// 		// db.Delete(config.TempFileStoreBucket, tfile.Id)
		// 	}
		// }
		// tmpFilesMutex.Unlock()
	}
}
