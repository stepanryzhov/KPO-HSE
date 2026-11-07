package repository

import (
	"fmt"

	"kpo_hw2/internal/core/domain"
)

type criterionDTO struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	MaxScore float64 `json:"max_score"`
	Weight   float64 `json:"weight"`
}

type evaluationDTO struct {
	StudentID   int64   `json:"student_id"`
	CriterionID int64   `json:"criterion_id"`
	Score       float64 `json:"score"`
}

type gradebookDTO struct {
	StudentIDs      []int64         `json:"student_ids"`
	LastCriterionID int64           `json:"last_criterion_id"`
	Criteria        []criterionDTO  `json:"criteria"`
	Evaluations     []evaluationDTO `json:"evaluations"`
}

func (dto gradebookDTO) toDomain(students domain.StudentRegistry) (*domain.Gradebook, error) {
	book, err := domain.RestoreGradebook(dto.LastCriterionID)
	if err != nil {
		return nil, fmt.Errorf("восстановление счётчика критериев: %w", err)
	}
	for _, criterion := range dto.Criteria {
		if err := book.RestoreCriterion(criterion.ID, criterion.Name, criterion.MaxScore, criterion.Weight); err != nil {
			return nil, fmt.Errorf("восстановление критерия %d: %w", criterion.ID, err)
		}
	}
	for _, rawID := range dto.StudentIDs {
		id, err := domain.NewID(rawID)
		if err != nil {
			return nil, err
		}
		if err := students.AddToGradebook(book, id); err != nil {
			return nil, fmt.Errorf("восстановление студента %d в ведомости: %w", rawID, err)
		}
	}
	seen := make(map[[2]int64]bool, len(dto.Evaluations))
	for _, evaluation := range dto.Evaluations {
		key := [2]int64{evaluation.StudentID, evaluation.CriterionID}
		if seen[key] {
			return nil, fmt.Errorf("повтор оценки студента %d по критерию %d", evaluation.StudentID, evaluation.CriterionID)
		}
		studentID, err := domain.NewID(evaluation.StudentID)
		if err != nil {
			return nil, err
		}
		criterionID, err := domain.NewID(evaluation.CriterionID)
		if err != nil {
			return nil, err
		}
		score, err := domain.NewScore(evaluation.Score)
		if err != nil {
			return nil, err
		}
		if err := book.SetEvaluation(studentID, criterionID, score); err != nil {
			return nil, fmt.Errorf("восстановление оценки студента %d по критерию %d: %w", evaluation.StudentID, evaluation.CriterionID, err)
		}
		seen[key] = true
	}
	return book, nil
}

func gradebookToDTO(book domain.Gradebook) (gradebookDTO, error) {
	dto := gradebookDTO{
		StudentIDs:      make([]int64, 0),
		LastCriterionID: book.LastCriterionID(),
		Criteria:        make([]criterionDTO, 0),
		Evaluations:     make([]evaluationDTO, 0),
	}
	for _, criterion := range book.Criteria() {
		dto.Criteria = append(dto.Criteria, criterionDTO{
			ID:       criterion.GetID().Value(),
			Name:     criterion.GetName().Value(),
			MaxScore: criterion.GetMaxScore().Value(),
			Weight:   criterion.GetWeight().Value(),
		})
	}
	for _, student := range book.Students() {
		id := student.GetID()
		dto.StudentIDs = append(dto.StudentIDs, id.Value())
		evaluations, err := book.EvaluationsFor(id)
		if err != nil {
			return gradebookDTO{}, err
		}
		for _, evaluation := range evaluations {
			dto.Evaluations = append(dto.Evaluations, evaluationDTO{
				StudentID:   evaluation.GetStudentID().Value(),
				CriterionID: evaluation.GetCriterionID().Value(),
				Score:       evaluation.GetScore().Value(),
			})
		}
	}
	return dto, nil
}
