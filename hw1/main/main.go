package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"main/app"
	"main/fileinput"
	"main/garage"
)

func main() {
	container := app.New()
	err := run(os.Stdin, os.Stdout, container.Garage)
	if err != nil {
		log.Fatal(err)
	}
}

func run(r io.Reader, w io.Writer, fleet *garage.Garage) error {
	input := bufio.NewScanner(r)
	fmt.Fprintln(w, "Выберите режим запуска:")
	fmt.Fprintln(w, "1 — Демонстрация")
	fmt.Fprintln(w, "2 — Загрузка данных из JSON-файла")

	for {
		fmt.Fprint(w, "Ваш выбор: ")
		choice, err := readLine(input)
		if err != nil {
			return fmt.Errorf("прочитать режим запуска: %w", err)
		}
		switch choice {
		case "1":
			return runDemo(w, fleet)
		case "2":
			fmt.Fprint(w, "Путь к JSON-файлу: ")
			path, err := readLine(input)
			if err != nil {
				return fmt.Errorf("прочитать путь к файлу: %w", err)
			}
			return runFile(w, fleet, path)
		default:
			fmt.Fprintln(w, "Введите 1 или 2.")
		}
	}
}

func readLine(input *bufio.Scanner) (string, error) {
	if input.Scan() {
		return strings.TrimSpace(input.Text()), nil
	}
	if err := input.Err(); err != nil {
		return "", err
	}
	return "", io.EOF
}

func runFile(w io.Writer, fleet *garage.Garage, path string) error {
	items, err := fileinput.ReadFile(path)
	if err != nil {
		return err
	}
	for _, item := range items {
		if err := fleet.Add(item); err != nil {
			fmt.Fprintf(w, "Не принят в парк: %s (инвентарный номер %d): %v\n", item.GetName(), item.GetID(), err)
			continue
		}
		fmt.Fprintf(w, "Принят в парк: %s (инвентарный номер %d)\n", item.GetName(), item.GetID())
	}
	fmt.Fprintln(w, "\nИТОГОВЫЕ ОТЧЁТЫ")
	printConsoleReports(w, fleet)
	return nil
}
