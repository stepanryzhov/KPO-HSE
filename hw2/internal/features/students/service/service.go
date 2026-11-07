package service

import (
	"errors"

	"kpo_hw2/internal/core/apperrors"
	"kpo_hw2/internal/core/domain"
)

type Repository interface {
	Save(registry domain.StudentRegistry) error
}

type Service struct {
	repository Repository
	registry   *domain.StudentRegistry
}

func NewService(repository Repository, registry *domain.StudentRegistry) (*Service, error) {
	if repository == nil {
		return nil, errors.New("репозиторий студентов не задан")
	}
	if registry == nil {
		return nil, errors.New("реестр студентов не задан")
	}
	return &Service{repository: repository, registry: registry}, nil
}

func (s *Service) AddStudent(rawName string) error {
	name, err := domain.NewName(rawName)
	if err != nil {
		return err
	}
	_, err = s.registry.CreateNewStudent(name)
	if err != nil {
		return err
	}
	if err := s.repository.Save(*s.registry); err != nil {
		return &apperrors.SaveError{Err: err}
	}
	return nil
}

func (s *Service) GetStudent(rawID int64) (domain.Student, error) {
	id, err := domain.NewID(rawID)
	if err != nil {
		return domain.Student{}, err
	}
	return s.registry.GetStudentByID(id)
}

func (s *Service) ListStudents() []domain.Student {
	return s.registry.Students()
}
