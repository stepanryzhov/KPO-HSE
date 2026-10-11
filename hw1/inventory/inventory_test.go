package inventory

import "testing"

func TestNewInventoryItem(t *testing.T) {
	if _, err := NewInventoryItem(0, "helmet"); err == nil {
		t.Fatal("ожидалась ошибка для неположительного ID")
	}

	item, err := NewInventoryItem(1, "helmet")
	if err != nil {
		t.Fatalf("не удалось создать инвентарный объект: %v", err)
	}
	if item.GetID() != 1 || item.GetName() != "helmet" {
		t.Errorf("получен объект с ID=%d и названием %q", item.GetID(), item.GetName())
	}
}

func TestNewHelmet(t *testing.T) {
	helmet, err := NewHelmet(2, 56)
	if err != nil {
		t.Fatalf("не удалось создать шлем: %v", err)
	}
	if helmet.GetID() != 2 || helmet.GetName() != "helmet" || helmet.size != 56 {
		t.Errorf("неожиданные данные шлема: ID=%d, name=%q, size=%d", helmet.GetID(), helmet.GetName(), helmet.size)
	}
}

func TestNewChargeStation(t *testing.T) {
	if _, err := NewChargeStation(1, -0.5); err == nil {
		t.Fatal("ожидалась ошибка для отрицательного энергопотребления")
	}
	if _, err := NewChargeStation(0, 1); err == nil {
		t.Fatal("ожидалась ошибка для неположительного ID")
	}

	station, err := NewChargeStation(3, 2.5)
	if err != nil {
		t.Fatalf("не удалось создать зарядную станцию: %v", err)
	}
	if station.GetID() != 3 || station.GetName() != "Charge Station" || station.GetEnergyConsumption() != 2.5 {
		t.Errorf("неожиданные данные зарядной станции: ID=%d, name=%q, consumption=%v",
			station.GetID(), station.GetName(), station.GetEnergyConsumption())
	}
}
