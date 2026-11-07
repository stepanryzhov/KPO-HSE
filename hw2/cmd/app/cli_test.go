package main

import (
	"bytes"
	"encoding/csv"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestCLIPresentation(t *testing.T) {
	dir := t.TempDir()
	container, err := NewContainer(dir)
	if err != nil {
		t.Fatal(err)
	}
	script, err := os.ReadFile("demo/presentation.txt")
	if err != nil {
		t.Fatal(err)
	}
	csvFile := filepath.Join(dir, "results.csv")
	input := strings.ReplaceAll(string(script), "demo-results.csv", csvFile)
	var output bytes.Buffer
	if err := NewCLI(container, strings.NewReader(input), &output).Run(); err != nil {
		t.Fatal(err)
	}
	if strings.Count(output.String(), "Ошибка:") != 1 || strings.Contains(output.String(), "Неизвестная команда") {
		t.Fatalf("сценарий должен содержать ровно один отказ для балла выше максимума:\n%s", output.String())
	}
	for _, text := range []string{"не должен превышать", "Оценка отменена.", "Итог студента #1: 9.75", "CSV сохранён:", "До свидания!"} {
		if !strings.Contains(output.String(), text) {
			t.Errorf("в выводе нет %q", text)
		}
	}
	data, err := os.ReadFile(csvFile)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := csv.NewReader(bytes.NewReader(data)).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	want := [][]string{
		{"Имя", "Домашние работы", "Проект", "Экзамен", "Итог"},
		{"Анна Смирнова", "10", "9", "10", "9.75"},
		{"Борис Иванов", "7", "8", "9", "8.25"},
		{"Вера Петрова", "10", "9", "8", "8.75"},
	}
	if !reflect.DeepEqual(rows, want) {
		t.Fatalf("CSV = %v, ожидается %v", rows, want)
	}
	report, err := container.Gradebooks.GetReport()
	if err != nil {
		t.Fatal(err)
	}
	reloaded, err := NewContainer(dir)
	if err != nil {
		t.Fatal(err)
	}
	reloadedReport, err := reloaded.Gradebooks.GetReport()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(report, reloadedReport) {
		t.Fatal("результат CLI не восстановился из JSON")
	}
}
