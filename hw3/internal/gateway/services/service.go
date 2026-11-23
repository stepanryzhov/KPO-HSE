package services

import (
	"context"
	"errors"

	"kpo_hw3_devenv/internal/dto"
)

var (
	ErrInvalidArchive     = errors.New("invalid ZIP archive")
	ErrSubmissionNotFound = errors.New("submission not found")
	ErrArchiveChanged     = errors.New("submission archive changed")
	ErrStorageUnavailable = errors.New("storage request failed")
	ErrAnalyzeUnavailable = errors.New("analyze request failed")
)

type StorageRepository interface {
	SaveSubmission(ctx context.Context, assignmentID, studentID string, archive []byte) (dto.ArchiveRef, bool, error)
	GetSubmission(ctx context.Context, assignmentID, studentID string) (dto.SubmissionResponse, error)
	GetAssignmentSummary(ctx context.Context, assignmentID string) (dto.AssignmentSummary, error)
	SaveReport(ctx context.Context, ref dto.ArchiveRef, report dto.BuildReport) (dto.SubmissionResponse, error)
}

type AnalyzeRepository interface {
	Analyze(ctx context.Context, ref dto.ArchiveRef) (dto.BuildReport, error)
}

// SubmissionError содержит ID сдачи, сохранение которой подтвердил Storage.
type SubmissionError struct {
	SubmissionID string
	Err          error
}

func (e *SubmissionError) Error() string { return e.Err.Error() }
func (e *SubmissionError) Unwrap() error { return e.Err }

type Service struct {
	storage StorageRepository
	analyze AnalyzeRepository
}

func New(storage StorageRepository, analyze AnalyzeRepository) *Service {
	return &Service{storage: storage, analyze: analyze}
}

func (s *Service) Submit(ctx context.Context, assignmentID, studentID string, archive []byte) (dto.SubmissionResponse, bool, error) {
	ref, created, err := s.storage.SaveSubmission(ctx, assignmentID, studentID, archive)
	if err != nil {
		return dto.SubmissionResponse{}, false, err
	}

	report, err := s.analyze.Analyze(ctx, ref)
	if err != nil {
		return dto.SubmissionResponse{}, false, &SubmissionError{SubmissionID: ref.SubmissionID, Err: err}
	}
	submission, err := s.storage.SaveReport(ctx, ref, report)
	if err != nil {
		return dto.SubmissionResponse{}, false, &SubmissionError{SubmissionID: ref.SubmissionID, Err: err}
	}
	return submission, created, nil
}

func (s *Service) GetSubmission(ctx context.Context, assignmentID, studentID string) (dto.SubmissionResponse, error) {
	return s.storage.GetSubmission(ctx, assignmentID, studentID)
}

func (s *Service) GetAssignmentSummary(ctx context.Context, assignmentID string) (dto.AssignmentSummary, error) {
	return s.storage.GetAssignmentSummary(ctx, assignmentID)
}
