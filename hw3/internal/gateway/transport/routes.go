package transport

import "net/http"

func (h *Handler) SetRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /assignments/{assignment_id}", h.submit)
	mux.HandleFunc("GET /assignments/{assignment_id}", h.getAssignment)
}
