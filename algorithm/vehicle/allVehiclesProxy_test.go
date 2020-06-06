package vehicle

import (
	"algorithm/types"
	"testing"
)

func assertExpectedActual(t *testing.T, e, a types.VehicleId) {
	if e != a {
		t.Error("Should be", e, ", actual:", a)
	}
}

func TestFifoQueue(t *testing.T) {

	// TODO - rewrite
	//p := AllVehiclesProxySingleton()
	//
	//if p.NextQueued() != nil {
	//	t.Fatal("Should be empty")
	//}
	//
	//p.Enqueue(&VehicleActor{Id: 1})
	//p.Enqueue(&VehicleActor{Id: 2})
	//p.Enqueue(&VehicleActor{Id: 3})
	//
	//assertExpectedActual(t, p.NextQueued().Id, 1)
	//assertExpectedActual(t, p.NextQueued().Id, 1)
	//assertExpectedActual(t, p.NextQueued().Id, 1)
	//
	//p.RegisterNextVehicleFromQueue()
	//
	//assertExpectedActual(t, p.NextQueued().Id, 2)
	//assertExpectedActual(t, p.NextQueued().Id, 2)
	//assertExpectedActual(t, p.NextQueued().Id, 2)
	//
	//p.Enqueue(&VehicleActor{Id: 4})
	//p.Enqueue(&VehicleActor{Id: 5})
	//
	//p.RegisterNextVehicleFromQueue()
	//assertExpectedActual(t, p.NextQueued().Id, 3)
	//
	//p.RegisterNextVehicleFromQueue()
	//assertExpectedActual(t, p.NextQueued().Id, 4)
	//
	//p.Enqueue(&VehicleActor{Id: 6})
	//
	//p.RegisterNextVehicleFromQueue()
	//assertExpectedActual(t, p.NextQueued().Id, 5)
	//
	//p.RegisterNextVehicleFromQueue()
	//assertExpectedActual(t, p.NextQueued().Id, 6)
	//
	//p.RegisterNextVehicleFromQueue()
	//
	//if p.NextQueued() != nil {
	//	t.Fatal("Should be empty")
	//}
}
