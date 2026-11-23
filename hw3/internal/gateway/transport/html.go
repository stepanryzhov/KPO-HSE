package transport

import (
	"bytes"
	"embed"
	"html/template"
	"log/slog"
	"net/http"
	"time"

	"kpo_hw3_devenv/internal/dto"
)

//go:embed templates/*.html
var templateFiles embed.FS

var moscow = time.FixedZone("МСК", 3*60*60)

var templates = template.Must(template.New("").Funcs(template.FuncMap{
	"reportStatus": reportStatus,
	"submittedAt": func(at time.Time) string {
		return at.In(moscow).Format("02.01.2006, 15:04 MST")
	},
}).ParseFS(templateFiles, "templates/*.html"))

type summaryPage struct {
	AssignmentID string
	dto.AssignmentSummary
}

func reportStatus(report *dto.BuildReport) string {
	if report == nil {
		return "Результат проверки пока недоступен"
	}
	if report.Success {
		return "Проверка пройдена"
	}
	return "Проверка не пройдена"
}

func writeHTML(w http.ResponseWriter, status int, name string, data any) {
	var body bytes.Buffer
	if err := templates.ExecuteTemplate(&body, name, data); err != nil {
		slog.Error("render gateway HTML response", "template", name, "error", err)
		http.Error(w, "Не удалось сформировать страницу", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if _, err := body.WriteTo(w); err != nil {
		slog.Warn("send gateway HTML response", "error", err)
	}
}
