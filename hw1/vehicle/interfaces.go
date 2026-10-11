package vehicle

type IVehicle interface {
	Ride()
	CheckIsBroken() bool
	GetID() int64
	GetEaseOfUse() int
	GetGosNumber() string
	GetName() string
}
