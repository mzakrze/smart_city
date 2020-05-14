package vehicle_communication

import (
	"algorithm2.0/types"
	"algorithm2.0/util"
)

func CommunicationLayerSingleton(proxy *util.AllVehicleProxy) *CommunicationLayer {
	if instance == nil {
		instance = &CommunicationLayer{proxy: proxy}
	}
	return instance
}

func (c *CommunicationLayer) SendDsrcV2R() {

}

func (c *CommunicationLayer) SendDsrcR2V() {

}

func (c *CommunicationLayer) VehicleReceive(id types.VehicleId) {

}

func (c *CommunicationLayer) IntersectionManagerReceive() {

}







var instance *CommunicationLayer = nil
type CommunicationLayer struct {
	proxy *util.AllVehicleProxy
}





