package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"kpo_hw3_devenv/internal/storage/services"

	_ "modernc.org/sqlite"
)

const createSubmissionsTable = `
CREATE TABLE IF NOT EXISTS submissions (
    submission_id TEXT NOT NULL PRIMARY KEY,
    assignment_id TEXT NOT NULL,
    student_id    TEXT NOT NULL,
    submitted_at  TEXT NOT NULL,
    archive_key   TEXT NOT NULL UNIQUE,
    archive_path  TEXT NOT NULL,
    report_json   TEXT,
    UNIQUE (assignment_id, student_id)
)`

const submissionColumns = `submission_id, assignment_id, student_id, submitted_at, archive_key, archive_path, report_json`

type submissionRow struct {
	submissionID string
	assignmentID string
	studentID    string
	submittedAt  string
	archiveKey   string
	archivePath  string
	reportJSON   sql.NullString
}

// SQLite хранит метаданные и отчёты в storage.sqlite внутри каталога данных.
type SQLite struct {
	db *sql.DB
}

func OpenSQLite(ctx context.Context, dataDir string) (*SQLite, error) {
	dataDir, err := filepath.Abs(dataDir)
	if err != nil {
		return nil, fmt.Errorf("resolve data directory: %w", err)
	}
	if err := os.MkdirAll(dataDir, 0o750); err != nil {
		return nil, fmt.Errorf("create data directory: %w", err)
	}

	// Блокировка на запись берётся перед чтением прежнего пути при пересдаче.
	params := url.Values{
		"_pragma": {"busy_timeout(5000)"},
		"_txlock": {"immediate"},
	}
	dsn := url.URL{
		Scheme:   "file",
		Path:     filepath.ToSlash(filepath.Join(dataDir, "storage.sqlite")),
		RawQuery: params.Encode(),
	}
	db, err := sql.Open("sqlite", dsn.String())
	if err != nil {
		return nil, fmt.Errorf("open SQLite: %w", err)
	}
	db.SetMaxOpenConns(1)
	if _, err := db.ExecContext(ctx, createSubmissionsTable); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("create submissions table: %w", err)
	}
	return &SQLite{db: db}, nil
}

func (r *SQLite) Close() error {
	return r.db.Close()
}

// SaveSubmission сохраняет ссылку на новый ZIP и сбрасывает отчёт.
// Прежний путь возвращается после коммита. При первой сдаче путь пустой.
func (r *SQLite) SaveSubmission(ctx context.Context, submission services.Submission, archivePath string) (services.Submission, string, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return services.Submission{}, "", fmt.Errorf("begin submission transaction: %w", err)
	}
	defer tx.Rollback()

	var previousPath string
	err = tx.QueryRowContext(ctx, `
        SELECT archive_path FROM submissions WHERE assignment_id = ? AND student_id = ?`,
		submission.AssignmentID, submission.StudentID,
	).Scan(&previousPath)
	
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return services.Submission{}, "", fmt.Errorf("read previous archive path: %w", err)
	}

	submission.SubmittedAt = submission.SubmittedAt.UTC()
	submission.Report = nil
	err = tx.QueryRowContext(ctx, `
        INSERT INTO submissions (submission_id, assignment_id, student_id, submitted_at, archive_key, archive_path)
        VALUES (?, ?, ?, ?, ?, ?)
        ON CONFLICT (assignment_id, student_id) DO UPDATE SET
            submitted_at = excluded.submitted_at,
            archive_key  = excluded.archive_key,
            archive_path = excluded.archive_path,
            report_json  = NULL
        RETURNING submission_id`,
		submission.SubmissionID, submission.AssignmentID, submission.StudentID,
		submission.SubmittedAt.Format(time.RFC3339Nano), submission.ArchiveKey, archivePath,
	).Scan(&submission.SubmissionID)
	if err != nil {
		return services.Submission{}, "", fmt.Errorf("save submission: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return services.Submission{}, "", fmt.Errorf("commit submission: %w", err)
	}
	return submission, previousPath, nil
}

