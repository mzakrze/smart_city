package util

func AllVehiclesProxySingleton() *AllVehicleProxy {
	if instance == nil {
		instance = &AllVehicleProxy{}
	}
	return instance
}

func (p *AllVehicleProxy) RegisterVehicle() {

}

func (p*AllVehicleProxy) UnregisterVehicle() {

}

func (p*AllVehicleProxy) GetAllVehicles() {

}





var instance *AllVehicleProxy = nil
type AllVehicleProxy struct {

}

