package services

import (
	"context"
	"crypto/rand"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"time"
)

var (
	ErrSubmissionNotFound = errors.New("submission not found")
	ErrArchiveChanged     = errors.New("submission archive changed")
	ErrInvalidArchive     = errors.New("invalid ZIP archive")
)

// Submission описывает текущую сдачу студента по заданию.
type Submission struct {
	SubmissionID string
	AssignmentID string
	StudentID    string
	SubmittedAt  time.Time
	ArchiveKey   string
	Report       *BuildReport
}

// BuildReport содержит результат проверки одной загрузки.
type BuildReport struct {
	Status  string
	Success bool
	Message string
}

// AssignmentSummary содержит текущие сдачи по заданию и их счётчики.
type AssignmentSummary struct {
	Submissions []Submission
	Counts      SummaryCounts
}

type SummaryCounts struct {
	Total   int
	Passed  int
	Failed  int
	Pending int
}

type SubmissionRepository interface {
	// SaveSubmission фиксирует замену и возвращает прежний путь. При первой сдаче он пустой.
	SaveSubmission(ctx context.Context, submission Submission, archivePath string) (saved Submission, previousPath string, err error)
	GetSubmission(ctx context.Context, assignmentID, studentID string) (Submission, error)
	ListSubmissions(ctx context.Context, assignmentID string) ([]Submission, error)
	ArchivePath(ctx context.Context, submissionID, archiveKey string) (string, error)
	SaveReport(ctx context.Context, submissionID, archiveKey string, report BuildReport) (Submission, error)
}

type ArchiveRepository interface {
	Save(src io.Reader) (key, archivePath string, err error)
	Open(archivePath string) (io.ReadCloser, error)
	Remove(archivePath string) error
}

type Service struct {
	submissions SubmissionRepository
	archives    ArchiveRepository
}

func New(submissions SubmissionRepository, archives ArchiveRepository) *Service {
	return &Service{submissions: submissions, archives: archives}
}

// SaveSubmission сохраняет ZIP и возвращает признак первой сдачи.
func (s *Service) SaveSubmission(ctx context.Context, assignmentID, studentID string, src io.Reader) (saved Submission, created bool, err error) {
	if err := ctx.Err(); err != nil {
		return Submission{}, false, err
	}
	key, path, err := s.archives.Save(src)
	if err != nil {
		return Submission{}, false, err
	}

	saved, previousPath, err := s.submissions.SaveSubmission(ctx, Submission{
		SubmissionID: rand.Text(),
		AssignmentID: assignmentID,
		StudentID:    studentID,
		SubmittedAt:  time.Now().UTC(),
		ArchiveKey:   key,
	}, path)
	if err != nil {
		return Submission{}, false, errors.Join(err, s.archives.Remove(path))
	}
	if previousPath == "" {
		return saved, true, nil
	}
	if err := s.archives.Remove(previousPath); err != nil {
		slog.WarnContext(ctx, "remove replaced archive", "path", previousPath, "error", err)
	}
	return saved, false, nil
}

func (s *Service) GetSubmission(ctx context.Context, assignmentID, studentID string) (Submission, error) {
	return s.submissions.GetSubmission(ctx, assignmentID, studentID)
}

func (s *Service) GetAssignmentSummary(ctx context.Context, assignmentID string) (AssignmentSummary, error) {
	submissions, err := s.submissions.ListSubmissions(ctx, assignmentID)
	if err != nil {
		return AssignmentSummary{}, err
	}
	if submissions == nil {
		submissions = []Submission{}
	}
	counts := SummaryCounts{Total: len(submissions)}
	for _, submission := range submissions {
		switch {
		case submission.Report == nil:
			counts.Pending++
		case submission.Report.Success:
			counts.Passed++
		default:
			counts.Failed++
		}
	}
	return AssignmentSummary{Submissions: submissions, Counts: counts}, nil
}

func (s *Service) OpenArchive(ctx context.Context, submissionID, archiveKey string) (io.ReadCloser, error) {
	path, err := s.submissions.ArchivePath(ctx, submissionID, archiveKey)
	if err != nil {
		return nil, err
	}
	archive, err := s.archives.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		// Пересдача могла удалить ZIP между чтением пути и открытием файла.
		if _, lookupErr := s.submissions.ArchivePath(ctx, submissionID, archiveKey); lookupErr != nil {
			return nil, lookupErr
		}
	}
	return archive, err
}

func (s *Service) SaveReport(ctx context.Context, submissionID, archiveKey string, report BuildReport) (Submission, error) {
	return s.submissions.SaveReport(ctx, submissionID, archiveKey, report)
}
