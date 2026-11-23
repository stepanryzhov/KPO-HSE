package transport

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"

	"kpo_hw3_devenv/internal/analyze/services"
	"kpo_hw3_devenv/internal/dto"
)

type AnalyzeService interface {
	Analyze(ctx context.Context, submissionID, archiveKey string) (dto.BuildReport, error)
}

type Handler struct {
	service AnalyzeService
}

func NewHandler(service AnalyzeService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) analyze(w http.ResponseWriter, r *http.Request) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{Message: "Ожидается Content-Type application/json"})
		return
	}

	var ref dto.ArchiveRef
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&ref); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{Message: "Некорректный JSON запроса"})
		return
	}
	if err := decoder.Decode(new(json.RawMessage)); !errors.Is(err, io.EOF) {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{Message: "Ожидается один JSON-объект запроса"})
		return
	}
	if ref.SubmissionID == "" || ref.ArchiveKey == "" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{Message: "Поля submission_id и archive_key обязательны"})
		return
	}

	report, err := h.service.Analyze(r.Context(), ref.SubmissionID, ref.ArchiveKey)
	if err != nil {
		writeServiceError(w, r, err, ref.SubmissionID)
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func writeServiceError(w http.ResponseWriter, r *http.Request, err error, submissionID string) {
	switch {
	case errors.Is(err, services.ErrArchiveNotFound):
		writeJSON(w, http.StatusNotFound, dto.ErrorResponse{Message: "Решение не найдено"})
	case errors.Is(err, services.ErrArchiveChanged):
		writeJSON(w, http.StatusConflict, dto.ErrorResponse{Message: "Сдача уже обновлена", SubmissionID: submissionID})
	case errors.Is(err, services.ErrInvalidArchive):
		slog.ErrorContext(r.Context(), "read submission ZIP", "submission_id", submissionID, "error", err)
		writeJSON(w, http.StatusBadGateway, dto.ErrorResponse{Message: "Storage вернул некорректный ZIP-архив", SubmissionID: submissionID})
	default:
		slog.ErrorContext(r.Context(), "analyze request failed", "submission_id", submissionID, "error", err)
		writeJSON(w, http.StatusServiceUnavailable, dto.ErrorResponse{Message: "Не удалось получить архив из Storage", SubmissionID: submissionID})
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Warn("send analyze JSON response", "error", err)
	}
}
