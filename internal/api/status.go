package api

import (
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
)

func (h *baseHandler) status(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	// todo sse
	s, _ := h.noiseRemover.Status(r.Context(), vars["fileId"])

	lp := filepath.Join(h.templateDir, "layout.html")
	fp := filepath.Join(h.templateDir, "status.html")

	h.executeTemplate(w, "layout", s, lp, fp)
}
