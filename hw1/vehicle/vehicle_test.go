package vehicle

import "testing"

func TestNewVehicleValidation(t *testing.T) {
	cases := []struct {
		name      string
		vehicle   string
		id        int64
		ease      int
		wantError bool
	}{
		{name: "valid fields", vehicle: "bicycle", id: 1, ease: 5},
		{name: "blank name", vehicle: " ", id: 1, ease: 5, wantError: true},
		{name: "non-positive ID", vehicle: "bicycle", id: 0, ease: 5, wantError: true},
		{name: "negative ID", vehicle: "bicycle", id: -1, ease: 5, wantError: true},
		{name: "ease below range", vehicle: "bicycle", id: 1, ease: 0, wantError: true},
		{name: "ease above range", vehicle: "bicycle", id: 1, ease: 11, wantError: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := newVehicle(tc.vehicle, tc.id, tc.ease, "ABC-123", false)
			if tc.wantError {
				if err == nil {
					t.Fatal("ожидалась ошибка для некорректных данных")
				}
				return
			}
			if err != nil {
				t.Fatalf("не ожидалась ошибка: %v", err)
			}
			if got.GetName() != tc.vehicle || got.GetID() != tc.id || got.GetEaseOfUse() != tc.ease {
				t.Fatalf("создан транспорт с неожиданными данными: %+v", got)
			}
		})
	}
}

func TestNewElectricTransportValidation(t *testing.T) {
	cases := []struct {
		name        string
		id          int64
		ease        int
		battery     int
		consumption float32
		wantError   bool
	}{
		{name: "valid fields", id: 1, ease: 5, battery: 80, consumption: 1.5},
		{name: "negative energy consumption", id: 1, ease: 5, battery: 80, consumption: -1, wantError: true},
		{name: "battery below range", id: 1, ease: 5, battery: -1, consumption: 1, wantError: true},
		{name: "battery above range", id: 1, ease: 5, battery: 101, consumption: 1, wantError: true},
		{name: "invalid ease", id: 1, ease: 11, battery: 80, consumption: 1, wantError: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := newElectricTransport("electric vehicle", tc.id, tc.ease, "EV-123", false, tc.battery, tc.consumption)
			if tc.wantError {
				if err == nil {
					t.Fatal("ожидалась ошибка для некорректных данных")
				}
				return
			}
			if err != nil {
				t.Fatalf("не ожидалась ошибка: %v", err)
			}
			if got.GetEnergyConsumption() != tc.consumption {
				t.Errorf("расход энергии = %v, ожидался %v", got.GetEnergyConsumption(), tc.consumption)
			}
		})
	}
}

func TestNewBicycle(t *testing.T) {
	bike, err := NewBicycle(1, "BIKE-1", false)
	if err != nil {
		t.Fatalf("не удалось создать велосипед: %v", err)
	}

	if bike.GetName() != "bicycle" {
		t.Errorf("название = %q, ожидалось %q", bike.GetName(), "bicycle")
	}
	if bike.GetEaseOfUse() != 7 {
		t.Errorf("простота = %d, ожидалось 7", bike.GetEaseOfUse())
	}
	if bike.GetGosNumber() != "BIKE-1" {
		t.Errorf("госномер = %q, ожидался %q", bike.GetGosNumber(), "BIKE-1")
	}
	if bike.CheckIsBroken() {
		t.Error("новый исправный велосипед помечен как сломанный")
	}

	var _ IVehicle = bike
}

func TestNewScooter(t *testing.T) {
	scooter, err := NewScooter(2, "SCOOTER-2", false, 75)
	if err != nil {
		t.Fatalf("не удалось создать самокат: %v", err)
	}

	if scooter.GetName() != "scooter" || scooter.GetEnergyConsumption() != 1 {
		t.Errorf("неожиданные данные самоката: name=%q consumption=%v", scooter.GetName(), scooter.GetEnergyConsumption())
	}
	var _ IVehicle = scooter
}

func TestNewElectricBicycle(t *testing.T) {
	bike, err := NewElectricBicycle(3, "EBIKE-3", false, 60)
	if err != nil {
		t.Fatalf("не удалось создать электровелосипед: %v", err)
	}

	if bike.GetName() != "electric bicycle" {
		t.Errorf("название = %q, ожидалось %q", bike.GetName(), "electric bicycle")
	}
	if bike.GetEaseOfUse() != 6 {
		t.Errorf("простота = %d, ожидалось 6", bike.GetEaseOfUse())
	}
	if bike.GetEnergyConsumption() != 1.5 {
		t.Errorf("суточный расход = %v, ожидался 1.5", bike.GetEnergyConsumption())
	}
	var _ IVehicle = bike
}
