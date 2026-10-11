package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"main/app"
)

func TestRunSelectsMode(t *testing.T) {
	data, err := os.ReadFile("../data/example.json")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "данные парка.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name  string
		input string
		demo  bool
	}{
		{"demo", "1\n", true},
		{"file path with spaces", "2\n" + path + "\n", false},
		{"retry invalid mode", "wrong\n 2 \r\n" + path, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fleet := app.New().Garage
			var output bytes.Buffer
			if err := run(strings.NewReader(tc.input), &output, fleet); err != nil {
				t.Fatalf("запуск: %v", err)
			}
			if len(fleet.AllItemsReport()) != 5 || fleet.DailyEnergyConsumptionReport() != 7.5 {
				t.Fatal("режим не загрузил ожидаемые объекты")
			}
			if strings.Contains(output.String(), "СЦЕНАРИЙ 1") != tc.demo {
				t.Fatal("запущен неверный режим")
			}
			if !strings.Contains(output.String(), "Всего объектов: 5 (транспорт: 3, вещи: 2)") {
				t.Fatal("не выведен итоговый отчёт")
			}
		})
	}
}

func TestRunFileContinuesAfterRejectedItems(t *testing.T) {
	path := filepath.Join(t.TempDir(), "items.json")
	data := `[
		{"type":"bicycle","id":1,"gos_number":"BIKE-1"},
		{"type":"helmet","id":1,"size":56},
		{"type":"scooter","id":2,"is_broken":true,"battery":80},
		{"type":"scooter","id":3,"is_broken":false,"battery":80},
		{"type":"charge_station","id":4,"daily_energy_consumption":2.5}
	]`
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	fleet := app.New().Garage
	var output bytes.Buffer
	if err := run(strings.NewReader("2\n"+path+"\n"), &output, fleet); err != nil {
		t.Fatal(err)
	}
	if len(fleet.AllItemsReport()) != 3 || fleet.DailyEnergyConsumptionReport() != 3.5 {
		t.Fatal("неверное состояние парка после отказов")
	}
	if strings.Count(output.String(), "Не принят в парк:") != 2 {
		t.Fatal("для каждого отказа должно выводиться сообщение")
	}
	if !strings.Contains(output.String(), "Всего объектов: 3 (транспорт: 2, вещи: 1)") {
		t.Fatal("не выведен отчёт с успешно принятыми объектами")
	}
}

func TestRunInputErrorsLeaveFleetEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "invalid.json")
	if err := os.WriteFile(path, []byte(`[{"type":"helmet","id":1},{"type":"bicycle","id":0}]`), 0600); err != nil {
		t.Fatal(err)
	}
	for _, input := range []string{"", "2\n", "2\n\n", "2\n" + path + "\n"} {
		t.Run(input, func(t *testing.T) {
			fleet := app.New().Garage
			var output bytes.Buffer
			if err := run(strings.NewReader(input), &output, fleet); err == nil {
				t.Fatal("ожидалась ошибка ввода")
			}
			if len(fleet.AllItemsReport()) != 0 {
				t.Fatal("при ошибке ввода имущество не должно добавляться в парк")
			}
		})
	}
}
