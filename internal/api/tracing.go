package api

import (
	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/normie7/nore/internal/noiseremover"
)

type tracingHandler struct {
	handler
}

func newTracingHandler(handler handler) handler {
	return &tracingHandler{handler: handler}
}

func (h *tracingHandler) index(w http.ResponseWriter, r *http.Request) {
	h.handler.index(w, r.WithContext(context.WithValue(r.Context(), noiseremover.LogToken, h.logToken())))
}

func (h *tracingHandler) status(w http.ResponseWriter, r *http.Request) {
	t := h.logToken()
	log.Println(t, "status endpoint hit", mux.Vars(r))
	h.handler.status(w, r.WithContext(context.WithValue(r.Context(), noiseremover.LogToken, t)))
}

func (h *tracingHandler) upload(w http.ResponseWriter, r *http.Request) {
	t := h.logToken()
	log.Println(t, "upload endpoint hit")
	h.handler.upload(w, r.WithContext(context.WithValue(r.Context(), noiseremover.LogToken, t)))
}

func (h *tracingHandler) download(w http.ResponseWriter, r *http.Request) {
	t := h.logToken()
	log.Println(t, "download endpoint hit")
	h.handler.download(w, r.WithContext(context.WithValue(r.Context(), noiseremover.LogToken, t)))
}

func (h *tracingHandler) logToken() string {
	u, err := uuid.NewRandom()
	if err != nil {
		return "error_generating_random_token"
	}
	return u.String()
}
