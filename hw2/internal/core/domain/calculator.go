package domain

import (
	"errors"
	"fmt"
	"math"
)

// CalculateFinalMark рассчитывает сумму баллов, умноженных на веса.
// Для расчёта необходимы оценки по всем критериям.
func CalculateFinalMark(book *Gradebook, sid studentID) (score, error) {
	if _, err := book.Student(sid); err != nil {
		return score{}, err
	}

	criteria := book.Criteria()
	if len(criteria) == 0 {
		return score{}, errors.New("нет критериев")
	}

	var total float64

	for _, criterion := range criteria {
		evaluation, exists := book.evaluations[sid][criterion.GetID()]
		if !exists {
			return score{}, fmt.Errorf(
				"нет оценки по критерию %q",
				criterion.GetName().Value(),
			)
		}

		total += evaluation.GetScore().Value() *
			criterion.GetWeight().Value()

		if math.IsInf(total, 0) {
			return score{}, errors.New("переполнение при расчёте итога")
		}
	}

	s, err := NewScore(total)
	if err != nil {
		return score{}, fmt.Errorf("Подсчет итоговой оценки: %v", err)
	}
	return s, nil
}

func CalculateAllFinalMarks(book *Gradebook) (map[studentID]score, error) {
	results := make(map[studentID]score)

	for _, student := range book.Students() {
		sid := student.GetID()

		mark, err := CalculateFinalMark(book, sid)
		if err != nil {
			return nil, fmt.Errorf(
				"не удалось рассчитать итог студента %d: %w",
				sid.Value(),
				err,
			)
		}

		results[sid] = mark
	}

	return results, nil
}
