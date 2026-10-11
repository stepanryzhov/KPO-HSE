package fileinput

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"main/garage"
	"main/inventory"
	"main/vehicle"
)

type itemData struct {
	Type                   string  `json:"type"`
	ID                     int64   `json:"id"`
	GosNumber              string  `json:"gos_number"`
	IsBroken               bool    `json:"is_broken"`
	Battery                int     `json:"battery"`
	Size                   int     `json:"size"`
	DailyEnergyConsumption float32 `json:"daily_energy_consumption"`
}

func ReadFile(path string) ([]garage.Item, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("открыть файл %q: %w", path, err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	var data []itemData
	if err := decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("прочитать JSON из файла %q: %w", path, err)
	}
	if data == nil {
		return nil, fmt.Errorf("файл %q должен содержать JSON-массив объектов", path)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("в файле %q должен быть только один JSON-массив", path)
		}
		return nil, fmt.Errorf("лишние данные после JSON-массива в файле %q: %w", path, err)
	}

	items := make([]garage.Item, 0, len(data))
	for i, record := range data {
		item, err := record.toItem()
		if err != nil {
			return nil, fmt.Errorf("файл %q, объект %d: %w", path, i+1, err)
		}
		items = append(items, item)
	}
	return items, nil
}

func (data itemData) toItem() (garage.Item, error) {
	switch data.Type {
	case "bicycle":
		return vehicle.NewBicycle(data.ID, data.GosNumber, data.IsBroken)
	case "scooter":
		return vehicle.NewScooter(data.ID, data.GosNumber, data.IsBroken, data.Battery)
	case "electric_bicycle":
		return vehicle.NewElectricBicycle(data.ID, data.GosNumber, data.IsBroken, data.Battery)
	case "helmet":
		return inventory.NewHelmet(data.ID, data.Size)
	case "charge_station":
		return inventory.NewChargeStation(data.ID, data.DailyEnergyConsumption)
	default:
		return nil, fmt.Errorf("неизвестный тип имущества %q", data.Type)
	}
}
