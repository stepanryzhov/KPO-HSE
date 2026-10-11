package main

import (
	"fmt"
	"io"

	"main/garage"
	"main/inventory"
	"main/vehicle"
)

func runDemo(w io.Writer, fleet *garage.Garage) error {
	fmt.Fprintln(w, "СЦЕНАРИЙ 1: ПУСТОЙ ПАРК")
	printConsoleReports(w, fleet)

	fmt.Fprintln(w, "\nСЦЕНАРИЙ 2: ПРИЁМ ИСПРАВНОЙ ТЕХНИКИ И ВЕЩЕЙ")
	bike, err := vehicle.NewBicycle(1, "BIKE-1", false)
	if err != nil {
		return fmt.Errorf("создать велосипед: %w", err)
	}
	scooter, err := vehicle.NewScooter(2, "SCOOTER-2", false, 80)
	if err != nil {
		return fmt.Errorf("создать самокат: %w", err)
	}
	electricBike, err := vehicle.NewElectricBicycle(3, "EBIKE-3", false, 60)
	if err != nil {
		return fmt.Errorf("создать электровелосипед: %w", err)
	}
	helmet, err := inventory.NewHelmet(4, 56)
	if err != nil {
		return fmt.Errorf("создать шлем: %w", err)
	}
	station, err := inventory.NewChargeStation(5, 5)
	if err != nil {
		return fmt.Errorf("создать зарядную станцию: %w", err)
	}

	for _, item := range []garage.Item{bike, scooter, electricBike, helmet, station} {
		if err := fleet.Add(item); err != nil {
			return fmt.Errorf("добавить %s в парк: %w", item.GetName(), err)
		}
		fmt.Fprintf(w, "Принят в парк: %s (инвентарный номер %d)\n", item.GetName(), item.GetID())
	}

	fmt.Fprintln(w, "\nСЦЕНАРИЙ 3: ОТКАЗ ТЕХОСМОТРА")
	brokenScooter, err := vehicle.NewScooter(6, "BROKEN-6", true, 30)
	if err != nil {
		return fmt.Errorf("создать сломанный самокат для демонстрации: %w", err)
	}
	if err := fleet.Add(brokenScooter); err != nil {
		fmt.Fprintf(w, "Ожидаемый отказ техосмотра: %v\n", err)
	} else {
		return fmt.Errorf("техосмотр принял сломанный транспорт")
	}

	fmt.Fprintln(w, "\nСЦЕНАРИЙ 4: ПОВТОРНЫЙ ИНВЕНТАРНЫЙ НОМЕР")
	duplicateHelmet, err := inventory.NewHelmet(1, 58)
	if err != nil {
		return fmt.Errorf("создать вещь с повторным номером для демонстрации: %w", err)
	}
	if err := fleet.Add(duplicateHelmet); err != nil {
		fmt.Fprintf(w, "Ожидаемый отказ при повторном инвентарном номере: %v\n", err)
	} else {
		return fmt.Errorf("гараж принял повторный инвентарный номер")
	}

	fmt.Fprintln(w, "\nСЦЕНАРИЙ 5: НЕКОРРЕКТНЫЕ ДАННЫЕ")
	if _, err := vehicle.NewBicycle(0, "INVALID-ID", false); err != nil {
		fmt.Fprintf(w, "Ожидаемый отказ для велосипеда с номером 0: %v\n", err)
	} else {
		return fmt.Errorf("создан велосипед с неположительным инвентарным номером")
	}
	if _, err := inventory.NewHelmet(-1, 56); err != nil {
		fmt.Fprintf(w, "Ожидаемый отказ для шлема с номером -1: %v\n", err)
	} else {
		return fmt.Errorf("создан шлем с отрицательным инвентарным номером")
	}
	for _, battery := range []int{-1, 101} {
		if _, err := vehicle.NewScooter(7, "INVALID-BATTERY", false, battery); err != nil {
			fmt.Fprintf(w, "Ожидаемый отказ при заряде %d%%: %v\n", battery, err)
		} else {
			return fmt.Errorf("создан самокат с недопустимым зарядом %d%%", battery)
		}
	}
	if _, err := inventory.NewChargeStation(8, -0.5); err != nil {
		fmt.Fprintf(w, "Ожидаемый отказ при расходе -0.5 кВт·ч/сутки: %v\n", err)
	} else {
		return fmt.Errorf("создана зарядная станция с отрицательным расходом энергии")
	}

	fmt.Fprintln(w, "\nСЦЕНАРИЙ 6: ИТОГОВЫЕ ОТЧЁТЫ")
	fmt.Fprintln(w, "После всех отказов на балансе остаются только успешно принятые объекты.")
	printConsoleReports(w, fleet)
	return nil
}
