package vehicle_sensor

import "algorithm2.0/util"

func SensorLayerSingleton(proxy *util.AllVehicleProxy) *SensorLayer {
	if instance == nil {
		instance = &SensorLayer{proxy: proxy}
	}
	return instance
}









var instance *SensorLayer = nil
type SensorLayer struct {
	proxy *util.AllVehicleProxy
}
