package app

import (
	"main/garage"
	"main/service"
)

type Container struct {
	Garage *garage.Garage
}

func New() *Container {
	return NewWithService(service.NewService())
}

func NewWithService(service service.IService) *Container {
	return &Container{
		Garage: garage.NewGarage(service),
	}
}
