package domain

import (
	"errors"
	"strings"
)

type name struct {
	value string
}

func NewName(value string) (name, error) {
	if strings.TrimSpace(value) == "" {
		return name{}, errors.New("имя не должно быть пустым")
	}

	return name{value: value}, nil
}

func (name name) Value() string {
	return name.value
}
