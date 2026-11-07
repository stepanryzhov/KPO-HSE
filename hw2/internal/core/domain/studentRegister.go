package domain

import (
	"errors"
	"fmt"
	"math"
	"sort"
)

type StudentRegistry struct {
	students map[studentID]*Student
	lastID   id
}

func NewStudentRegistry(lastIssuedID int64) (*StudentRegistry, error) {
	lastID, err := NewID(lastIssuedID)
	if err != nil {
		return nil, err
	}
	return &StudentRegistry{
		students: make(map[studentID]*Student),
		lastID:   lastID,
	}, nil
}

func (r *StudentRegistry) CreateNewStudent(name name) (Student, error) {

	if r.lastID.Value() == math.MaxInt64 {
		return Student{}, errors.New("Невозможно выдать новый ID студента: номера закончились")
	}

	nextID, err := NewID(r.lastID.Value() + 1)

	if err != nil {
		return Student{}, err
	}
	createdStudent, err := newStudent(nextID, name)
	if err != nil {
		return Student{}, err
	}

	if r.students == nil {
		r.students = make(map[studentID]*Student)
	}

	r.students[nextID] = createdStudent
	r.lastID = nextID
	return *createdStudent, nil
}

func (r *StudentRegistry) GetStudentByID(id studentID) (Student, error) {
	s, exists := r.students[id]
	if !exists {
		return Student{}, errors.New("Студент не найден в реестре")
	}
	return *s, nil
}

func (r *StudentRegistry) Students() []Student {
	students := make([]Student, 0, len(r.students))
	for _, s := range r.students {
		students = append(students, *s)
	}
	sort.Slice(students, func(i, j int) bool {
		return students[i].GetID().Value() < students[j].GetID().Value()
	})
	return students
}

func (r *StudentRegistry) AddToGradebook(book *Gradebook, id studentID) error {
	if book == nil {
		return errors.New("ведомость не задана")
	}
	s, exists := r.students[id]
	if !exists {
		return errors.New("Студент не найден в реестре")
	}
	return book.addStudent(s)
}

func (r *StudentRegistry) LastID() int64 {
	return r.lastID.Value()
}

// RestoreStudent восстанавливает сохранённую запись без выдачи нового ID.
func (r *StudentRegistry) RestoreStudent(rawID int64, rawName string) error {
	if rawID <= 0 || rawID > r.LastID() {
		return fmt.Errorf("ID студента %d не согласован со счётчиком %d", rawID, r.LastID())
	}
	identifier, err := NewID(rawID)
	if err != nil {
		return err
	}
	if _, exists := r.students[identifier]; exists {
		return fmt.Errorf("повтор ID студента %d", rawID)
	}
	studentName, err := NewName(rawName)
	if err != nil {
		return err
	}
	s, err := newStudent(identifier, studentName)
	if err != nil {
		return err
	}
	if r.students == nil {
		r.students = make(map[studentID]*Student)
	}
	r.students[identifier] = s
	return nil
}
