package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"kpo_hw2/internal/core/apperrors"
)

func main() {
	if err := run(); err != nil {
		var saveError *apperrors.SaveError
		if errors.As(err, &saveError) {
			fmt.Fprintln(os.Stderr, "Не удалось сохранить изменения. Последняя операция не сохранена.")
			fmt.Fprintln(os.Stderr, "Программа завершена. Проверьте доступ к папке данных и свободное место на диске.")
			fmt.Fprintln(os.Stderr, "Причина:", saveError.Err)
		} else {
			fmt.Fprintln(os.Stderr, "Ошибка:", err)
		}
		os.Exit(1)
	}
}

func run() error {
	demo := flag.Bool("demo", false, "открыть готовую демонстрационную ведомость")
	flag.Parse()

	dataDirectory := "data"
	if *demo {
		directory, err := prepareDemo()
		if err != nil {
			return err
		}
		defer os.RemoveAll(directory)
		dataDirectory = directory
		fmt.Fprintln(os.Stdout, "Демо-режим: изменения JSON действуют только до выхода.")
	}

	container, err := NewContainer(dataDirectory)
	if err != nil {
		return err
	}
	return NewCLI(container, os.Stdin, os.Stdout).Run()
}
