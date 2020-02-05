package api

import (
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"github.com/normie7/nore/internal/noiseremover"
)

func (h *baseHandler) download(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if dd, err := h.noiseRemover.Find(r.Context(), vars["fileId"]); err == nil {
		http.ServeFile(w, r, dd.FullPath)
	} else {
		s := http.StatusInternalServerError
		switch {
		case os.IsNotExist(err), errors.Is(err, noiseremover.ErrWrongTokenFormat), errors.Is(err, noiseremover.ErrFileNotFound):
			s = http.StatusNotFound
		}
		log.Println(s, r.Context().Value(noiseremover.LogToken), "error downloading the file:", vars["fileId"], err)
		http.Error(w, http.StatusText(s), s)
	}
}
