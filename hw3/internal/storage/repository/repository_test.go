package repository_test

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	"kpo_hw3_devenv/internal/storage/repository"
	"kpo_hw3_devenv/internal/storage/services"
)

func TestResubmissionRejectsReportForPreviousArchive(t *testing.T) {
	ctx := t.Context()
	dataDir := filepath.Join(t.TempDir(), "ВышBuild #1")
	db := openSQLite(t, dataDir)
	archives, err := repository.NewArchiveStore(dataDir)
	if err != nil {
		t.Fatal(err)
	}
	firstKey, firstPath, err := archives.Save(bytes.NewReader(archiveWithComment(t, "first archive")))
	if err != nil {
		t.Fatal(err)
	}
	first := services.Submission{
		SubmissionID: "sub-1",
		AssignmentID: "hw3",
		StudentID:    "42",
		SubmittedAt:  time.Now(),
		ArchiveKey:   firstKey,
	}
	_, previousPath, err := db.SaveSubmission(ctx, first, firstPath)
	if err != nil || previousPath != "" {
		t.Fatalf("save first submission: previous path = %q, error = %v", previousPath, err)
	}
	passed := services.BuildReport{Status: "passed", Success: true, Message: "В архиве найден main.go"}
	if _, err := db.SaveReport(ctx, first.SubmissionID, firstKey, passed); err != nil {
		t.Fatal(err)
	}

	secondArchive := archiveWithComment(t, "second archive")
	secondKey, secondPath, err := archives.Save(bytes.NewReader(secondArchive))
	if err != nil {
		t.Fatal(err)
	}
	second := first
	second.SubmissionID = "sub-2"
	second.SubmittedAt = first.SubmittedAt.Add(time.Second)
	second.ArchiveKey = secondKey
	saved, previousPath, err := db.SaveSubmission(ctx, second, secondPath)
	if err != nil {
		t.Fatal(err)
	}
	if saved.SubmissionID != first.SubmissionID || saved.ArchiveKey == firstKey || previousPath != firstPath {
		t.Fatalf("replace submission: saved = %+v, previous path = %q", saved, previousPath)
	}
	if err := archives.Remove(previousPath); err != nil {
		t.Fatal(err)
	}
	if _, err := db.SaveReport(ctx, saved.SubmissionID, firstKey, passed); !errors.Is(err, services.ErrArchiveChanged) {
		t.Fatalf("save outdated report: got %v, want ErrArchiveChanged", err)
	}
	submissions, err := db.ListSubmissions(ctx, first.AssignmentID)
	if err != nil {
		t.Fatal(err)
	}
	if len(submissions) != 1 || submissions[0].Report != nil || !submissions[0].SubmittedAt.Equal(second.SubmittedAt) {
		t.Fatalf("list submissions after replacement: got %+v", submissions)
	}

	failed := services.BuildReport{Status: "failed", Message: "В архиве нет main.go"}
	if _, err := db.SaveReport(ctx, saved.SubmissionID, secondKey, failed); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db = openSQLite(t, dataDir)
	current, err := db.GetSubmission(ctx, first.AssignmentID, first.StudentID)
	if err != nil || current.Report == nil || *current.Report != failed {
		t.Fatalf("read persisted report: got %+v, error = %v", current, err)
	}
	currentPath, err := db.ArchivePath(ctx, current.SubmissionID, current.ArchiveKey)
	if err != nil || currentPath != secondPath {
		t.Fatalf("read persisted archive path: got %q, error = %v", currentPath, err)
	}
	reader, err := archives.Open(currentPath)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	data, err := io.ReadAll(reader)
	if err != nil || !bytes.Equal(data, secondArchive) {
		t.Fatalf("read current archive: got %q, error = %v", data, err)
	}
}

func TestConcurrentSubmissionsReturnTheReplacedArchive(t *testing.T) {
	dataDir := t.TempDir()
	firstDB, secondDB := openSQLite(t, dataDir), openSQLite(t, dataDir)
	ctx := t.Context()
	type result struct {
		submission services.Submission
		path       string
		previous   string
		err        error
	}
	start := make(chan struct{})
	results := make(chan result, 2)
	save := func(db *repository.SQLite, key string) {
		<-start
		path := "archives/" + key + ".zip"
		saved, previous, err := db.SaveSubmission(ctx, services.Submission{
			SubmissionID: key,
			AssignmentID: "hw3",
			StudentID:    "42",
			SubmittedAt:  time.Now(),
			ArchiveKey:   key,
		}, path)
		results <- result{saved, path, previous, err}
	}
	go save(firstDB, "first")
	go save(secondDB, "second")
	close(start)
	one, two := <-results, <-results
	if one.err != nil || two.err != nil {
		t.Fatalf("concurrent saves: %v, %v", one.err, two.err)
	}
	if one.previous != "" {
		one, two = two, one
	}
	if one.previous != "" || two.previous != one.path || one.submission.SubmissionID != two.submission.SubmissionID {
		t.Fatalf("expected one creation and one replacement: %+v, %+v", one, two)
	}
}

func TestArchiveStoreRemovesPartialFile(t *testing.T) {
	dataDir := t.TempDir()
	archives, err := repository.NewArchiveStore(dataDir)
	if err != nil {
		t.Fatal(err)
	}
	readErr := errors.New("upload interrupted")
	src := io.MultiReader(strings.NewReader("partial archive"), iotest.ErrReader(readErr))
	if _, _, err := archives.Save(src); !errors.Is(err, readErr) {
		t.Fatalf("write interrupted archive: got %v, want %v", err, readErr)
	}
	files, err := os.ReadDir(filepath.Join(dataDir, "archives"))
	if err != nil || len(files) != 0 {
		t.Fatalf("archive directory after interrupted save: entries = %v, error = %v", files, err)
	}
}

func openSQLite(t *testing.T, dataDir string) *repository.SQLite {
	t.Helper()
	db, err := repository.OpenSQLite(t.Context(), dataDir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
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
