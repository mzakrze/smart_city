package vehicle

import (
	"algorithm/types"
)

func AllVehiclesProxySingleton() *AllVehicleProxy {
	if proxyInstance == nil {
		proxyInstance = &AllVehicleProxy{
			activeVehicles: []*VehicleActor{},
			enqueuedVehicles: map[types.WayId][]*VehicleActor {},
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
	if _, e := p.enqueuedVehicles[v.EntryPoint.Id]; e == false {
		p.enqueuedVehicles[v.EntryPoint.Id] = []*VehicleActor{}
	}
	p.enqueuedVehicles[v.EntryPoint.Id] = append(p.enqueuedVehicles[v.EntryPoint.Id], v)
}

func (p *AllVehicleProxy) NextQueued(nId types.NodeId) *VehicleActor {
	if len(p.enqueuedVehicles[nId]) == 0 {
		return nil
	}
	return p.enqueuedVehicles[nId][0]
}

func (p *AllVehicleProxy) RegisterVehicle(v *VehicleActor) {

	if p.enqueuedVehicles[v.EntryPoint.Id][0].Id != v.Id {
		panic("Oops")
	}
	p.enqueuedVehicles[v.EntryPoint.Id] = p.enqueuedVehicles[v.EntryPoint.Id][1:]
	p.activeVehicles = append(p.activeVehicles, v)
}

func (p *AllVehicleProxy) GetAllVehiclesIntroduced() []*VehicleActor {
	res := []*VehicleActor{}

	for nId, _ := range p.enqueuedVehicles {
		res = append(res, p.enqueuedVehicles[nId]...)
	}
	res = append(res, p.activeVehicles...)
	res = append(res, p.finishedVehicles...)

	return res
}

var proxyInstance *AllVehicleProxy = nil
type AllVehicleProxy struct {
	activeVehicles []*VehicleActor
	enqueuedVehicles map[types.NodeId][]*VehicleActor
	finishedVehicles []*VehicleActor
}

