package domain

import (
	"errors"
	"math"
)

type weight struct {
	value float64
}

func NewWeight(value float64) (weight, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return weight{}, errors.New("Вес критерия должен быть конечным числом")
	}
	if value <= 0 {
		return weight{}, errors.New("Вес критерия должен быть строго больше нуля")
	}
	return weight{value: value}, nil
}

func (w weight) Value() float64 {
	return w.value
}
