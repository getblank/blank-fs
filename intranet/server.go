package intranet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/getblank/blank-fs/store"
)

const (
	maxFileSize = 1024 * 1024 * 100 // 100 MB
)

var (
	noFileInForm               = []byte("no file in form")
	fileExists                 = []byte("file exists")
	fileStored                 = []byte("file stored")
	fileNotFound               = []byte("file not found")
	invalidParams              = []byte("invalid params")
	requestMethodsRestrictions = []byte("Only GET and POST request is allowed")
	invalidTakeSkipParams      = []byte("invalid take or skip param")
	fileDeleted                = []byte("deleted")
	noFileName                 = []byte("no file-name header")
)

func startHTTPServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", httpHandler)

	log.Infof("Start listening http requests at port %s", httpPort)
	err := http.ListenAndServe(":"+httpPort, mux)
	if err != nil {
		panic(err)
	}
}

func httpHandler(rw http.ResponseWriter, request *http.Request) {
	uri := strings.Trim(request.URL.Path, "/")
	splitted := strings.Split(uri, "/")
	var storeName, fileID string
	if len(splitted) >= 2 {
		storeName = splitted[0]
		fileID = splitted[1]
	} else {
		if len(splitted) == 0 {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(invalidParams)
			return
		}

		storeName = splitted[0]
	}

	switch request.Method {
	case http.MethodPost:
		postHandler(storeName, fileID, rw, request)
	case http.MethodGet:
		if len(fileID) == 0 {
			listHandler(storeName, rw, request)
			return
		}
		getHandler(storeName, fileID, rw)
	case http.MethodDelete:
		deleteHandler(storeName, fileID, rw)
	default:
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(requestMethodsRestrictions)
	}
}

func listHandler(storeName string, rw http.ResponseWriter, request *http.Request) {
	log.Debugf("New LIST request. Store: %s", storeName)
	var skip, take int
	query := request.URL.Query()
	var err error
	if t := query.Get("skip"); len(t) > 0 {
		skip, err = strconv.Atoi(t)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(invalidTakeSkipParams)
			return
		}

		log.Debugf("LIST request for store: %s skip param: %d", storeName, skip)
	}

	if t := query.Get("take"); len(t) > 0 {
		take, err = strconv.Atoi(t)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(invalidTakeSkipParams)
			return
		}

		log.Debugf("LIST request for store: %s take param: %d", storeName, take)
	}

	list, err := store.List(storeName, skip, take)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(err.Error()))
		return
	}

	encoded, err := json.Marshal(list)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(err.Error()))
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	rw.Write(encoded)
}

func getHandler(storeName, fileID string, rw http.ResponseWriter) {
	log.Debugf("New GET request. Store: %s, fileID: %s", storeName, fileID)
	fileName, content, err := store.Get(storeName, fileID)
	if err != nil {
		if err == store.ErrNotFound {
			rw.WriteHeader(http.StatusNotFound)
			rw.Write(fileNotFound)

			return
		}

		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(err.Error()))

		return
	}

	rw.Header().Set("File-Name", fileName) // TODO: remove this header after updating router and worker
	rw.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	rw.Header().Set("Content-Type", detectContentType(fileName, content))
	rw.Header().Set("Content-Length", strconv.Itoa(len(content)))
	rw.WriteHeader(http.StatusOK)
	log.Debugf("Send %s for GET request. Store: %s, fileID: %s", fileName, storeName, fileID)
	rw.Write(content)
}

func postHandler(storeName, fileID string, rw http.ResponseWriter, request *http.Request) {
	log.Debugf("New POST request. Store: %s, fileID: %s", storeName, fileID)
	if store.Exists(storeName, fileID) {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(fileExists)
		return
	}

	var fileName string
	if _, params, err := mime.ParseMediaType(request.Header.Get("Content-Disposition")); err == nil && len(params["filename"]) > 0 {
		fileName = params["filename"]
	} else if fileName = request.Header.Get("File-Name"); len(fileName) == 0 {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(noFileName)
		return
	}

	buf := bytes.NewBuffer(nil)
	n, err := buf.ReadFrom(request.Body)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(err.Error()))
		return
	}

	if n == 0 {
		log.WithField("filename", fileName).Warn("Uploaded file is empty")
	}

	err = store.File(storeName, fileID, fileName, buf.Bytes())
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(err.Error()))
		return
	}

	rw.Write(fileStored)
}

func deleteHandler(storeName, fileID string, rw http.ResponseWriter) {
	log.Debugf("New DELETE request. Store: %s, fileID: %s", storeName, fileID)
	err := store.Del(storeName, fileID)
	if err != nil {
		if err == store.ErrNotFound {
			rw.WriteHeader(http.StatusNotFound)
			rw.Write(fileNotFound)
			return
		}

		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(err.Error()))
		return
	}

	rw.Write(fileDeleted)
}

func detectContentType(fileName string, content []byte) string {
	ctype := mime.TypeByExtension(filepath.Ext(fileName))
	if ctype == "" {
		ctype = http.DetectContentType(content)
	}

	return ctype
}
