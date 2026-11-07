package export

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"

	gradebookservice "kpo_hw2/internal/features/gradebooks/service"
)

// CSV записывает таблицу оценок и итогов. Поток открывает и закрывает вызывающий код.
type CSV struct{}

func (CSV) Write(writer io.Writer, report gradebookservice.Report) error {
	header := []string{"Имя"}
	for _, criterion := range report.Criteria {
		header = append(header, criterion.GetName().Value())
	}
	header = append(header, "Итог")
	records := make([][]string, 0, len(report.Rows)+1)
	records = append(records, header)

	for _, result := range report.Rows {
		row := []string{result.Student.GetName().Value()}
		for _, criterion := range report.Criteria {
			score, exists := result.Scores[criterion.GetID().Value()]
			if !exists {
				return fmt.Errorf("нет оценки студента %d по критерию %q", result.Student.GetID().Value(), criterion.GetName().Value())
			}
			row = append(row, strconv.FormatFloat(score, 'f', -1, 64))
		}
		row = append(row, strconv.FormatFloat(result.FinalMark, 'f', -1, 64))
		records = append(records, row)
	}

	return csv.NewWriter(writer).WriteAll(records)
}
