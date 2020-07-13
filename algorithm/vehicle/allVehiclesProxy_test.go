package vehicle

import (
	"algorithm/types"
	"algorithm/util"
	"testing"
)

func assertExpectedActual(t *testing.T, e, a types.VehicleId) {
	if e != a {
		t.Error("Should be", e, ", actual:", a)
	}
}

func assertExpectedActualInt(t *testing.T, e, a int) {
	if e != a {
		t.Error("Should be", e, ", actual:", a)
	}
}

func TestFifoQueuePushPop(t *testing.T) {

	p := AllVehiclesProxySingleton()

	if p.NextQueued(0) != nil {
		t.Fatal("Should be empty")
	}

	p.Enqueue(&VehicleActor{Id: 1, EntryPoint: &util.Node{Id: 1}})
	p.Enqueue(&VehicleActor{Id: 2, EntryPoint: &util.Node{Id: 1}})
	p.Enqueue(&VehicleActor{Id: 3, EntryPoint: &util.Node{Id: 1}})

	assertExpectedActual(t, p.NextQueued(1).Id, 1)
	assertExpectedActual(t, p.NextQueued(1).Id, 1)
	assertExpectedActual(t, p.NextQueued(1).Id, 1)

	assertExpectedActualInt(t, 0, len(p.GetAllActiveVehicles()))
	p.RegisterVehicle(p.NextQueued(1))
	assertExpectedActualInt(t, 1, len(p.GetAllActiveVehicles()))

	assertExpectedActual(t, p.NextQueued(1).Id, 2)
	assertExpectedActual(t, p.NextQueued(1).Id, 2)
	assertExpectedActual(t, p.NextQueued(1).Id, 2)

	p.Enqueue(&VehicleActor{Id: 4, EntryPoint: &util.Node{Id: 1}})
	p.Enqueue(&VehicleActor{Id: 5, EntryPoint: &util.Node{Id: 1}})

	assertExpectedActualInt(t, 1, len(p.GetAllActiveVehicles()))
	p.RegisterVehicle(p.NextQueued(1))
	assertExpectedActual(t, p.NextQueued(1).Id, 3)
	assertExpectedActualInt(t, 2, len(p.GetAllActiveVehicles()))

	p.RegisterVehicle(p.NextQueued(1))
	assertExpectedActual(t, p.NextQueued(1).Id, 4)

	p.Enqueue(&VehicleActor{Id: 6, EntryPoint: &util.Node{Id: 1}})

	p.RegisterVehicle(p.NextQueued(1))
	assertExpectedActual(t, p.NextQueued(1).Id, 5)

	p.RegisterVehicle(p.NextQueued(1))
	assertExpectedActual(t, p.NextQueued(1).Id, 6)

	p.RegisterVehicle(p.NextQueued(1))

	assertExpectedActualInt(t, 6, len(p.GetAllVehiclesIntroduced()))
	if p.NextQueued(1) != nil {
		t.Fatal("Should be empty")
	}
}

func TestFifoQueueFinish(t *testing.T) {
	proxyInstance = nil
	p := AllVehiclesProxySingleton()
	assertExpectedActualInt(t, 0, len(p.finishedVehicles))
	v := &VehicleActor{Id: 1, EntryPoint: &util.Node{Id: 1}}
	p.Enqueue(v)
	assertExpectedActualInt(t, 0, len(p.finishedVehicles))
	p.RegisterVehicle(p.NextQueued(1))
	p.VehicleFinished(v)
	assertExpectedActualInt(t, 1, len(p.finishedVehicles))
	assertExpectedActualInt(t, 1, len(p.GetAllVehiclesIntroduced()))
}
