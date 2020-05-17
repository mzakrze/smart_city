package vehicle

func AllVehiclesProxySingleton() *AllVehicleProxy {
	if proxyInstance == nil {
		proxyInstance = &AllVehicleProxy{}
	}
	return proxyInstance
}

func (p *AllVehicleProxy) RegisterVehicle(actor *VehicleActor) {
	p.registeredVehicles = append(p.registeredVehicles, actor)
}

func (p*AllVehicleProxy) UnregisterVehicle(actor *VehicleActor) {
	var index = -1
	for i := range p.registeredVehicles {
		if p.registeredVehicles[i].Id == actor.Id {
			index = i
			break
		}
	}
	if index < 0 {
		panic("Illegal state")
	}

	p.registeredVehicles = append(p.registeredVehicles[:index], p.registeredVehicles[index + 1:]...)
}

func (p*AllVehicleProxy) GetAllVehicles() []*VehicleActor {
	return p.registeredVehicles
}

func (p *AllVehicleProxy) AllVehiclesDone() bool {
	return true // TODO
}



var proxyInstance *AllVehicleProxy = nil
type AllVehicleProxy struct {
	registeredVehicles []*VehicleActor
}

