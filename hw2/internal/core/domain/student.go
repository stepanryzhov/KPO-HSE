package domain

import "errors"

type studentID = id

type Student struct {
	id   studentID
	name name
}

func newStudent(id studentID, name name) (*Student, error) {
	s := &Student{id: id, name: name}
	if err := s.validate(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Student) validate() error {
	if s == nil {
		return errors.New("Студент не должен быть nil")
	}
	if s.id.Value() <= 0 {
		return errors.New("ID студента должен быть положительным")
	}
	_, err := NewName(s.name.Value())
	return err
}

func (s Student) GetID() studentID {
	return s.id
}

func (s Student) GetName() name {
	return s.name
}
