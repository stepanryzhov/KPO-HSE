package inventory

import "errors"

type inventoryItem struct {
	id   int64
	name string
}

func NewInventoryItem(id int64, name string) (*inventoryItem, error) {
	if id < 1 {
		return nil, errors.New("id должен быть натуральным числом")
	}
	return &inventoryItem{id, name}, nil
}

func (i inventoryItem) GetID() int64 {
	return i.id
}

func (i inventoryItem) GetName() string {
	return i.name
}
