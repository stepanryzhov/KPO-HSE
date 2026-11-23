package repository

import (
	"context"
	"net/http"

	"kpo_hw3_devenv/internal/dto"
	"kpo_hw3_devenv/internal/gateway/services"
)

type Analyze struct {
	client client
}

func NewAnalyze(baseURL string, httpClient *http.Client) *Analyze {
	return &Analyze{client: newClient(baseURL, httpClient, services.ErrAnalyzeUnavailable)}
}

func (a *Analyze) Analyze(ctx context.Context, ref dto.ArchiveRef) (dto.BuildReport, error) {
	var report dto.BuildReport
	status, err := a.client.sendJSON(ctx, http.MethodPost, "/internal/analyze", ref, &report)
	if status == http.StatusConflict {
		return dto.BuildReport{}, services.ErrArchiveChanged
	}
	return report, err
}
