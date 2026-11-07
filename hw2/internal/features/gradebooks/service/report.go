package service

import "kpo_hw2/internal/core/domain"

// Report — сводка для консоли и экспорта: доменные объекты и рассчитанные итоги.
type Report struct {
	Criteria []domain.Criterion
	Rows     []ReportRow
}

type ReportRow struct {
	Student   domain.Student
	Scores    map[int64]float64
	FinalMark float64
}
