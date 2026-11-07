package domain

import "errors"

type id struct {
	value int64
}

func NewID(value int64) (id, error) {
	if value < 0 {
		return id{}, errors.New("ID должен быть неотрицалельным")
	}

	return id{value: value}, nil
}

func (id id) Value() int64 {
	return id.value
}
