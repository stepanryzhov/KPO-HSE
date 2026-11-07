package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"

	"kpo_hw2/internal/core/apperrors"
	"kpo_hw2/internal/core/domain"
	gradebookservice "kpo_hw2/internal/features/gradebooks/service"
)

// ReportExporter задаёт контракт экспорта, нужный консоли.
type ReportExporter interface {
	Write(io.Writer, gradebookservice.Report) error
}

type CLI struct {
	container *Container
	input     *bufio.Scanner
	output    io.Writer
}

// command связывает пункт меню с выполняемым действием.
type command struct {
	name    string
	execute func() error
}

var errCanceled = errors.New("операция отменена")

func NewCLI(container *Container, input io.Reader, output io.Writer) *CLI {
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 1024), 1024*1024)
	return &CLI{container: container, input: scanner, output: output}
}

func (c *CLI) Run() error {
	commands := []command{
		{"Добавить студента", c.addStudent},
		{"Все студенты", c.listStudents},
		{"Добавить студента в ведомость", c.enrollStudent},
		{"Студенты ведомости", c.listEnrolledStudents},
		{"Добавить критерий", c.createCriterion},
		{"Критерии ведомости", c.listCriteria},
		{"Выставить или исправить оценку", c.setEvaluation},
		{"Оценки студента", c.listEvaluations},
		{"Отменить оценку", c.removeEvaluation},
		{"Итог студента", c.calculateFinalMark},
		{"Полная ведомость и итоги", c.showReport},
		{"Экспорт ведомости в CSV", c.exportCSV},
		{"Выбрать или создать ведомость", c.openGradebook},
	}
	for {
		fmt.Fprintf(c.output, "\nПОС — ведомость %s\n", c.container.gradebookName)
		for i, command := range commands {
			fmt.Fprintf(c.output, "%2d. %s\n", i+1, command.name)
		}
		fmt.Fprintln(c.output, " 0. Выход")

		value, err := c.readLine("Команда: ")
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		number, err := strconv.Atoi(value)
		if err != nil || number < 0 || number > len(commands) {
			fmt.Fprintln(c.output, "Неизвестная команда. Введите номер из меню.")
			continue
		}
		if number == 0 {
			fmt.Fprintln(c.output, "До свидания!")
			return nil
		}
		if err := commands[number-1].execute(); err != nil {
			var saveError *apperrors.SaveError
			if errors.As(err, &saveError) {
				return err
			}
			if errors.Is(err, io.EOF) {
				return nil
			}
			if errors.Is(err, errCanceled) {
				continue
			}
			fmt.Fprintln(c.output, "Ошибка:", err)
		}
	}
}

func (c *CLI) addStudent() error {
	name, err := c.readLine("Имя студента: ")
	if err != nil {
		return err
	}
	if err := c.container.Students.AddStudent(name); err != nil {
		return err
	}
	fmt.Fprintln(c.output, "Студент добавлен.")
	return c.listStudents()
}

func (c *CLI) listStudents() error {
	return c.printStudents(c.container.Students.ListStudents())
}

func (c *CLI) listEnrolledStudents() error {
	return c.printStudents(c.container.Gradebooks.ListStudents())
}

func (c *CLI) printStudents(students []domain.Student) error {
	if len(students) == 0 {
		fmt.Fprintln(c.output, "Студентов пока нет.")
		return nil
	}
	rows := [][]string{{"ID", "Имя"}}
	for _, student := range students {
		rows = append(rows, []string{
			strconv.FormatInt(student.GetID().Value(), 10),
			student.GetName().Value(),
		})
	}
	return c.printTable(rows)
}

func (c *CLI) enrollStudent() error {
	enrolled := make(map[int64]bool)
	for _, student := range c.container.Gradebooks.ListStudents() {
		enrolled[student.GetID().Value()] = true
	}
	available := make([]domain.Student, 0)
	for _, student := range c.container.Students.ListStudents() {
		if !enrolled[student.GetID().Value()] {
			available = append(available, student)
		}
	}
	fmt.Fprintln(c.output, "Студенты, которых можно добавить в ведомость:")
	student, err := c.selectStudent(available, "Нет студентов для зачисления. Добавьте нового студента командой 1.")
	if err != nil {
		return err
	}
	if err := c.container.Gradebooks.EnrollStudent(student.GetID().Value()); err != nil {
		return err
	}
	fmt.Fprintln(c.output, "Студент добавлен в ведомость.")
	return nil
}

