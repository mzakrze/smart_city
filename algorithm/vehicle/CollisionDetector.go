package vehicle

type CollisionDetector struct {

}

func NewCollisionDetector(proxy *AllVehicleProxy) *CollisionDetector {
	return &CollisionDetector{}
}
