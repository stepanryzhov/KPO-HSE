package garage

import (
	"testing"

	"main/inventory"
	"main/service"
	"main/vehicle"
)

func newTestGarage() *Garage {
	return NewGarage(service.NewService())
}

type fakeVehicle struct {
	id        int64
	name      string
	ease      int
	gosNumber string
	broken    bool
}

func (v fakeVehicle) Ride()                {}
func (v fakeVehicle) CheckIsBroken() bool  { return v.broken }
func (v fakeVehicle) GetID() int64         { return v.id }
func (v fakeVehicle) GetEaseOfUse() int    { return v.ease }
func (v fakeVehicle) GetGosNumber() string { return v.gosNumber }
func (v fakeVehicle) GetName() string      { return v.name }

type fakeEnergyItem struct {
	id          int64
	name        string
	consumption float32
}

func (i fakeEnergyItem) GetID() int64                  { return i.id }
func (i fakeEnergyItem) GetName() string               { return i.name }
func (i fakeEnergyItem) GetEnergyConsumption() float32 { return i.consumption }

type plainItem struct {
	id   int64
	name string
}

func (i plainItem) GetID() int64    { return i.id }
func (i plainItem) GetName() string { return i.name }

func TestAddAcceptsWorkingVehicle(t *testing.T) {
	g := newTestGarage()
	bike := fakeVehicle{id: 1, name: "bicycle", gosNumber: "BIKE-1", ease: 7}

	if err := g.Add(bike); err != nil {
		t.Fatalf("исправный транспорт отклонён: %v", err)
	}
	if got := len(g.VehiclesReport()); got != 1 {
		t.Errorf("в отчёте %d транспортных средств, ожидалось 1", got)
	}
}

func TestAddRejectsBrokenVehicle(t *testing.T) {
	g := newTestGarage()
	bike := fakeVehicle{id: 1, name: "bicycle", gosNumber: "BIKE-1", ease: 7, broken: true}

	if err := g.Add(bike); err == nil {
		t.Fatal("гараж принял транспорт, который отклонил техосмотр")
	}
	if got := len(g.AllItemsReport()); got != 0 {
		t.Errorf("после отказа в гараже осталось %d объектов, ожидалось 0", got)
	}
}

func TestAddRejectsDuplicateIDAcrossItemTypes(t *testing.T) {
	g := newTestGarage()
	bike := fakeVehicle{id: 1, name: "bicycle", gosNumber: "BIKE-1", ease: 7}
	helmet, err := inventory.NewHelmet(1, 56)
	if err != nil {
		t.Fatal(err)
	}
	if err := g.Add(bike); err != nil {
		t.Fatal(err)
	}
	if err := g.Add(helmet); err == nil {
		t.Fatal("гараж принял вещь с ID, который уже занят транспортом")
	}
	if got := len(g.AllItemsReport()); got != 1 {
		t.Errorf("в гараже %d объектов, ожидался 1", got)
	}
}

func TestVehiclesForBeginnersUsesInclusiveThreshold(t *testing.T) {
	g := newTestGarage()
	arr := []fakeVehicle{
		{id: 1, name: "not beginner friendly", ease: 5},
		{id: 2, name: "beginner friendly", ease: 6},
	}
	for _, candidate := range arr {
		err := g.Add(candidate)
		if err != nil {
			t.Fatal(err)
		}
	}

	got := g.VehiclesForBeginners()
	if len(got) != 1 {
		t.Fatalf("получено %d транспортных средств, ожидалось 1", len(got))
	}
	if got[0].GetID() != 2 {
		t.Errorf("в список попал транспорт с ID=%d, ожидался ID 2", got[0].GetID())
	}
}

func TestDailyEnergyConsumptionReportSumsConsumersOnly(t *testing.T) {
	g := newTestGarage()
	for _, item := range []Item{
		fakeEnergyItem{id: 1, name: "charger A", consumption: 1.25},
		fakeEnergyItem{id: 2, name: "charger B", consumption: 2.75},
		plainItem{id: 3, name: "helmet"},
	} {
		if err := g.Add(item); err != nil {
			t.Fatal(err)
		}
	}

	if got, want := g.DailyEnergyConsumptionReport(), float32(4); got != want {
		t.Errorf("суточный расход = %v, ожидался %v", got, want)
	}
}

func TestInventoryAndAllItemsReports(t *testing.T) {
	g := newTestGarage()
	bike := fakeVehicle{id: 1, name: "bicycle", gosNumber: "BIKE-1", ease: 7}
	helmet, err := inventory.NewHelmet(2, 56)
	if err != nil {
		t.Fatal(err)
	}
	station, err := inventory.NewChargeStation(3, 2)
	if err != nil {
		t.Fatal(err)
	}
	sparePart := plainItem{id: 4, name: "spare wheel"}
	for _, item := range []Item{bike, helmet, station, sparePart} {
		if err := g.Add(item); err != nil {
			t.Fatal(err)
		}
	}

	items := g.InventoryReport()
	if len(items) != 3 {
		t.Fatalf("инвентарных вещей %d, ожидалось 3", len(items))
	}
	want := map[int64]string{2: "helmet", 3: "Charge Station", 4: "spare wheel"}
	for _, item := range items {
		name, ok := want[item.GetID()]
		if !ok || item.GetName() != name {
			t.Errorf("неожиданная вещь в отчёте: ID=%d, name=%q", item.GetID(), item.GetName())
		}
		delete(want, item.GetID())
	}
	if len(want) != 0 {
		t.Errorf("в отчёте отсутствуют вещи: %v", want)
	}
	if got := len(g.AllItemsReport()); got != 4 {
		t.Errorf("всех объектов %d, ожидалось 4", got)
	}

	var _ vehicle.IVehicle = bike
}
