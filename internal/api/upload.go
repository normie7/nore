package api

import (
	"errors"
	"log"
	"net/http"

	"github.com/normie7/nore/internal/noiseremover"
)

func (h *baseHandler) upload(w http.ResponseWriter, r *http.Request) {
	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 10 MB files.
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20+512)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		log.Println(http.StatusBadRequest, r.Context().Value(noiseremover.LogToken), "error while uploading (", err, "), sending - ", ErrFileTooBig.Error())
		http.Error(w, ErrFileTooBig.Error(), http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("audioFile")
	if err != nil {
		log.Println(http.StatusBadRequest, r.Context().Value(noiseremover.LogToken), "error retrieving the file:", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	defer file.Close()

	f, err := h.noiseRemover.Store(r.Context(), file, handler)
	if err != nil {
		s := http.StatusInternalServerError
		if errors.Is(err, noiseremover.ErrWrongFileType) {
			s = http.StatusBadRequest
		}
		log.Println(s, r.Context().Value(noiseremover.LogToken), "error saving the file:", err)
		http.Error(w, http.StatusText(s), s)
		return
	}

	// todo change link
	w.Write([]byte("<script>window.location.href = \"/status/" + f.Id + "\";</script>"))
}
