package vehicle

import (
	"algorithm/constants"
	"algorithm/types"
	"algorithm/util"
	"testing"
)

func TestIMPCreatesPolicyFcfs(t *testing.T) {
	graph := generateSimpleRoad()
	c := util.Configuration{IntersectionPolicy: "fcfs"}
	p := AllVehiclesProxySingleton()
	instance = nil
	nc := CommunicationLayerSingleton(p, c)

	im, err := IntersectionManagerSingleton(graph, nc, c)

	if err != nil {
		t.Fatal("Error while creating IM")
	}

	_ = im.policy.(*IntersectionPolicyFcfs) // type assertion
}

func TestIMPCreatesPolicyTrafficLights(t *testing.T) {
	graph := generateSimpleRoad()
	c := util.Configuration{IntersectionPolicy: "trafficLights"}
	p := AllVehiclesProxySingleton()
	nc := CommunicationLayerSingleton(p, c)

	instance = nil
	im, err := IntersectionManagerSingleton(graph, nc, c)

	if err != nil {
		t.Fatal("Error while creating IM")
	}

	_ = im.policy.(*IntersectionPolicyTrafficLights) // type assertion
}

func TestIMPings(t *testing.T) {

	graph := generateSimpleRoad()
	v1 := &VehicleActor{Id: 1, EntryPoint: graph.Entrypoints[0]}

	c := util.Configuration{IntersectionPolicy: "fcfs"}
	p := AllVehiclesProxySingleton()
	nc := CommunicationLayerSingleton(p, c)

	instance = nil
	im, err := IntersectionManagerSingleton(graph, nc, c)

	if err != nil {
		panic(err)
	}

	p.Enqueue(v1)
	p.RegisterVehicle(p.NextQueued(1))

	nc.SendDsrcV2R(DsrcV2RMessage{
		MsgType: AimProtocolMsgRequest,
		TsSent: 10,
		Sender: 10,
		VehicleX: 50,
		VehicleY: 50,
		VehicleSpeed: 10,
		VehicleMass: 1300,
		VehiclePower: 735 * 70,
		ApproachConflictZoneMinTs: 10,
		ApproachConflictZoneSpeed: 10,
		ConflictZoneNodeEnter: nil,
		ConflictZoneNodeExit: nil,
		MaxSpeedOnCurve: 10,
		IsTurning: false,
		EntryPointId: 1,
		ExitPointId: 1,
		Route: []types.Location{{X: 50, Y: 50}, {X: 60, Y: 60}},
	})

	for ts := types.Millisecond(10); ts < 10e2; ts += constants.SimulationStepInterval {
		im.Ping(ts)
	}

	if len(nc.VehicleReceive(10e2, 10)) == 0 {
		t.Error("No replies from Intersection manager")
	}
}