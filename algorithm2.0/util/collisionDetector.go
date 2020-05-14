package util

type CollisionDetector struct {
	proxy *AllVehicleProxy
}

func NewCollisionDetector(proxy *AllVehicleProxy) *CollisionDetector {
	return &CollisionDetector{proxy: proxy}
}
