package vehicle

import "algorithm/types"

func AllVehiclesProxySingleton() *AllVehicleProxy {
	enqueuedVehicles := map[types.WayId][]*VehicleActor {
		1: []*VehicleActor{},
		2: []*VehicleActor{},
		3: []*VehicleActor{},
		4: []*VehicleActor{},
	}
	if proxyInstance == nil {
		proxyInstance = &AllVehicleProxy{
			activeVehicles: []*VehicleActor{},
			enqueuedVehicles: enqueuedVehicles,
			finishedVehicles: []*VehicleActor{},
		}
	}
	return proxyInstance
}

func (p*AllVehicleProxy) VehicleFinished(actor *VehicleActor) {
	var index = -1
	for i := range p.activeVehicles {
		if p.activeVehicles[i].Id == actor.Id {
			index = i
			break
		}
	}
	if index < 0 {
		panic("Illegal State")
	}

	p.activeVehicles = append(p.activeVehicles[:index], p.activeVehicles[index + 1:]...)
	p.finishedVehicles = append(p.finishedVehicles, actor)
}

func (p*AllVehicleProxy) GetAllActiveVehicles() []*VehicleActor {
	return p.activeVehicles
}

func (p *AllVehicleProxy) Enqueue(v *VehicleActor) {
	p.enqueuedVehicles[v.EntryPoint.WayId] = append(p.enqueuedVehicles[v.EntryPoint.WayId], v)
}

func (p *AllVehicleProxy) NextQueued(wId types.WayId) *VehicleActor {
	if len(p.enqueuedVehicles[wId]) == 0 {
		return nil
	}
	return p.enqueuedVehicles[wId][0]
}

func (p *AllVehicleProxy) RegisterVehicle(v *VehicleActor) {
	for _, wId := range []types.WayId{1,2,3,4} {
		if len(p.enqueuedVehicles[wId]) > 0 && p.enqueuedVehicles[wId][0].Id == v.Id {
			p.enqueuedVehicles[wId] = p.enqueuedVehicles[wId][1:]
			p.activeVehicles = append(p.activeVehicles, v)
			return
		}
	}

	panic("Oops")
}

func (p *AllVehicleProxy) GetAllVehiclesIntroduced() []*VehicleActor {
	res := []*VehicleActor{}

	res = append(res, p.activeVehicles...)
	res = append(res, p.finishedVehicles...)

	return res
}

var proxyInstance *AllVehicleProxy = nil
type AllVehicleProxy struct {
	activeVehicles []*VehicleActor
	enqueuedVehicles map[types.WayId][]*VehicleActor
	finishedVehicles []*VehicleActor
}

