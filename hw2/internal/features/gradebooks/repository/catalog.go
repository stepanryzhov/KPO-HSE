package repository

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"kpo_hw2/internal/core/apperrors"
	"kpo_hw2/internal/core/domain"
)

// Catalog хранит именованные ведомости в одной папке.
type Catalog struct {
	directory string
}

func NewCatalog(directory string) *Catalog {
	return &Catalog{directory: directory}
}

func validateName(name string) error {
	if name == "" || strings.TrimSpace(name) != name {
		return errors.New("название ведомости не должно быть пустым или начинаться и заканчиваться пробелами")
	}
	for _, char := range name {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != ' ' && char != '-' && char != '_' {
			return errors.New("в названии ведомости допустимы буквы, цифры, пробелы, дефис и подчёркивание")
		}
	}
	return nil
}

func (c *Catalog) Repository(name string) (*JSONRepository, error) {
	if err := validateName(name); err != nil {
		return nil, err
	}
	path := filepath.Join(c.directory, name+".json")
	info, err := os.Lstat(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	if info != nil && !info.Mode().IsRegular() {
		return nil, errors.New("ведомость должна быть обычным JSON-файлом")
	}
	return NewJSONRepository(path), nil
}

func (c *Catalog) Names() ([]string, error) {
	entries, err := os.ReadDir(c.directory)
	if errors.Is(err, os.ErrNotExist) {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.Type().IsRegular() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".json")
		if validateName(name) == nil {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names, nil
}

func (c *Catalog) Create(name string) error {
	repository, err := c.Repository(name)
	if err != nil {
		return err
	}
	names, err := c.Names()
	if err != nil {
		return err
	}
	for _, existing := range names {
		if strings.EqualFold(existing, name) {
			return fmt.Errorf("ведомость %q уже существует — выберите её из списка", existing)
		}
	}
	if _, err := os.Lstat(repository.path); err == nil {
		return fmt.Errorf("файл ведомости %q уже существует", name)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := repository.Save(*domain.NewGradebook()); err != nil {
		return &apperrors.SaveError{Err: err}
	}
	return nil
}
