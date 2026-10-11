package garage

type Item interface {
	GetID() int64
	GetName() string
}

type IElectroConsumer interface {
	GetEnergyConsumption() float32
}
