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
	content, err := store.Get(_store, fileID)
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
	rw.WriteHeader(http.StatusOK)
	rw.Write(content)
}

func postHandler(_store, fileID string, rw http.ResponseWriter, request *http.Request) {
	if store.Exists(_store, fileID) {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(fileExists)
		return
	}
	err := request.ParseMultipartForm(maxFileSize)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(err.Error()))
		return
	}

	_, fileHeader, err := request.FormFile("file")
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(noFileInForm)
		return
	}

	fileName := fileHeader.Filename
	uploadedFile, err := fileHeader.Open()
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(err.Error()))
		return
	}
	defer uploadedFile.Close()

	buf := new(bytes.Buffer)

	n, err := buf.ReadFrom(uploadedFile)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(err.Error()))
		return
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
