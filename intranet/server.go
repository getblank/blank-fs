package intranet

import (
	"bytes"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/getblank/blank-filestore/store"
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
	uri := strings.Trim(request.RequestURI, "/")
	splitted := strings.Split(uri, "/")
	if len(splitted) != 2 {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(invalidParams)
		return
	}
	switch request.Method {
	case http.MethodPost:
		postHandler(splitted[0], splitted[1], rw, request)
	case http.MethodGet:
		getHandler(splitted[0], splitted[1], rw)
	case http.MethodDelete:
		deleteHandler(splitted[0], splitted[1], rw)
	default:
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(requestMethodsRestrictions)
	}
}

func getHandler(_store, fileID string, rw http.ResponseWriter) {
	log.Debugf("New GET request. Store: %s, fileID: %s", _store, fileID)
	fileName, content, err := store.Get(_store, fileID)
	if err != nil {
		if err == store.ErrNotFound {
			rw.WriteHeader(http.StatusNotFound)
			rw.Write(fileNotFound)
		} else {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(err.Error()))
		}
		return
	}
	rw.Header().Set("file-name", fileName)
	rw.WriteHeader(http.StatusOK)
	log.Debugf("Send %s for GET request. Store: %s, fileID: %s", fileName, _store, fileID)
	rw.Write(content)
}

func postHandler(_store, fileID string, rw http.ResponseWriter, request *http.Request) {
	log.Debugf("New POST request. Store: %s, fileID: %s", _store, fileID)
	if store.Exists(_store, fileID) {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(fileExists)
		return
	}
	fileName := request.Header.Get("file-name")
	if fileName == "" {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(noFileName)
		return
	}

	buf := bytes.NewBuffer(nil)
	n, err := buf.ReadFrom(request.Body)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(err.Error()))
	}
	if n == 0 {
		log.WithField("filename", fileName).Warn("Uploaded file is empty")
	}

	err = store.File(_store, fileID, fileName, buf.Bytes())
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(err.Error()))
		return
	}
	rw.Write(fileStored)
}

func deleteHandler(_store, fileID string, rw http.ResponseWriter) {
	log.Debugf("New DELETE request. Store: %s, fileID: %s", _store, fileID)
	err := store.Del(_store, fileID)
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
