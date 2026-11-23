// Пакет dto содержит структуры межсервисных HTTP-запросов и ответов.
package dto

import "time"

// ArchiveRef указывает на ZIP конкретной загрузки.
type ArchiveRef struct {
	SubmissionID string `json:"submission_id"`
	ArchiveKey   string `json:"archive_key"`
}

// BuildReport содержит результат проверки ZIP-архива.
type BuildReport struct {
	Status  string `json:"status"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// SubmissionResponse содержит текущую сдачу и её отчёт, если он уже сохранён.
type SubmissionResponse struct {
	SubmissionID string       `json:"submission_id"`
	AssignmentID string       `json:"assignment_id"`
	StudentID    string       `json:"student_id"`
	SubmittedAt  time.Time    `json:"submitted_at"`
	Report       *BuildReport `json:"report"`
}

// AssignmentSummary содержит текущие сдачи и счётчики по одному заданию.
type AssignmentSummary struct {
	// Для задания без сдач Submissions заполняется пустым слайсом.
	Submissions []SubmissionResponse `json:"submissions"`
	Counts      SummaryCounts        `json:"counts"`
}

// SummaryCounts содержит общее число сдач и количество по каждому результату проверки.
type SummaryCounts struct {
	Total   int `json:"total"`
	Passed  int `json:"passed"`
	Failed  int `json:"failed"`
	Pending int `json:"pending"`
}

// ErrorResponse описывает HTTP-ошибку в JSON.
// Поле SubmissionID заполняется, если Storage подтвердил сохранение ZIP.
type ErrorResponse struct {
	Message      string `json:"message"`
	SubmissionID string `json:"submission_id,omitempty"`
}
