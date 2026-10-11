package inventory

type helmet struct {
	inventoryItem
	size int
}

func NewHelmet(id int64, size int) (*helmet, error) {
	base, err := NewInventoryItem(id, "helmet")
	if err != nil {
		return nil, err
	}
	return &helmet{*base, size}, nil
}
