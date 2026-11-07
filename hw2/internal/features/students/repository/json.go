package repository

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"kpo_hw2/internal/core/domain"
	"kpo_hw2/internal/features/students/service"
)

var _ service.Repository = (*JSONRepository)(nil)

type JSONRepository struct {
	path string
}

func NewJSONRepository(path string) *JSONRepository {
	return &JSONRepository{path: path}
}

func (r *JSONRepository) Load() (*domain.StudentRegistry, error) {
	if r.path == "" {
		return nil, errors.New("путь к файлу студентов пуст")
	}
	data, err := os.ReadFile(r.path)
	if errors.Is(err, os.ErrNotExist) {
		return domain.NewStudentRegistry(0)
	}
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var dto *studentRegistryDTO
	if err := decoder.Decode(&dto); err != nil {
		return nil, fmt.Errorf("чтение JSON студентов: %w", err)
	}
	if dto == nil {
		return nil, errors.New("вместо реестра студентов записан null")
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return nil, errors.New("файл должен содержать один JSON-объект")
	}
	return dto.toDomain()
}

func (r *JSONRepository) Save(registry domain.StudentRegistry) error {
	if r.path == "" {
		return errors.New("путь к файлу студентов пуст")
	}
	data := registryToDTO(registry)
	dir := filepath.Dir(r.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	file, err := os.CreateTemp(dir, ".students-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), r.path)
}
