package domain

import (
	"errors"
	"math"
)

type score struct {
	value float64
}

func NewScore(value float64) (score, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return score{}, errors.New("оценка должна быть конечным числом")
	}
	if value < 0 {
		return score{}, errors.New("оценка должна быть неотрицательной")
	}
	return score{value: value}, nil
}

func (ms score) Value() float64 {
	return ms.value
}
