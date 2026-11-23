package repository

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"kpo_hw3_devenv/internal/analyze/services"
)

type Storage struct {
	baseURL string
	client  *http.Client
}

func NewStorage(baseURL string, client *http.Client) *Storage {
	return &Storage{baseURL: strings.TrimRight(baseURL, "/"), client: client}
}

func (s *Storage) GetArchive(ctx context.Context, submissionID, archiveKey string) ([]byte, error) {
	endpoint := s.baseURL + "/internal/submissions/" + url.PathEscape(submissionID) + "/archives/" + url.PathEscape(archiveKey)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create storage request: %w", err)
	}

	response, err := s.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("get archive from storage: %w", err)
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusNotFound:
		return nil, services.ErrArchiveNotFound
	case http.StatusConflict:
		return nil, services.ErrArchiveChanged
	case http.StatusOK:
		data, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("read archive from storage: %w", err)
		}
		return data, nil
	default:
		return nil, fmt.Errorf("get archive from storage: HTTP %d", response.StatusCode)
	}
}
