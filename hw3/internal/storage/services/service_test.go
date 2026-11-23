package services_test

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"kpo_hw3_devenv/internal/storage/repository"
	"kpo_hw3_devenv/internal/storage/services"
)

func TestServiceReplacesArchiveAndCountsReports(t *testing.T) {
	ctx := t.Context()
	svc, db, archives, _ := openStorage(t)
	first, created, err := svc.SaveSubmission(ctx, "hw3", "42", bytes.NewReader(archiveWithComment(t, "first archive")))
	if err != nil || !created {
		t.Fatalf("first submission: created = %t, error = %v", created, err)
	}
	oldPath, err := db.ArchivePath(ctx, first.SubmissionID, first.ArchiveKey)
	if err != nil {
		t.Fatal(err)
	}
	newArchive := archiveWithComment(t, "new archive")
	replaced, created, err := svc.SaveSubmission(ctx, "hw3", "42", bytes.NewReader(newArchive))
	if err != nil || created || replaced.SubmissionID != first.SubmissionID {
		t.Fatalf("replace submission: saved = %+v, created = %t, error = %v", replaced, created, err)
	}
	if old, err := archives.Open(oldPath); !errors.Is(err, fs.ErrNotExist) {
		if old != nil {
			_ = old.Close()
		}
		t.Fatalf("open replaced archive: got %v, want ErrNotExist", err)
	}

	reader, err := svc.OpenArchive(ctx, replaced.SubmissionID, replaced.ArchiveKey)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	data, err := io.ReadAll(reader)
	if err != nil || !bytes.Equal(data, newArchive) {
		t.Fatalf("read current archive: got %q, error = %v", data, err)
	}
	if _, err := svc.SaveReport(ctx, first.SubmissionID, first.ArchiveKey, services.BuildReport{Status: "failed"}); !errors.Is(err, services.ErrArchiveChanged) {
		t.Fatalf("save stale report: got %v, want ErrArchiveChanged", err)
	}
	if _, err := svc.SaveReport(ctx, replaced.SubmissionID, replaced.ArchiveKey, services.BuildReport{Status: "passed", Success: true}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.SaveSubmission(ctx, "hw3", "43", bytes.NewReader(archiveWithComment(t, "other archive"))); err != nil {
		t.Fatal(err)
	}
	summary, err := svc.GetAssignmentSummary(ctx, "hw3")
	want := services.SummaryCounts{Total: 2, Passed: 1, Pending: 1}
	if err != nil || summary.Counts != want || len(summary.Submissions) != 2 {
		t.Fatalf("summary: got %+v, want counts %+v, error = %v", summary, want, err)
	}
}

func TestServiceRemovesNewArchiveWhenDatabaseSaveFails(t *testing.T) {
	svc, db, _, dataDir := openStorage(t)
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.SaveSubmission(t.Context(), "hw3", "42", bytes.NewReader(archiveWithComment(t, "archive"))); err == nil {
		t.Fatal("expected a database error")
	}
	files, err := os.ReadDir(filepath.Join(dataDir, "archives"))
	if err != nil || len(files) != 0 {
		t.Fatalf("archive directory after database error: entries = %v, error = %v", files, err)
	}
}

func openStorage(t *testing.T) (*services.Service, *repository.SQLite, *repository.ArchiveStore, string) {
	t.Helper()
	dataDir := t.TempDir()
	db, err := repository.OpenSQLite(t.Context(), dataDir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	archives, err := repository.NewArchiveStore(dataDir)
	if err != nil {
		t.Fatal(err)
	}
	return services.New(db, archives), db, archives, dataDir
}

func archiveWithComment(t *testing.T, comment string) []byte {
	t.Helper()
	var data bytes.Buffer
	writer := zip.NewWriter(&data)
	if err := writer.SetComment(comment); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return data.Bytes()
}
