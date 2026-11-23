package transport

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"net/url"

	"kpo_hw3_devenv/internal/dto"
	"kpo_hw3_devenv/internal/gateway/services"
)

type GatewayService interface {
	Submit(ctx context.Context, assignmentID, studentID string, archive []byte) (dto.SubmissionResponse, bool, error)
	GetSubmission(ctx context.Context, assignmentID, studentID string) (dto.SubmissionResponse, error)
	GetAssignmentSummary(ctx context.Context, assignmentID string) (dto.AssignmentSummary, error)
}

type Handler struct {
	service     GatewayService
	maxZIPBytes int64
}

func NewHandler(service GatewayService, maxZIPBytes int64) *Handler {
	return &Handler{service: service, maxZIPBytes: maxZIPBytes}
}

func (h *Handler) submit(w http.ResponseWriter, r *http.Request) {
	query, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		writeBadRequest(w, false, "Некорректные параметры запроса")
		return
	}
	studentID := query.Get("student_id")
	if studentID == "" {
		writeBadRequest(w, false, "Параметр student_id обязателен")
		return
	}
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/zip" {
		writeBadRequest(w, false, "Ожидается Content-Type application/zip")
		return
	}
	archive, err := io.ReadAll(http.MaxBytesReader(w, r.Body, h.maxZIPBytes))
	if err != nil {
		message := "Не удалось прочитать тело запроса"
		var limitErr *http.MaxBytesError
		if errors.As(err, &limitErr) {
			message = "Размер ZIP-архива превышает допустимый лимит"
		}
		writeBadRequest(w, false, message)
		return
	}
	submission, created, err := h.service.Submit(r.Context(), r.PathValue("assignment_id"), studentID, archive)
	if err != nil {
		writeServiceError(w, r, false, err)
		return
	}
	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	writeJSON(w, status, submission)
}

func (h *Handler) getAssignment(w http.ResponseWriter, r *http.Request) {
	query, err := url.ParseQuery(r.URL.RawQuery)
	format := query.Get("format")
	asHTML := format == "html"
	if query.Has("format") && format != "json" && !asHTML {
		writeBadRequest(w, false, "Допустимые значения format — json и html")
		return
	}
	if err != nil {
		writeBadRequest(w, asHTML, "Некорректные параметры запроса")
		return
	}
	if !query.Has("student_id") {
		h.getSummary(w, r, asHTML)
		return
	}
	studentID := query.Get("student_id")
	if studentID == "" {
		writeBadRequest(w, asHTML, "Параметр student_id не может быть пустым")
		return
	}
	submission, err := h.service.GetSubmission(r.Context(), r.PathValue("assignment_id"), studentID)
	if err != nil {
		writeServiceError(w, r, asHTML, err)
		return
	}
	if asHTML {
		writeHTML(w, http.StatusOK, "submission.html", submission)
		return
	}
	writeJSON(w, http.StatusOK, submission)
}

func (h *Handler) getSummary(w http.ResponseWriter, r *http.Request, asHTML bool) {
	assignmentID := r.PathValue("assignment_id")
	summary, err := h.service.GetAssignmentSummary(r.Context(), assignmentID)
	if err != nil {
		writeServiceError(w, r, asHTML, err)
		return
	}
	if asHTML {
		writeHTML(w, http.StatusOK, "summary.html", summaryPage{AssignmentID: assignmentID, AssignmentSummary: summary})
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func writeServiceError(w http.ResponseWriter, r *http.Request, asHTML bool, err error) {
	status := http.StatusServiceUnavailable
	message := "Не удалось выполнить межсервисный запрос"
	switch {
	case errors.Is(err, services.ErrInvalidArchive):
		status, message = http.StatusBadRequest, "Некорректный ZIP-архив"
	case errors.Is(err, services.ErrSubmissionNotFound):
		status, message = http.StatusNotFound, "Решение не найдено"
	case errors.Is(err, services.ErrArchiveChanged):
		status, message = http.StatusConflict, "Сдача уже обновлена"
	case errors.Is(err, services.ErrStorageUnavailable):
		message = dependencyErrorMessage("Storage", err)
	case errors.Is(err, services.ErrAnalyzeUnavailable):
		message = dependencyErrorMessage("Analyze", err)
	}
	if status == http.StatusServiceUnavailable {
		slog.WarnContext(r.Context(), "gateway upstream request failed", "method", r.Method, "path", r.URL.Path, "error", err)
	}
	response := dto.ErrorResponse{Message: message}
	var submissionErr *services.SubmissionError
	if errors.As(err, &submissionErr) {
		response.SubmissionID = submissionErr.SubmissionID
	}
	writeError(w, status, asHTML, response)
}

func dependencyErrorMessage(service string, err error) string {
	if errors.Is(err, context.DeadlineExceeded) {
		return service + " не ответил вовремя"
	}
	return "Не удалось выполнить запрос к " + service
}

func writeBadRequest(w http.ResponseWriter, asHTML bool, message string) {
	writeError(w, http.StatusBadRequest, asHTML, dto.ErrorResponse{Message: message})
}

func writeError(w http.ResponseWriter, status int, asHTML bool, response dto.ErrorResponse) {
	if asHTML {
		writeHTML(w, status, "error.html", response)
		return
	}
	writeJSON(w, status, response)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Warn("send gateway JSON response", "error", err)
	}
}
