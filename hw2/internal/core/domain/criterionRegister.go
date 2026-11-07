package domain

import (
	"errors"
	"fmt"
	"math"
	"sort"
)

// CriterionRegistry хранит рубрикатор одной ведомости и выдаёт локальные ID.
type CriterionRegistry struct {
	criteria map[criterionID]*Criterion
	lastID   id
}

func NewCriterionRegistry() CriterionRegistry {
	id, err := NewID(0)
	if err != nil {
		return CriterionRegistry{}
	}
	return CriterionRegistry{
		criteria: make(map[criterionID]*Criterion),
		lastID:   id,
	}
}

func (r *CriterionRegistry) CreateNewCriterion(name name, maxScore score, weight weight) (Criterion, error) {
	if r.lastID.Value() == math.MaxInt64 {
		return Criterion{}, errors.New("Невозможно выдать новый ID критерия: номера закончились")
	}

	nextID, err := NewID(r.lastID.Value() + 1)
	if err != nil {
		return Criterion{}, err
	}
	createdCriterion, err := newCriterion(nextID, name, maxScore, weight)
	if err != nil {
		return Criterion{}, err
	}

	if r.criteria == nil {
		r.criteria = make(map[criterionID]*Criterion)
	}
	r.criteria[nextID] = createdCriterion
	r.lastID = id(nextID)
	return *createdCriterion, nil
}

func (r *CriterionRegistry) GetCriterionByID(id criterionID) (Criterion, error) {
	c, exists := r.criteria[id]
	if !exists {
		return Criterion{}, errors.New("Критерий не найден в реестре")
	}
	return *c, nil
}

func (r *CriterionRegistry) Criteria() []Criterion {
	criteria := make([]Criterion, 0, len(r.criteria))
	for _, c := range r.criteria {
		criteria = append(criteria, *c)
	}
	sort.Slice(criteria, func(i, j int) bool {
		return criteria[i].GetID().Value() < criteria[j].GetID().Value()
	})
	return criteria
}

// RestoreCriterion восстанавливает критерий без выдачи нового ID.
func (r *CriterionRegistry) RestoreCriterion(rawID int64, rawName string, rawMaxScore, rawWeight float64) error {
	if rawID <= 0 || rawID > r.lastID.Value() {
		return fmt.Errorf("ID критерия %d не согласован со счётчиком %d", rawID, r.lastID.Value())
	}
	identifier, err := NewID(rawID)
	if err != nil {
		return err
	}
	if _, exists := r.criteria[identifier]; exists {
		return fmt.Errorf("повтор ID критерия %d", rawID)
	}
	criterionName, err := NewName(rawName)
	if err != nil {
		return err
	}
	maxScore, err := NewScore(rawMaxScore)
	if err != nil {
		return err
	}
	criterionWeight, err := NewWeight(rawWeight)
	if err != nil {
		return err
	}
	criterion, err := newCriterion(identifier, criterionName, maxScore, criterionWeight)
	if err != nil {
		return err
	}
	if r.criteria == nil {
		r.criteria = make(map[criterionID]*Criterion)
	}
	r.criteria[identifier] = criterion
	return nil
}
