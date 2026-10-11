package service

type ICheckable interface {
	CheckIsBroken() bool
}

type IService interface {
	CheckIsBroken(ICheckable) bool
}