func (c *CLI) createCriterion() error {
	name, err := c.readLine("Название критерия: ")
	if err != nil {
		return err
	}
	maxScore, err := c.readNumber("Максимальный балл: ")
	if err != nil {
		return err
	}
	weight, err := c.readNumber("Вес: ")
	if err != nil {
		return err
	}
	criterion, err := c.container.Gradebooks.CreateCriterion(name, maxScore, weight)
	if err != nil {
		return err
	}
	fmt.Fprintf(c.output, "Критерий #%d «%s» добавлен.\n", criterion.GetID().Value(), criterion.GetName().Value())
	return nil
}

func (c *CLI) listCriteria() error {
	criteria := c.container.Gradebooks.ListCriteria()
	if len(criteria) == 0 {
		fmt.Fprintln(c.output, "Критериев пока нет.")
		return nil
	}
	rows := [][]string{{"ID", "Критерий", "Максимум", "Вес"}}
	for _, criterion := range criteria {
		rows = append(rows, []string{
			strconv.FormatInt(criterion.GetID().Value(), 10),
			criterion.GetName().Value(),
			formatNumber(criterion.GetMaxScore().Value()),
			formatNumber(criterion.GetWeight().Value()),
		})
	}
	return c.printTable(rows)
}

func (c *CLI) setEvaluation() error {
	student, err := c.selectEnrolledStudent()
	if err != nil {
		return err
	}
	criterion, err := c.selectCriterion(student, false)
	if err != nil {
		return err
	}
	score, err := c.readNumber(fmt.Sprintf("Балл по «%s» (0..%s): ", criterion.GetName().Value(), formatNumber(criterion.GetMaxScore().Value())))
	if err != nil {
		return err
	}
	if err := c.container.Gradebooks.SetEvaluation(student.GetID().Value(), criterion.GetID().Value(), score); err != nil {
		return err
	}
	fmt.Fprintln(c.output, "Оценка сохранена.")
	return nil
}

func (c *CLI) listEvaluations() error {
	student, err := c.selectEnrolledStudent()
	if err != nil {
		return err
	}
	evaluations, err := c.container.Gradebooks.GetEvaluations(student.GetID().Value())
	if err != nil {
		return err
	}
	fmt.Fprintf(c.output, "Оценки: %s (#%d)\n", student.GetName().Value(), student.GetID().Value())
	if len(evaluations) == 0 {
		fmt.Fprintln(c.output, "Оценок пока нет.")
		return nil
	}
	rows := [][]string{{"ID критерия", "Критерий", "Балл", "Максимум", "Вес"}}
	for _, evaluation := range evaluations {
		criterion, err := c.container.Gradebooks.GetCriterion(evaluation.GetCriterionID().Value())
		if err != nil {
			return err
		}
		rows = append(rows, []string{
			strconv.FormatInt(criterion.GetID().Value(), 10),
			criterion.GetName().Value(),
			formatNumber(evaluation.GetScore().Value()),
			formatNumber(criterion.GetMaxScore().Value()),
			formatNumber(criterion.GetWeight().Value()),
		})
	}
	return c.printTable(rows)
}

func (c *CLI) removeEvaluation() error {
	student, err := c.selectEnrolledStudent()
	if err != nil {
		return err
	}
	criterion, err := c.selectCriterion(student, true)
	if err != nil {
		return err
	}
	if err := c.container.Gradebooks.RemoveEvaluation(student.GetID().Value(), criterion.GetID().Value()); err != nil {
		return err
	}
	fmt.Fprintln(c.output, "Оценка отменена.")
	return nil
}

func (c *CLI) calculateFinalMark() error {
	student, err := c.selectEnrolledStudent()
	if err != nil {
		return err
	}
	mark, err := c.container.Gradebooks.CalculateFinalMark(student.GetID().Value())
	if err != nil {
		return err
	}
	fmt.Fprintf(c.output, "Итог студента #%d: %s\n", student.GetID().Value(), formatNumber(mark))
	return nil
}

