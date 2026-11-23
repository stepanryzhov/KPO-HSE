package services

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"path"

	"kpo_hw3_devenv/internal/dto"
)

var (
	ErrArchiveNotFound = errors.New("archive not found")
	ErrArchiveChanged  = errors.New("submission archive changed")
	ErrInvalidArchive  = errors.New("invalid ZIP archive")
)

type ArchiveRepository interface {
	GetArchive(ctx context.Context, submissionID, archiveKey string) ([]byte, error)
}

type Service struct {
	archives ArchiveRepository
}

func New(archives ArchiveRepository) *Service {
	return &Service{archives: archives}
}

func (s *Service) Analyze(ctx context.Context, submissionID, archiveKey string) (dto.BuildReport, error) {
	data, err := s.archives.GetArchive(ctx, submissionID, archiveKey)
	if err != nil {
		return dto.BuildReport{}, err
	}

	archive, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return dto.BuildReport{}, fmt.Errorf("%w: %w", ErrInvalidArchive, err)
	}

	for _, file := range archive.File {
		if file.Mode().IsRegular() && path.Base(file.Name) == "main.go" {
			return dto.BuildReport{Status: "passed", Success: true, Message: "В архиве найден main.go"}, nil
		}
	}
	return dto.BuildReport{Status: "failed", Success: false, Message: "В архиве нет main.go"}, nil
}