func (r *SQLite) GetSubmission(ctx context.Context, assignmentID, studentID string) (services.Submission, error) {
	submission, err := scanSubmission(r.db.QueryRowContext(ctx,
		`SELECT `+submissionColumns+` FROM submissions WHERE assignment_id = ? AND student_id = ?`,
		assignmentID, studentID,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return services.Submission{}, services.ErrSubmissionNotFound
	}
	if err != nil {
		return services.Submission{}, fmt.Errorf("get submission: %w", err)
	}
	return submission, nil
}

func (r *SQLite) ListSubmissions(ctx context.Context, assignmentID string) ([]services.Submission, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+submissionColumns+` FROM submissions WHERE assignment_id = ? ORDER BY student_id`,
		assignmentID,
	)
	if err != nil {
		return nil, fmt.Errorf("list submissions: %w", err)
	}
	defer rows.Close()

	submissions := make([]services.Submission, 0)
	for rows.Next() {
		submission, err := scanSubmission(rows)
		if err != nil {
			return nil, fmt.Errorf("read submission: %w", err)
		}
		submissions = append(submissions, submission)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list submissions: %w", err)
	}
	return submissions, nil
}

func (r *SQLite) ArchivePath(ctx context.Context, submissionID, archiveKey string) (string, error) {
	var currentKey, archivePath string
	err := r.db.QueryRowContext(ctx, `SELECT archive_key, archive_path FROM submissions WHERE submission_id = ?`,
		submissionID,
	).Scan(&currentKey, &archivePath)
	if errors.Is(err, sql.ErrNoRows) {
		return "", services.ErrSubmissionNotFound
	}
	if err != nil {
		return "", fmt.Errorf("get archive path: %w", err)
	}
	if currentKey != archiveKey {
		return "", services.ErrArchiveChanged
	}
	return archivePath, nil
}

func (r *SQLite) SaveReport(ctx context.Context, submissionID, archiveKey string, report services.BuildReport) (services.Submission, error) {
	data, err := json.Marshal(report)
	if err != nil {
		return services.Submission{}, fmt.Errorf("encode build report: %w", err)
	}
	submission, err := scanSubmission(r.db.QueryRowContext(ctx, `
        UPDATE submissions SET report_json = ?
        WHERE submission_id = ? AND archive_key = ?
        RETURNING `+submissionColumns,
		string(data), submissionID, archiveKey,
	))
	if errors.Is(err, sql.ErrNoRows) {
		// Уточняем, существует ли сдача с таким ID для ответа вызывающему сервису.
		if _, err := r.ArchivePath(ctx, submissionID, archiveKey); err != nil {
			return services.Submission{}, err
		}
		return services.Submission{}, services.ErrArchiveChanged
	}
	if err != nil {
		return services.Submission{}, fmt.Errorf("save build report: %w", err)
	}
	return submission, nil
}

func scanSubmission(row interface{ Scan(...any) error }) (services.Submission, error) {
	var stored submissionRow
	if err := row.Scan(
		&stored.submissionID, &stored.assignmentID, &stored.studentID, &stored.submittedAt,
		&stored.archiveKey, &stored.archivePath, &stored.reportJSON,
	); err != nil {
		return services.Submission{}, err
	}
	submittedAt, err := time.Parse(time.RFC3339Nano, stored.submittedAt)
	if err != nil {
		return services.Submission{}, fmt.Errorf("parse submitted_at: %w", err)
	}
	var report *services.BuildReport
	if stored.reportJSON.Valid {
		report = new(services.BuildReport)
		if err := json.Unmarshal([]byte(stored.reportJSON.String), report); err != nil {
			return services.Submission{}, fmt.Errorf("decode build report: %w", err)
		}
	}
	return services.Submission{
		SubmissionID: stored.submissionID,
		AssignmentID: stored.assignmentID,
		StudentID:    stored.studentID,
		SubmittedAt:  submittedAt,
		ArchiveKey:   stored.archiveKey,
		Report:       report,
	}, nil
}
