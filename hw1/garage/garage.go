package garage

import (
	"errors"
	"main/service"
	"main/vehicle"
)

type Garage struct {
	items   map[int64]Item
	service service.IService
}

func NewGarage(s service.IService) *Garage {
	return &Garage{
		items:   make(map[int64]Item),
		service: s,
	}
}

func (g *Garage) Add(item Item) error {
	_, ok := g.items[item.GetID()]
	if ok {
		return errors.New("Объект с таким id уже есть")
	}
	transport, ok := item.(vehicle.IVehicle)
	if ok {
		if g.service.CheckIsBroken(transport) {
			return errors.New("Транспорт сломан")
		}
	}
	g.items[item.GetID()] = item
	return nil
}

func (g *Garage) VehiclesReport() []vehicle.IVehicle {
	result := make([]vehicle.IVehicle, 0)

	for _, val := range g.items {
		transport, ok := val.(vehicle.IVehicle)
		if ok {
			result = append(result, transport)
		}
	}
	return result
}

func (g *Garage) VehiclesForBeginners() []vehicle.IVehicle {
	result := make([]vehicle.IVehicle, 0)

	for _, val := range g.items {
		transport, ok := val.(vehicle.IVehicle)

		if ok {
			if transport.GetEaseOfUse() >= 6 {
				result = append(result, transport)
			}
		}
	}
	return result
}

func (g *Garage) DailyEnergyConsumptionReport() float32 {
	var sm float32
	sm = 0
	for _, val := range g.items {
		econ, ok := val.(IElectroConsumer)
		if ok {
			sm += econ.GetEnergyConsumption()
		}
	}
	return sm
}

// InventoryReport возвращает вещи на балансе, исключая транспорт.
func (g *Garage) InventoryReport() []Item {
	result := make([]Item, 0)
	for _, val := range g.items {
		if _, isVehicle := val.(vehicle.IVehicle); !isVehicle {
			result = append(result, val)
		}
	}
	return result
}

func (g *Garage) AllItemsReport() []Item {
	result := make([]Item, 0)
	for _, val := range g.items {
		result = append(result, val)
	}
	return result
}
