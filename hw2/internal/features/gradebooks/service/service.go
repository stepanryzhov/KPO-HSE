package service

import (
	"errors"

	"kpo_hw2/internal/core/apperrors"
	"kpo_hw2/internal/core/domain"
)

type Repository interface {
	Save(book domain.Gradebook) error
}

// Service работает с одной ведомостью и общим реестром студентов.
type Service struct {
	repository Repository
	students   *domain.StudentRegistry
	book       *domain.Gradebook
}

func NewService(repository Repository, students *domain.StudentRegistry, book *domain.Gradebook) (*Service, error) {
	if repository == nil {
		return nil, errors.New("репозиторий ведомости не задан")
	}
	if students == nil {
		return nil, errors.New("реестр студентов не задан")
	}
	if book == nil {
		return nil, errors.New("ведомость не задана")
	}
	return &Service{repository: repository, students: students, book: book}, nil
}

func (s *Service) EnrollStudent(rawID int64) error {
	id, err := domain.NewID(rawID)
	if err != nil {
		return err
	}
	if err := s.students.AddToGradebook(s.book, id); err != nil {
		return err
	}
	return s.save()
}

func (s *Service) ListStudents() []domain.Student {
	return s.book.Students()
}

func (s *Service) GetEvaluations(rawID int64) ([]domain.Evaluation, error) {
	id, err := domain.NewID(rawID)
	if err != nil {
		return nil, err
	}
	return s.book.EvaluationsFor(id)
}

func (s *Service) CreateCriterion(rawName string, rawMaxScore, rawWeight float64) (domain.Criterion, error) {
	name, err := domain.NewName(rawName)
	if err != nil {
		return domain.Criterion{}, err
	}
	maxScore, err := domain.NewScore(rawMaxScore)
	if err != nil {
		return domain.Criterion{}, err
	}
	weight, err := domain.NewWeight(rawWeight)
	if err != nil {
		return domain.Criterion{}, err
	}
	criterion, err := s.book.CreateCriterion(name, maxScore, weight)
	if err != nil {
		return domain.Criterion{}, err
	}
	return criterion, s.save()
}

func (s *Service) ListCriteria() []domain.Criterion {
	return s.book.Criteria()
}

func (s *Service) GetCriterion(rawID int64) (domain.Criterion, error) {
	id, err := domain.NewID(rawID)
	if err != nil {
		return domain.Criterion{}, err
	}
	return s.book.Criterion(id)
}

// SetEvaluation выставляет новую оценку или заменяет существующую.
func (s *Service) SetEvaluation(rawStudentID, rawCriterionID int64, rawScore float64) error {
	studentID, err := domain.NewID(rawStudentID)
	if err != nil {
		return err
	}
	criterionID, err := domain.NewID(rawCriterionID)
	if err != nil {
		return err
	}
	score, err := domain.NewScore(rawScore)
	if err != nil {
		return err
	}
	if err := s.book.SetEvaluation(studentID, criterionID, score); err != nil {
		return err
	}
	return s.save()
}

func (s *Service) RemoveEvaluation(rawStudentID, rawCriterionID int64) error {
	studentID, err := domain.NewID(rawStudentID)
	if err != nil {
		return err
	}
	criterionID, err := domain.NewID(rawCriterionID)
	if err != nil {
		return err
	}
	if err := s.book.RemoveEvaluation(studentID, criterionID); err != nil {
		return err
	}
	return s.save()
}

func (s *Service) CalculateFinalMark(rawID int64) (float64, error) {
	id, err := domain.NewID(rawID)
	if err != nil {
		return 0, err
	}
	mark, err := domain.CalculateFinalMark(s.book, id)
	if err != nil {
		return 0, err
	}
	return mark.Value(), nil
}

func (s *Service) GetReport() (Report, error) {
	marks, err := domain.CalculateAllFinalMarks(s.book)
	if err != nil {
		return Report{}, err
	}
	report := Report{Criteria: s.ListCriteria(), Rows: make([]ReportRow, 0, len(marks))}
	for _, student := range s.book.Students() {
		evaluations, err := s.book.EvaluationsFor(student.GetID())
		if err != nil {
			return Report{}, err
		}
		scores := make(map[int64]float64, len(evaluations))
		for _, evaluation := range evaluations {
			scores[evaluation.GetCriterionID().Value()] = evaluation.GetScore().Value()
		}
		report.Rows = append(report.Rows, ReportRow{
			Student:   student,
			Scores:    scores,
			FinalMark: marks[student.GetID()].Value(),
		})
	}
	return report, nil
}

func (s *Service) save() error {
	if err := s.repository.Save(*s.book); err != nil {
		return &apperrors.SaveError{Err: err}
	}
	return nil
}
