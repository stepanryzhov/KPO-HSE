package vehicle

import (
	"fmt"
)

type scooter struct {
	electricTransport
}

func (s scooter) Ride() {
	fmt.Println("scooter", s.id, "started ride")
}

func NewScooter(id int64, gosnum string, isBroken bool, battery int) (*scooter, error) {

	base, err := newElectricTransport("scooter", id, 4, gosnum, isBroken, battery, 1)

	if err != nil {
		return nil, err
	}

	return &scooter{*base}, nil

}
