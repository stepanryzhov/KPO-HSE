package main

import (
	"fmt"
	"path/filepath"

	"kpo_hw2/internal/core/domain"
	exportcsv "kpo_hw2/internal/features/export"
	gradebookrepo "kpo_hw2/internal/features/gradebooks/repository"
	gradebookservice "kpo_hw2/internal/features/gradebooks/service"
	studentrepo "kpo_hw2/internal/features/students/repository"
	studentservice "kpo_hw2/internal/features/students/service"
)

type Container struct {
	Students   *studentservice.Service
	Gradebooks *gradebookservice.Service
	Exporter   ReportExporter

	registry      *domain.StudentRegistry
	catalog       *gradebookrepo.Catalog
	gradebookName string
}

// NewContainer получает корневую папку данных для изоляции тестов и демо.
func NewContainer(dataDirectory string) (*Container, error) {
	studentsRepository := studentrepo.NewJSONRepository(filepath.Join(dataDirectory, "students.json"))
	registry, err := studentsRepository.Load()
	if err != nil {
		return nil, fmt.Errorf("загрузка студентов: %w", err)
	}
	studentService, err := studentservice.NewService(studentsRepository, registry)
	if err != nil {
		return nil, err
	}

	container := &Container{
		Students: studentService,
		Exporter: exportcsv.CSV{},
		registry: registry,
		catalog:  gradebookrepo.NewCatalog(filepath.Join(dataDirectory, "gradebooks")),
	}
	if err := container.OpenGradebook("default"); err != nil {
		return nil, err
	}
	return container, nil
}

func (c *Container) ListGradebooks() ([]string, error) {
	return c.catalog.Names()
}

func (c *Container) CreateGradebook(name string) error {
	if err := c.catalog.Create(name); err != nil {
		return err
	}
	return c.OpenGradebook(name)
}

func (c *Container) OpenGradebook(name string) error {
	gradebookRepository, err := c.catalog.Repository(name)
	if err != nil {
		return err
	}
	book, err := gradebookRepository.Load(*c.registry)
	if err != nil {
		return fmt.Errorf("загрузка ведомости: %w", err)
	}
	gradebookService, err := gradebookservice.NewService(gradebookRepository, c.registry, book)
	if err != nil {
		return err
	}

	c.Gradebooks = gradebookService
	c.gradebookName = name
	return nil
}
