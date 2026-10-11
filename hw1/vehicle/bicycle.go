package vehicle

import (
	"fmt"
)

type bicycle struct {
	vehicle
}

func NewBicycle(id int64, gosnum string, isBroken bool) (*bicycle, error) {

	base, err := newVehicle("bicycle", id, 7, gosnum, isBroken)

	if err != nil {
		return nil, err
	}

	return &bicycle{*base}, nil

}

func (b bicycle) Ride() {
	fmt.Println("bicycle", b.id, "started ride")
}
