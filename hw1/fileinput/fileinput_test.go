package fileinput

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"main/app"
	"main/vehicle"
)

func TestReadFileExample(t *testing.T) {
	items, err := ReadFile("../data/example.json")
	if err != nil {
		t.Fatalf("прочитать пример: %v", err)
	}
	wantNames := []string{"bicycle", "scooter", "electric bicycle", "helmet", "Charge Station"}
	if len(items) != len(wantNames) {
		t.Fatalf("прочитано %d объектов, ожидалось %d", len(items), len(wantNames))
	}
	fleet := app.New().Garage
	for i, item := range items {
		if item.GetID() != int64(i+1) || item.GetName() != wantNames[i] {
			t.Errorf("неожиданный объект: ID=%d, name=%q", item.GetID(), item.GetName())
		}
		if transport, ok := item.(vehicle.IVehicle); ok {
			if transport.GetGosNumber() == "" || transport.CheckIsBroken() {
				t.Errorf("неожиданные данные транспорта: %+v", transport)
			}
		}
		if err := fleet.Add(item); err != nil {
			t.Fatalf("добавить объект в парк: %v", err)
		}
	}
	if got := fleet.DailyEnergyConsumptionReport(); got != 7.5 {
		t.Errorf("расход = %v, ожидалось 7.5", got)
	}
}

func TestReadFileInvalidData(t *testing.T) {
	cases := []struct {
		name string
		data string
		want string
	}{
		{"empty file", "", "прочитать JSON"},
		{"malformed JSON", "[", "прочитать JSON"},
		{"object instead of array", `{}`, "прочитать JSON"},
		{"null instead of array", `null`, "JSON-массив"},
		{"unknown field", `[{"type":"helmet","id":1,"szie":56}]`, "unknown field"},
		{"wrong field type", `[{"type":"helmet","id":"1"}]`, "прочитать JSON"},
		{"unknown item", `[{"type":"car","id":1}]`, "объект 1"},
		{"missing type", `[{"id":1}]`, "объект 1"},
		{"missing ID", `[{"type":"helmet"}]`, "объект 1"},
		{"invalid battery", `[{"type":"scooter","id":1,"battery":101}]`, "объект 1"},
		{"negative energy", `[{"type":"charge_station","id":1,"daily_energy_consumption":-1}]`, "объект 1"},
		{"invalid second item", `[{"type":"helmet","id":1},{"type":"bicycle","id":0}]`, "объект 2"},
		{"trailing garbage", `[{"type":"helmet","id":1}] garbage`, "лишние данные"},
		{"second JSON array", `[{"type":"helmet","id":1}] [{"type":"helmet","id":2}]`, "только один JSON-массив"},
		{"trailing null", `[{"type":"helmet","id":1}] null`, "только один JSON-массив"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "items.json")
			if err := os.WriteFile(path, []byte(tc.data), 0600); err != nil {
				t.Fatal(err)
			}
			items, err := ReadFile(path)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("ошибка = %v, ожидалась ошибка с %q", err, tc.want)
			}
			if items != nil {
				t.Fatal("при ошибке не должен возвращаться частичный результат")
			}
		})
	}
}

func TestReadFileEmptyArrayAndMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "items.json")
	if _, err := ReadFile(path); err == nil {
		t.Fatal("ожидалась ошибка открытия отсутствующего файла")
	}
	if err := os.WriteFile(path, []byte("[] \t\r\n"), 0600); err != nil {
		t.Fatal(err)
	}
	items, err := ReadFile(path)
	if err != nil || len(items) != 0 {
		t.Fatalf("пустой массив: items=%v, err=%v", items, err)
	}
}
