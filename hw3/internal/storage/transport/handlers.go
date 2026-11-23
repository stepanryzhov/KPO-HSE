package transport

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"

	"kpo_hw3_devenv/internal/dto"
	"kpo_hw3_devenv/internal/storage/services"
)

type StorageService interface {
	SaveSubmission(ctx context.Context, assignmentID, studentID string, src io.Reader) (services.Submission, bool, error)
	GetSubmission(ctx context.Context, assignmentID, studentID string) (services.Submission, error)
	GetAssignmentSummary(ctx context.Context, assignmentID string) (services.AssignmentSummary, error)
	OpenArchive(ctx context.Context, submissionID, archiveKey string) (io.ReadCloser, error)
	SaveReport(ctx context.Context, submissionID, archiveKey string, report services.BuildReport) (services.Submission, error)
}

type Handler struct {
	service StorageService
}

func NewHandler(service StorageService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) saveSubmission(w http.ResponseWriter, r *http.Request) {
	studentID := r.URL.Query().Get("student_id")
	if studentID == "" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{Message: "Параметр student_id обязателен"})
		return
	}
	if !hasContentType(r, "application/zip") {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{Message: "Ожидается Content-Type application/zip"})
		return
	}

	submission, created, err := h.service.SaveSubmission(r.Context(), r.PathValue("assignment_id"), studentID, r.Body)
	if err != nil {
		writeServiceError(w, r, err, "")
		return
	}
	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	writeJSON(w, status, dto.ArchiveRef{SubmissionID: submission.SubmissionID, ArchiveKey: submission.ArchiveKey})
}

func (h *Handler) getSubmission(w http.ResponseWriter, r *http.Request) {
	studentID := r.URL.Query().Get("student_id")
	if studentID == "" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{Message: "Параметр student_id обязателен"})
		return
	}
	submission, err := h.service.GetSubmission(r.Context(), r.PathValue("assignment_id"), studentID)
	if err != nil {
		writeServiceError(w, r, err, "")
		return
	}
	writeJSON(w, http.StatusOK, submissionResponse(submission))
}

func (h *Handler) getAssignmentSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.service.GetAssignmentSummary(r.Context(), r.PathValue("assignment_id"))
	if err != nil {
		writeServiceError(w, r, err, "")
		return
	}
	submissions := make([]dto.SubmissionResponse, len(summary.Submissions))
	for i, submission := range summary.Submissions {
		submissions[i] = submissionResponse(submission)
	}
	writeJSON(w, http.StatusOK, dto.AssignmentSummary{
		Submissions: submissions,
		Counts:      dto.SummaryCounts(summary.Counts),
	})
}

func (h *Handler) getArchive(w http.ResponseWriter, r *http.Request) {
	submissionID := r.PathValue("submission_id")
	archive, err := h.service.OpenArchive(r.Context(), submissionID, r.PathValue("archive_key"))
	if err != nil {
		writeServiceError(w, r, err, submissionID)
		return
	}
	defer archive.Close()

	w.Header().Set("Content-Type", "application/zip")
	if _, err := io.Copy(w, archive); err != nil {
		slog.WarnContext(r.Context(), "send archive", "submission_id", submissionID, "error", err)
	}
}

func (h *Handler) saveReport(w http.ResponseWriter, r *http.Request) {
	if !hasContentType(r, "application/json") {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{Message: "Ожидается Content-Type application/json"})
		return
	}
	var report dto.BuildReport
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&report); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{Message: "Некорректный JSON отчёта"})
		return
	}
	if err := decoder.Decode(new(json.RawMessage)); !errors.Is(err, io.EOF) {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{Message: "Ожидается один JSON-объект отчёта"})
		return
	}
	if (report.Status != "passed" && report.Status != "failed") || report.Success != (report.Status == "passed") {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{Message: "Некорректный статус отчёта"})
		return
	}

	submissionID := r.PathValue("submission_id")
	submission, err := h.service.SaveReport(r.Context(), submissionID, r.PathValue("archive_key"), services.BuildReport(report))
	if err != nil {
		writeServiceError(w, r, err, submissionID)
		return
	}
	writeJSON(w, http.StatusOK, submissionResponse(submission))
}

func submissionResponse(submission services.Submission) dto.SubmissionResponse {
	response := dto.SubmissionResponse{
		SubmissionID: submission.SubmissionID,
		AssignmentID: submission.AssignmentID,
		StudentID:    submission.StudentID,
		SubmittedAt:  submission.SubmittedAt,
	}
	if submission.Report != nil {
		report := dto.BuildReport(*submission.Report)
		response.Report = &report
	}
	return response
}

func hasContentType(r *http.Request, expected string) bool {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	return err == nil && mediaType == expected
}

func writeServiceError(w http.ResponseWriter, r *http.Request, err error, submissionID string) {
	switch {
	case errors.Is(err, services.ErrInvalidArchive):
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{Message: "Некорректный ZIP-архив"})
	case errors.Is(err, services.ErrSubmissionNotFound):
		writeJSON(w, http.StatusNotFound, dto.ErrorResponse{Message: "Решение не найдено"})
	case errors.Is(err, services.ErrArchiveChanged):
		writeJSON(w, http.StatusConflict, dto.ErrorResponse{Message: "Сдача уже обновлена", SubmissionID: submissionID})
	default:
		slog.ErrorContext(r.Context(), "storage request failed", "method", r.Method, "path", r.URL.Path, "error", err)
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{Message: "Ошибка хранилища"})
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Warn("send storage JSON response", "error", err)
	}
}
