package service

type ServiceCenter struct {
}

func (s ServiceCenter) CheckIsBroken(sr ICheckable) bool {
	return sr.CheckIsBroken()
}

func NewService() *ServiceCenter {
	return &ServiceCenter{}
}
