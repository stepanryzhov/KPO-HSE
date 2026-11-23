package repository

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"kpo_hw3_devenv/internal/dto"
	"kpo_hw3_devenv/internal/gateway/services"
)

type Storage struct {
	client client
}

func NewStorage(baseURL string, httpClient *http.Client) *Storage {
	return &Storage{client: newClient(baseURL, httpClient, services.ErrStorageUnavailable)}
}

func (s *Storage) SaveSubmission(ctx context.Context, assignmentID, studentID string, archive []byte) (dto.ArchiveRef, bool, error) {
	var ref dto.ArchiveRef
	status, err := s.client.request(ctx, http.MethodPost, studentSubmissionPath(assignmentID, studentID),
		"application/zip", bytes.NewReader(archive), &ref, http.StatusCreated, http.StatusOK)
	if status == http.StatusBadRequest {
		return dto.ArchiveRef{}, false, services.ErrInvalidArchive
	}
	if err != nil {
		return dto.ArchiveRef{}, false, err
	}
	if ref.SubmissionID == "" || ref.ArchiveKey == "" {
		return dto.ArchiveRef{}, false, fmt.Errorf("%w: missing archive reference", services.ErrStorageUnavailable)
	}
	return ref, status == http.StatusCreated, nil
}

func (s *Storage) GetSubmission(ctx context.Context, assignmentID, studentID string) (dto.SubmissionResponse, error) {
	var submission dto.SubmissionResponse
	status, err := s.client.request(ctx, http.MethodGet, studentSubmissionPath(assignmentID, studentID), "", nil, &submission, http.StatusOK)
	if status == http.StatusNotFound {
		return dto.SubmissionResponse{}, services.ErrSubmissionNotFound
	}
	if err == nil && submission.SubmissionID == "" {
		return dto.SubmissionResponse{}, fmt.Errorf("%w: missing submission ID", services.ErrStorageUnavailable)
	}
	return submission, err
}

func (s *Storage) GetAssignmentSummary(ctx context.Context, assignmentID string) (dto.AssignmentSummary, error) {
	var summary dto.AssignmentSummary
	path := assignmentPath(assignmentID) + "/summary"
	_, err := s.client.request(ctx, http.MethodGet, path, "", nil, &summary, http.StatusOK)
	if err == nil && summary.Submissions == nil {
		return dto.AssignmentSummary{}, fmt.Errorf("%w: missing submissions in summary", services.ErrStorageUnavailable)
	}
	return summary, err
}

func (s *Storage) SaveReport(ctx context.Context, ref dto.ArchiveRef, report dto.BuildReport) (dto.SubmissionResponse, error) {
	var submission dto.SubmissionResponse
	path := "/internal/submissions/" + url.PathEscape(ref.SubmissionID) + "/archives/" + url.PathEscape(ref.ArchiveKey) + "/report"
	status, err := s.client.sendJSON(ctx, http.MethodPut, path, report, &submission)
	if status == http.StatusConflict {
		return dto.SubmissionResponse{}, services.ErrArchiveChanged
	}
	if err == nil && (submission.SubmissionID != ref.SubmissionID || submission.Report == nil) {
		return dto.SubmissionResponse{}, fmt.Errorf("%w: missing saved report for %s", services.ErrStorageUnavailable, ref.SubmissionID)
	}
	return submission, err
}

func assignmentPath(assignmentID string) string {
	segment := url.PathEscape(assignmentID)
	// ServeMux очищает буквальные . и .. в пути.
	if segment == "." || segment == ".." {
		segment = strings.ReplaceAll(segment, ".", "%2E")
	}
	return "/internal/assignments/" + segment
}

func studentSubmissionPath(assignmentID, studentID string) string {
	query := url.Values{"student_id": {studentID}}
	return assignmentPath(assignmentID) + "/submissions?" + query.Encode()
}
