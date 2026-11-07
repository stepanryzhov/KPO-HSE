package repository

import (
	"fmt"

	"kpo_hw2/internal/core/domain"
)

type studentDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type studentRegistryDTO struct {
	LastID   int64        `json:"last_id"`
	Students []studentDTO `json:"students"`
}

func (dto studentRegistryDTO) toDomain() (*domain.StudentRegistry, error) {
	registry, err := domain.NewStudentRegistry(dto.LastID)
	if err != nil {
		return nil, err
	}
	for _, student := range dto.Students {
		if err := registry.RestoreStudent(student.ID, student.Name); err != nil {
			return nil, fmt.Errorf("восстановление студента %d: %w", student.ID, err)
		}
	}
	return registry, nil
}

func registryToDTO(registry domain.StudentRegistry) studentRegistryDTO {
	dto := studentRegistryDTO{LastID: registry.LastID(), Students: make([]studentDTO, 0)}
	for _, student := range registry.Students() {
		dto.Students = append(dto.Students, studentDTO{
			ID: student.GetID().Value(), Name: student.GetName().Value(),
		})
	}
	return dto
}
