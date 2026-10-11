package main

import (
	"fmt"
	"io"
	"sort"

	"main/garage"
	"main/vehicle"
)

type reportSource interface {
	VehiclesReport() []vehicle.IVehicle
	VehiclesForBeginners() []vehicle.IVehicle
	DailyEnergyConsumptionReport() float32
	AllItemsReport() []garage.Item
}

func printConsoleReports(w io.Writer, source reportSource) {
	vehicles := source.VehiclesReport()
	sort.Slice(vehicles, func(i, j int) bool {
		return vehicles[i].GetID() < vehicles[j].GetID()
	})

	fmt.Fprintln(w, "=== Транспорт парка ===")
	fmt.Fprintf(w, "Количество транспортных средств: %d\n", len(vehicles))
	if len(vehicles) == 0 {
		fmt.Fprintln(w, "Транспорта пока нет.")
	} else {
		for _, item := range vehicles {
			fmt.Fprintf(w, "ID: %d | %s | Госномер: %s | Простота для новичка: %d/10\n",
				item.GetID(), item.GetName(), item.GetGosNumber(), item.GetEaseOfUse())
		}
	}

	beginnerVehicles := source.VehiclesForBeginners()
	sort.Slice(beginnerVehicles, func(i, j int) bool {
		return beginnerVehicles[i].GetID() < beginnerVehicles[j].GetID()
	})

	fmt.Fprintln(w, "\n=== Транспорт для новичков (простота от 6/10) ===")
	if len(beginnerVehicles) == 0 {
		fmt.Fprintln(w, "Подходящего транспорта пока нет.")
	} else {
		for _, item := range beginnerVehicles {
			fmt.Fprintf(w, "ID: %d | %s | Госномер: %s | Простота: %d/10\n",
				item.GetID(), item.GetName(), item.GetGosNumber(), item.GetEaseOfUse())
		}
	}

	fmt.Fprintln(w, "\n=== Суточное потребление энергии ===")
	fmt.Fprintf(w, "Всего: %.2f кВт·ч/сутки\n", source.DailyEnergyConsumptionReport())

	items := source.AllItemsReport()
	sort.Slice(items, func(i, j int) bool {
		return items[i].GetID() < items[j].GetID()
	})

	fmt.Fprintln(w, "\n=== Имущество на балансе ===")
	fmt.Fprintf(w, "Всего объектов: %d (транспорт: %d, вещи: %d)\n",
		len(items), len(vehicles), len(items)-len(vehicles))
	if len(items) == 0 {
		fmt.Fprintln(w, "Имущества пока нет.")
	} else {
		for _, item := range items {
			fmt.Fprintf(w, "Инвентарный номер: %d | %s\n", item.GetID(), item.GetName())
		}
	}
}
