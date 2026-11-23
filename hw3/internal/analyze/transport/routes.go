package transport

import "net/http"

func (h *Handler) SetRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /internal/analyze", h.analyze)
}
