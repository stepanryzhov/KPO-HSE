package transport

import "net/http"

func (h *Handler) SetRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /internal/assignments/{assignment_id}/submissions", h.saveSubmission)
	mux.HandleFunc("GET /internal/assignments/{assignment_id}/submissions", h.getSubmission)
	mux.HandleFunc("GET /internal/assignments/{assignment_id}/summary", h.getAssignmentSummary)
	mux.HandleFunc("GET /internal/submissions/{submission_id}/archives/{archive_key}", h.getArchive)
	mux.HandleFunc("PUT /internal/submissions/{submission_id}/archives/{archive_key}/report", h.saveReport)
}
