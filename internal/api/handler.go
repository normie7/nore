package api

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/normie7/nore/internal/noiseremover"
)

type handler interface {
	index(w http.ResponseWriter, r *http.Request)
	status(w http.ResponseWriter, r *http.Request)
	upload(w http.ResponseWriter, r *http.Request)
	download(w http.ResponseWriter, r *http.Request)
}

type baseHandler struct {
	noiseRemover  noiseremover.Service
	templateDir   string
	globalSiteTag string
}

func newBaseHandler(s noiseremover.Service, webDir string, globalSiteTag string) handler {
	return &baseHandler{
		noiseRemover:  s,
		templateDir:   filepath.Join(webDir, "templates"),
		globalSiteTag: globalSiteTag,
	}
}

func (h *baseHandler) index(w http.ResponseWriter, r *http.Request) {
	lp := filepath.Join(h.templateDir, "layout.html")
	fp := filepath.Join(h.templateDir, "index.html")

	h.executeTemplate(w, "layout", nil, lp, fp)
}

func (h *baseHandler) executeTemplate(wr http.ResponseWriter, name string, data interface{}, filenames ...string) {
	type TemplateData struct {
		GST string
		D   interface{}
	}

	tmpl, err := template.ParseFiles(filenames...)
	if err != nil {
		log.Println("error parsing files:", err)
		http.Error(wr, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(
		wr,
		name,
		TemplateData{
			GST: h.globalSiteTag,
			D:   data,
		})
	if err != nil {
		log.Println("error executing template:", err)
		http.Error(wr, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
