package vehicle

import (
	"errors"
	"strings"
)

type electricTransport struct {
	vehicle
	battery                int
	energyDailyConsumption float32
}

type vehicle struct {
	name      string
	id        int64
	easeOfUse int
	gosNumber string
	isBroken  bool
}

func (v vehicle) CheckIsBroken() bool {
	return v.isBroken
}

func (v vehicle) GetID() int64 {
	return v.id
}

func (v vehicle) GetGosNumber() string {
	return v.gosNumber
}

func (v vehicle) GetName() string {
	return v.name
}

func (v vehicle) GetEaseOfUse() int {
	return v.easeOfUse
}

func (e electricTransport) GetEnergyConsumption() float32 {
	return e.energyDailyConsumption
}

func newVehicle(name string, id int64, easeOfUse int, gosNumber string, isBroken bool) (*vehicle, error) {
	if id < 1 {
		return nil, errors.New("id должен быть натуральным числом")
	}
	if easeOfUse < 1 || easeOfUse > 10 {
		return nil, errors.New("простота должна быть от 1 до 10")
	}

	if strings.TrimSpace(name) == "" {
		return nil, errors.New("название транспорта не должно быть пустым")
	}
	return &vehicle{
		name:      name,
		id:        id,
		easeOfUse: easeOfUse,
		gosNumber: gosNumber,
		isBroken:  isBroken,
	}, nil
}

func newElectricTransport(name string, id int64, easeOfUse int, gosNumber string, isBroken bool, battery int, energyDailyConsumption float32) (*electricTransport, error) {
	if energyDailyConsumption < 0 {
		return nil, errors.New("Energy Daily Consumption должна быть неотрицательной")
	}
	if battery > 100 || battery < 0 {
		return nil, errors.New("заряд батереи должен быть от 0 до 100")
	}
	base, err := newVehicle(name, id, easeOfUse, gosNumber, isBroken)
	if err != nil {
		return nil, err
	}
	return &electricTransport{
		vehicle:                *base,
		battery:                battery,
		energyDailyConsumption: energyDailyConsumption,
	}, nil

}
