package vehicle

import "fmt"

type electricBicycle struct {
	electricTransport
}

func NewElectricBicycle(id int64, gosNumber string, isBroken bool, battery int) (*electricBicycle, error) {
	base, err := newElectricTransport("electric bicycle", id, 6, gosNumber, isBroken, battery, 1.5)
	if err != nil {
		return nil, err
	}

	return &electricBicycle{electricTransport: *base}, nil
}

func (b electricBicycle) Ride() {
	fmt.Println("electric bicycle", b.id, "started ride")
}
