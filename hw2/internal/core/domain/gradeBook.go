package domain

import (
	"errors"
	"sort"
)

type Gradebook struct {
	students    map[studentID]*Student
	criteria    CriterionRegistry
	evaluations map[studentID]map[criterionID]Evaluation
}

func NewGradebook() *Gradebook {
	return &Gradebook{
		students:    make(map[studentID]*Student),
		criteria:    NewCriterionRegistry(),
		evaluations: make(map[studentID]map[criterionID]Evaluation),
	}
}

// RestoreGradebook создаёт пустую ведомость с сохранённым счётчиком критериев.
// Студенты, критерии и оценки восстанавливаются её доменными методами.
func RestoreGradebook(lastIssuedCriterionID int64) (*Gradebook, error) {
	lastID, err := NewID(lastIssuedCriterionID)
	if err != nil {
		return nil, err
	}
	book := NewGradebook()
	book.criteria.lastID = lastID
	return book, nil
}

func (g *Gradebook) addStudent(s *Student) error {
	if err := s.validate(); err != nil {
		return err
	}
	sid := s.GetID()
	if _, exists := g.students[sid]; exists {
		return errors.New("Студент с таким ID уже есть в ведомости")
	}
	if g.students == nil {
		g.students = make(map[studentID]*Student)
	}
	g.students[sid] = s
	return nil
}

func (g *Gradebook) CreateCriterion(name name, maxScore score, weight weight) (Criterion, error) {
	return g.criteria.CreateNewCriterion(name, maxScore, weight)
}

func (g *Gradebook) RestoreCriterion(rawID int64, rawName string, rawMaxScore, rawWeight float64) error {
	return g.criteria.RestoreCriterion(rawID, rawName, rawMaxScore, rawWeight)
}

func (g *Gradebook) Student(id studentID) (Student, error) {
	s, exists := g.students[id]
	if !exists {
		return Student{}, errors.New("Студент не найден в ведомости")
	}
	return *s, nil
}

func (g *Gradebook) Criterion(id criterionID) (Criterion, error) {
	return g.criteria.GetCriterionByID(id)
}

func (g *Gradebook) Students() []Student {
	students := make([]Student, 0, len(g.students))
	for _, student := range g.students {
		students = append(students, *student)
	}
	sort.Slice(students, func(i, j int) bool {
		return students[i].GetID().Value() < students[j].GetID().Value()
	})
	return students
}

func (g *Gradebook) Criteria() []Criterion {
	return g.criteria.Criteria()
}

func (g *Gradebook) LastCriterionID() int64 {
	return g.criteria.lastID.Value()
}

func (g *Gradebook) SetEvaluation(sid studentID, cid criterionID, score score) error {
	student, err := g.Student(sid)
	if err != nil {
		return err
	}
	criterion, err := g.Criterion(cid)
	if err != nil {
		return err
	}
	value, err := NewEvaluation(student, criterion, score)
	if err != nil {
		return err
	}

	if g.evaluations == nil {
		g.evaluations = make(map[studentID]map[criterionID]Evaluation)
	}
	if g.evaluations[sid] == nil {
		g.evaluations[sid] = make(map[criterionID]Evaluation)
	}
	g.evaluations[sid][cid] = value
	return nil
}

func (g *Gradebook) RemoveEvaluation(studentID studentID, criterionID criterionID) error {
	if _, err := g.Student(studentID); err != nil {
		return err
	}
	if _, err := g.Criterion(criterionID); err != nil {
		return err
	}
	if _, exists := g.evaluations[studentID][criterionID]; !exists {
		return errors.New("Оценка не найдена")
	}

	delete(g.evaluations[studentID], criterionID)
	if len(g.evaluations[studentID]) == 0 {
		delete(g.evaluations, studentID)
	}
	return nil
}

func (g *Gradebook) EvaluationsFor(studentID studentID) ([]Evaluation, error) {
	if _, err := g.Student(studentID); err != nil {
		return nil, err
	}

	values := make([]Evaluation, 0, len(g.evaluations[studentID]))
	for _, value := range g.evaluations[studentID] {
		values = append(values, value)
	}
	sort.Slice(values, func(i, j int) bool {
		return values[i].GetCriterionID().Value() < values[j].GetCriterionID().Value()
	})
	return values, nil
}
