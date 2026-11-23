package transport_test

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"kpo_hw3_devenv/internal/dto"
	"kpo_hw3_devenv/internal/storage/repository"
	"kpo_hw3_devenv/internal/storage/services"
	"kpo_hw3_devenv/internal/storage/transport"
)

func TestStorageHTTPSubmissionLifecycle(t *testing.T) {
	dataDir := t.TempDir()
	db, err := repository.OpenSQLite(t.Context(), dataDir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	archives, err := repository.NewArchiveStore(dataDir)
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	transport.NewHandler(services.New(db, archives)).SetRoutes(mux)

	const submissionPath = "/internal/assignments/hw3/submissions?student_id=42"
	const summaryPath = "/internal/assignments/hw3/summary"
	var summary dto.AssignmentSummary
	decodeResponse(t, request(t, mux, http.MethodGet, summaryPath, "", nil, http.StatusOK), &summary)
	if summary.Submissions == nil || len(summary.Submissions) != 0 || summary.Counts != (dto.SummaryCounts{}) {
		t.Fatalf("empty summary: %+v", summary)
	}
	request(t, mux, http.MethodGet, submissionPath, "", nil, http.StatusNotFound)

	firstZIP := zipWithMain(t, "package main")
	var first dto.ArchiveRef
	decodeResponse(t, request(t, mux, http.MethodPost, submissionPath, "application/zip", firstZIP, http.StatusCreated), &first)
	if first.SubmissionID == "" || first.ArchiveKey == "" {
		t.Fatalf("first archive ref: %+v", first)
	}
	archivePath := "/internal/submissions/" + first.SubmissionID + "/archives/" + first.ArchiveKey
	response := request(t, mux, http.MethodGet, archivePath, "", nil, http.StatusOK)
	if response.Header().Get("Content-Type") != "application/zip" || !bytes.Equal(response.Body.Bytes(), firstZIP) {
		t.Fatal("archive response must contain the original ZIP bytes and media type")
	}

	reportPath := archivePath + "/report"
	invalidReport := []byte(`{"status":"passed","success":false,"message":"main.go найден"}`)
	request(t, mux, http.MethodPut, reportPath, "application/json", invalidReport, http.StatusBadRequest)
	report := []byte(`{"status":"passed","success":true,"message":"main.go найден"}`)
	var checked dto.SubmissionResponse
	decodeResponse(t, request(t, mux, http.MethodPut, reportPath, "application/json", report, http.StatusOK), &checked)
	if checked.Report == nil || !checked.Report.Success || checked.SubmissionID != first.SubmissionID {
		t.Fatalf("saved report: %+v", checked)
	}
	var invalidZIP dto.ErrorResponse
	decodeResponse(t, request(t, mux, http.MethodPost, submissionPath, "application/zip", []byte("invalid ZIP"), http.StatusBadRequest), &invalidZIP)
	if invalidZIP.Message != "Некорректный ZIP-архив" || invalidZIP.SubmissionID != "" {
		t.Fatalf("invalid ZIP response: %+v", invalidZIP)
	}
	var afterInvalidZIP dto.SubmissionResponse
	decodeResponse(t, request(t, mux, http.MethodGet, submissionPath, "", nil, http.StatusOK), &afterInvalidZIP)
	if afterInvalidZIP.SubmissionID != checked.SubmissionID || !afterInvalidZIP.SubmittedAt.Equal(checked.SubmittedAt) ||
		afterInvalidZIP.Report == nil || *afterInvalidZIP.Report != *checked.Report {
		t.Fatalf("submission after invalid replacement: %+v", afterInvalidZIP)
	}
	files, err := os.ReadDir(filepath.Join(dataDir, "archives"))
	if err != nil || len(files) != 1 || files[0].Name() != first.ArchiveKey+".zip" {
		t.Fatalf("archive directory after invalid replacement: entries = %v, error = %v", files, err)
	}

	var second dto.ArchiveRef
	secondZIP := zipWithMain(t, "package main\nfunc main() {}")
	decodeResponse(t, request(t, mux, http.MethodPost, submissionPath, "application/zip", secondZIP, http.StatusOK), &second)
	if second.SubmissionID != first.SubmissionID || second.ArchiveKey == first.ArchiveKey {
		t.Fatalf("replacement archive ref: %+v", second)
	}
	var conflict dto.ErrorResponse
	decodeResponse(t, request(t, mux, http.MethodPut, reportPath, "application/json", report, http.StatusConflict), &conflict)
	if conflict.SubmissionID != first.SubmissionID {
		t.Fatalf("stale report response: %+v", conflict)
	}
	var current dto.SubmissionResponse
	decodeResponse(t, request(t, mux, http.MethodGet, submissionPath, "", nil, http.StatusOK), &current)
	if current.Report != nil || current.SubmissionID != first.SubmissionID {
		t.Fatalf("current submission after replacement: %+v", current)
	}
	decodeResponse(t, request(t, mux, http.MethodGet, summaryPath, "", nil, http.StatusOK), &summary)
	if len(summary.Submissions) != 1 || summary.Counts != (dto.SummaryCounts{Total: 1, Pending: 1}) {
		t.Fatalf("summary after replacement: %+v", summary)
	}
}

func request(t *testing.T, handler http.Handler, method, path, contentType string, body []byte, wantStatus int) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	if response.Code != wantStatus {
		t.Fatalf("%s %s: got %d, want %d, body = %s", method, path, response.Code, wantStatus, response.Body)
	}
	return response
}

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder, destination any) {
	t.Helper()
	if contentType := response.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("content type = %q, want application/json", contentType)
	}
	if err := json.Unmarshal(response.Body.Bytes(), destination); err != nil {
		t.Fatal(err)
	}
}

func zipWithMain(t *testing.T, source string) []byte {
	t.Helper()
	var buffer bytes.Buffer
	archive := zip.NewWriter(&buffer)
	file, err := archive.Create("main.go")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.Write([]byte(source)); err != nil {
		t.Fatal(err)
	}
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}
