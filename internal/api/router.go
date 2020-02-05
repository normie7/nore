package api

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/normie7/nore/internal/noiseremover"
)

var (
	ErrFileTooBig = errors.New("file is too big. upload file less than 10mb")
)

func NewRouter(redirectService noiseremover.Service, webDir string, globalSiteTag string) *mux.Router {
	var h handler
	h = newBaseHandler(redirectService, webDir, globalSiteTag)
	h = newTracingHandler(h)

	var sh *staticHandler
	sh = newStaticHandler(webDir)

	r := mux.NewRouter()
	// api
	r.HandleFunc("/upload", h.upload).Methods(http.MethodPost)
	r.HandleFunc("/download/{fileId}", h.download).Methods(http.MethodGet)
	// web
	r.HandleFunc("/status/{fileId}", h.status).Methods(http.MethodGet)
	r.HandleFunc("/", h.index).Methods(http.MethodGet)
	// static
	fs := http.FileServer(FileSystem{sh.assetsDir})
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs)).Methods(http.MethodGet)
	return r
}
