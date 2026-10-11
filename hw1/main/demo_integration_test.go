package main

import (
	"io"
	"testing"

	"main/app"
	"main/service"
)

func TestRunDemoWithDefaultContainer(t *testing.T) {
	container := app.New()
	if err := runDemo(io.Discard, container.Garage); err != nil {
		t.Fatalf("демонстрация с настоящими зависимостями завершилась ошибкой: %v", err)
	}

	// Отказы не должны добавлять объекты или заменять велосипед шлемом
	// с повторным номером. Порядок обхода карты гаража не важен.
	want := map[int64]string{
		1: "bicycle",
		2: "scooter",
		3: "electric bicycle",
		4: "helmet",
		5: "Charge Station",
	}
	items := container.Garage.AllItemsReport()
	if len(items) != len(want) {
		t.Fatalf("на балансе %d объектов, ожидалось %d", len(items), len(want))
	}
	for _, item := range items {
		name, ok := want[item.GetID()]
		if !ok || item.GetName() != name {
			t.Errorf("неожиданный объект после демонстрации: ID=%d, name=%q", item.GetID(), item.GetName())
		}
		delete(want, item.GetID())
	}
	if len(want) != 0 {
		t.Errorf("после демонстрации отсутствуют объекты: %v", want)
	}

	// Проверяем совместную работу реальных моделей и расчёта гаража:
	// самокат, электровелосипед и станция учитываются, отклонённый самокат — нет.
	if got := container.Garage.DailyEnergyConsumptionReport(); got != 7.5 {
		t.Errorf("расход парка = %v кВт·ч/сутки, ожидался 7.5", got)
	}
}

type inspectionFunc func(service.ICheckable) bool

func (f inspectionFunc) CheckIsBroken(item service.ICheckable) bool {
	return f(item)
}

func TestRunDemoStopsAfterInspectionFailure(t *testing.T) {
	inspections := 0
	container := app.NewWithService(inspectionFunc(func(_ service.ICheckable) bool {
		inspections++
		return inspections == 2
	}))

	if err := runDemo(io.Discard, container.Garage); err == nil {
		t.Fatal("демонстрация должна вернуть ошибку, если сервис отклонил технику при первоначальном наполнении")
	}
	if inspections != 2 {
		t.Errorf("выполнено %d техосмотров, ожидалось 2: сценарий должен остановиться после отказа", inspections)
	}
	items := container.Garage.AllItemsReport()
	if len(items) != 1 {
		t.Fatalf("после отказа на балансе %d объектов, должен остаться только принятый велосипед", len(items))
	}
	if items[0].GetID() != 1 || items[0].GetName() != "bicycle" {
		t.Errorf("потерян ранее принятый велосипед: ID=%d, name=%q", items[0].GetID(), items[0].GetName())
	}
}
