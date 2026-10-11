package inventory

import "errors"

type chargeStation struct {
	inventoryItem
	dailyEnergyConsumption float32
}

func NewChargeStation(id int64, dailyEnergyConsumption float32) (*chargeStation, error) {
	if dailyEnergyConsumption < 0 {
		return nil, errors.New("суточный расход энергии должен быть неотрицательным")
	}
	base, err := NewInventoryItem(id, "Charge Station")
	if err != nil {
		return nil, err
	}
	return &chargeStation{*base, dailyEnergyConsumption}, nil
}

func (s chargeStation) GetEnergyConsumption() float32 {
	return s.dailyEnergyConsumption
}
