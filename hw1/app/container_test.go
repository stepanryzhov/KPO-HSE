package app_test

import (
	"testing"

	"main/app"
	"main/garage"
	"main/inventory"
	"main/service"
	"main/vehicle"
)

type fakeService struct {
	reject  bool
	checked []service.ICheckable
}

func (s *fakeService) CheckIsBroken(item service.ICheckable) bool {
	s.checked = append(s.checked, item)
	return s.reject
}

var _ service.IService = (*fakeService)(nil)

func TestContainerUsesInjectedInspectionDecision(t *testing.T) {
	for _, tc := range []struct {
		name   string
		broken bool
		reject bool
	}{
		{name: "service rejects working vehicle", broken: false, reject: true},
		{name: "service accepts working vehicle", broken: false, reject: false},
		{name: "service overrides vehicle status", broken: true, reject: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			inspection := &fakeService{reject: tc.reject}
			container := app.NewWithService(inspection)
			bike, err := vehicle.NewBicycle(1, "BIKE-1", tc.broken)
			if err != nil {
				t.Fatal(err)
			}

			err = container.Garage.Add(bike)
			if (err != nil) != tc.reject {
				t.Fatalf("Add() вернул %v, отказ сервиса: %v", err, tc.reject)
			}
			if len(inspection.checked) != 1 || inspection.checked[0] != bike {
				t.Fatal("сервис должен проверить ровно один раз именно добавляемый транспорт")
			}

			items := container.Garage.AllItemsReport()
			if tc.reject {
				if len(items) != 0 {
					t.Fatal("отклонённый транспорт попал на баланс")
				}
			} else if len(items) != 1 || items[0] != bike {
				t.Fatal("принятый транспорт не появился на балансе")
			}
		})
	}
}

func TestContainerDoesNotInspectInventory(t *testing.T) {
	inspection := &fakeService{reject: true}
	container := app.NewWithService(inspection)
	helmet, err := inventory.NewHelmet(1, 56)
	if err != nil {
		t.Fatal(err)
	}
	station, err := inventory.NewChargeStation(2, 2.5)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range []garage.Item{helmet, station} {
		if err := container.Garage.Add(item); err != nil {
			t.Fatalf("вещь отклонена: %v", err)
		}
	}
	if len(inspection.checked) != 0 {
		t.Fatal("сервис техосмотра не должен вызываться для вещей")
	}
	if got := len(container.Garage.InventoryReport()); got != 2 {
		t.Errorf("вещей на балансе %d, ожидалось 2", got)
	}
}