func (c *CLI) showReport() error {
	report, err := c.container.Gradebooks.GetReport()
	if err != nil {
		return err
	}
	if len(report.Rows) == 0 {
		fmt.Fprintln(c.output, "Ведомость пока пуста.")
		return nil
	}
	header := []string{"Имя"}
	for _, criterion := range report.Criteria {
		header = append(header, criterion.GetName().Value())
	}
	rows := [][]string{append(header, "Итог")}
	for _, result := range report.Rows {
		row := []string{result.Student.GetName().Value()}
		for _, criterion := range report.Criteria {
			row = append(row, formatNumber(result.Scores[criterion.GetID().Value()]))
		}
		rows = append(rows, append(row, formatNumber(result.FinalMark)))
	}
	return c.printTable(rows)
}

func (c *CLI) exportCSV() error {
	path, err := c.readLine("Путь к CSV [results.csv]: ")
	if err != nil {
		return err
	}
	if path == "" {
		path = "results.csv"
	}
	if !strings.EqualFold(filepath.Ext(path), ".csv") {
		return errors.New("для экспорта укажите файл с расширением .csv")
	}
	report, err := c.container.Gradebooks.GetReport()
	if err != nil {
		return err
	}
	info, err := os.Stat(path)
	if err == nil {
		if info.IsDir() {
			return errors.New("укажите имя CSV-файла, а не папку")
		}
		confirmed, err := c.confirm(fmt.Sprintf("Файл %q уже существует. Перезаписать? [д/Н]: ", path))
		if err != nil {
			return err
		}
		if !confirmed {
			fmt.Fprintln(c.output, "Экспорт отменён.")
			return nil
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	writeErr := c.container.Exporter.Write(file, report)
	closeErr := file.Close()
	if writeErr != nil {
		return writeErr
	}
	if closeErr != nil {
		return closeErr
	}
	fmt.Fprintf(c.output, "CSV сохранён: %s\n", path)
	return nil
}

func (c *CLI) openGradebook() error {
	names, err := c.container.ListGradebooks()
	if err != nil {
		return err
	}
	if len(names) == 0 {
		fmt.Fprintln(c.output, "Сохранённых ведомостей пока нет.")
	}
	for i, name := range names {
		label := name
		if name == c.container.gradebookName {
			label += " (текущая)"
		}
		fmt.Fprintf(c.output, "%d. %s\n", i+1, label)
	}
	fmt.Fprintf(c.output, "%d. Создать новую ведомость\n0. Назад\n", len(names)+1)
	for {
		choice, err := c.readID("Номер ведомости: ")
		if err != nil {
			return err
		}
		if choice == 0 {
			return errCanceled
		}
		if choice > 0 && choice <= int64(len(names)) {
			if err := c.container.OpenGradebook(names[choice-1]); err != nil {
				return err
			}
			fmt.Fprintf(c.output, "Открыта ведомость: %s\n", c.container.gradebookName)
			return nil
		}
		if choice == int64(len(names)+1) {
			name, err := c.readLine("Название новой ведомости (пусто — назад): ")
			if err != nil {
				return err
			}
			if name == "" {
				return errCanceled
			}
			if err := c.container.CreateGradebook(name); err != nil {
				return err
			}
			fmt.Fprintf(c.output, "Создана ведомость: %s\n", c.container.gradebookName)
			return nil
		}
		fmt.Fprintln(c.output, "Выберите номер из списка.")
	}
}

func (c *CLI) readLine(prompt string) (string, error) {
	if _, err := fmt.Fprint(c.output, prompt); err != nil {
		return "", err
	}
	if !c.input.Scan() {
		if err := c.input.Err(); err != nil {
			return "", err
		}
		return "", io.EOF
	}
	return strings.TrimSpace(c.input.Text()), nil
}

func (c *CLI) readID(prompt string) (int64, error) {
	for {
		value, err := c.readLine(prompt)
		if err != nil {
			return 0, err
		}
		id, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			fmt.Fprintln(c.output, "Введите целое число для ID.")
			continue
		}
		return id, nil
	}
}

func (c *CLI) readNumber(prompt string) (float64, error) {
	for {
		value, err := c.readLine(prompt)
		if err != nil {
			return 0, err
		}
		number, err := strconv.ParseFloat(strings.ReplaceAll(value, ",", "."), 64)
		if err != nil {
			fmt.Fprintln(c.output, "Введите число, например 8.5.")
			continue
		}
		return number, nil
	}
}

func (c *CLI) selectEnrolledStudent() (domain.Student, error) {
	return c.selectStudent(c.container.Gradebooks.ListStudents(), "В ведомости нет студентов. Добавьте их командой 3.")
}

func (c *CLI) selectStudent(students []domain.Student, emptyMessage string) (domain.Student, error) {
	if len(students) == 0 {
		fmt.Fprintln(c.output, emptyMessage)
		return domain.Student{}, errCanceled
	}
	if err := c.printStudents(students); err != nil {
		return domain.Student{}, err
	}
	for {
		id, err := c.readID("ID студента (0 — назад): ")
		if err != nil {
			return domain.Student{}, err
		}
		if id == 0 {
			return domain.Student{}, errCanceled
		}
		for _, student := range students {
			if student.GetID().Value() == id {
				return student, nil
			}
		}
		fmt.Fprintln(c.output, "Выберите ID студента из списка.")
	}
}

func (c *CLI) selectCriterion(student domain.Student, onlyGraded bool) (domain.Criterion, error) {
	criteria := c.container.Gradebooks.ListCriteria()
	if len(criteria) == 0 {
		fmt.Fprintln(c.output, "Критериев пока нет. Добавьте их командой 5.")
		return domain.Criterion{}, errCanceled
	}
	evaluations, err := c.container.Gradebooks.GetEvaluations(student.GetID().Value())
	if err != nil {
		return domain.Criterion{}, err
	}
	scores := make(map[int64]float64, len(evaluations))
	for _, evaluation := range evaluations {
		scores[evaluation.GetCriterionID().Value()] = evaluation.GetScore().Value()
	}
	available := make([]domain.Criterion, 0, len(criteria))
	rows := [][]string{{"ID", "Критерий", "Текущий балл", "Максимум", "Вес"}}
	for _, criterion := range criteria {
		score, exists := scores[criterion.GetID().Value()]
		if onlyGraded && !exists {
			continue
		}
		value := "—"
		if exists {
			value = formatNumber(score)
		}
		available = append(available, criterion)
		rows = append(rows, []string{
			strconv.FormatInt(criterion.GetID().Value(), 10),
			criterion.GetName().Value(),
			value,
			formatNumber(criterion.GetMaxScore().Value()),
			formatNumber(criterion.GetWeight().Value()),
		})
	}
	if len(available) == 0 {
		fmt.Fprintln(c.output, "У студента пока нет оценок для отмены.")
		return domain.Criterion{}, errCanceled
	}
	fmt.Fprintf(c.output, "Оценки: %s (#%d)\n", student.GetName().Value(), student.GetID().Value())
	if err := c.printTable(rows); err != nil {
		return domain.Criterion{}, err
	}
	for {
		id, err := c.readID("ID критерия (0 — назад): ")
		if err != nil {
			return domain.Criterion{}, err
		}
		if id == 0 {
			return domain.Criterion{}, errCanceled
		}
		for _, criterion := range available {
			if criterion.GetID().Value() == id {
				return criterion, nil
			}
		}
		fmt.Fprintln(c.output, "Выберите ID критерия из списка.")
	}
}

func (c *CLI) confirm(prompt string) (bool, error) {
	answer, err := c.readLine(prompt)
	if err != nil {
		return false, err
	}
	switch strings.ToLower(answer) {
	case "д", "да", "y", "yes":
		return true, nil
	default:
		return false, nil
	}
}

func (c *CLI) printTable(rows [][]string) error {
	table := tabwriter.NewWriter(c.output, 0, 4, 2, ' ', 0)
	for _, row := range rows {
		if _, err := fmt.Fprintln(table, strings.Join(row, "\t")); err != nil {
			return err
		}
	}
	return table.Flush()
}

func formatNumber(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}
