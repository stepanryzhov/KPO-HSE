package domain

import (
	"errors"
)

type criterionID = id

type Criterion struct {
	id       criterionID
	name     name
	maxScore score
	weight   weight
}

func newCriterion(id criterionID, name name, maxScore score, weight weight) (*Criterion, error) {
	c := &Criterion{id: id, name: name, maxScore: maxScore, weight: weight}
	if err := c.validate(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Criterion) validate() error {
	if c == nil {
		return errors.New("Критерий не должен быть nil")
	}
	if c.id.Value() <= 0 {
		return errors.New("ID критерия должен быть положительным")
	}
	if _, err := NewName(c.name.Value()); err != nil {
		return err
	}
	if _, err := NewScore(c.maxScore.Value()); err != nil {
		return err
	}
	if c.maxScore.Value() <= 0 {
		return errors.New("Максимальный балл критерия должен быть больше нуля")
	}
	_, err := NewWeight(c.weight.Value())
	return err
}

func (c Criterion) GetID() criterionID {
	return c.id
}

func (c Criterion) GetName() name {
	return c.name
}

func (c Criterion) GetMaxScore() score {
	return c.maxScore
}

func (c Criterion) GetWeight() weight {
	return c.weight
}

func (c Criterion) ValidateScore(value score) error {
	if _, err := NewScore(value.Value()); err != nil {
		return err
	}
	if value.Value() > c.maxScore.Value() {
		return errors.New("Балл не должен превышать максимальную оценку критерия")
	}
	return nil
}
