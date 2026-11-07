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
	"kpo_hw2/internal/features/gradebooks/service"
)

var _ service.Repository = (*JSONRepository)(nil)

// JSONRepository хранит одну ведомость в одном файле.
type JSONRepository struct {
	path string
}

func NewJSONRepository(path string) *JSONRepository {
	return &JSONRepository{path: path}
}

func (r *JSONRepository) Load(students domain.StudentRegistry) (*domain.Gradebook, error) {
	if r.path == "" {
		return nil, errors.New("путь к файлу ведомости пуст")
	}
	data, err := os.ReadFile(r.path)
	if errors.Is(err, os.ErrNotExist) {
		return domain.NewGradebook(), nil
	}
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var dto *gradebookDTO
	if err := decoder.Decode(&dto); err != nil {
		return nil, fmt.Errorf("чтение JSON ведомости: %w", err)
	}
	if dto == nil {
		return nil, errors.New("вместо ведомости записан null")
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return nil, errors.New("файл должен содержать один JSON-объект")
	}
	return dto.toDomain(students)
}

func (r *JSONRepository) Save(book domain.Gradebook) error {
	if r.path == "" {
		return errors.New("путь к файлу ведомости пуст")
	}
	dto, err := gradebookToDTO(book)
	if err != nil {
		return err
	}
	dir := filepath.Dir(r.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	file, err := os.CreateTemp(dir, ".gradebook-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(dto); err != nil {
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
