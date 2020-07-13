package vehicle

import (
	"algorithm/constants"
	"algorithm/types"
	"algorithm/util"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"
)

func TestCleanUpOldReservation(t *testing.T) {
	graph := generateSimpleRoad()
	c := util.Configuration{IntersectionPolicy: "fcfs"}
	p := AllVehiclesProxySingleton()
	nc := CommunicationLayerSingleton(p, c)

	instance = nil
	im, err := IntersectionManagerSingleton(graph, nc, c)

	if err != nil {
		t.Fatal("Error while creating IM")
	}

	ipFcfs := im.policy.(*IntersectionPolicyFcfs) // type assertion

	nc.SendDsrcV2R(DsrcV2RMessage{
		MsgType: AimProtocolMsgRequest,
		TsSent: 10,
		Sender: 1,
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
	const expectedReservationTo = 1090

	for ts := types.Millisecond(10); ts < 3000; ts += constants.SimulationStepInterval {
		im.Ping(ts)

		if ts % 990 == 0 {
			before := len(ipFcfs.reservations)
			ipFcfs.cleanupOldReservations(ts + 10)
			after := len(ipFcfs.reservations)

			if before != 0 && after >= before && ts > expectedReservationTo {
				t.Error(fmt.Sprintf("Doesnt work, before: %d, after: %d", before, after))
			}
		}
	}

	msgs := nc.VehicleReceive(3000, 1)
	if len(msgs) == 0 {
		fmt.Println("No reservations created")
	}

	//if msgs[0].reservationToTs != expectedReservationTo {
	//	t.Error("Incorrect value")
	//}
}

func excluded_TestReservationTableVisualy(t *testing.T) {

	graph := &util.Graph{
		ConflictZone: util.ConflictZone{
			MinX: 0,
			MaxX: 40,
			MinY: 0,
			MaxY: 40,
		},
	}

	fcfs := CreateIntersectionPolicyFcfs(graph, util.Configuration{})

	for counter := 0; counter < 100; counter++ {

		colors := []color.Color{color.Black, color.White}
		rect := image.Rect(0, 0, gridNoX, gridNoY)
		newFrame := image.NewPaletted(rect, colors)
		for x := range fcfs.reservationTable[counter] {
			for y := range fcfs.reservationTable[counter][x] {
				if fcfs.reservationTable[counter][x][y] == taken {
					if x  < 0 || y < 0 {
						continue
					}
					if x >= gridNoX || y >= gridNoY {
						continue
					}
					newFrame.Set(x, gridNoY - y , color.Black)
				} else {
					newFrame.Set(x, gridNoY - y, color.White)
				}
			}
		}

		f, _ := os.Create(fmt.Sprintf("img/image%05d.png", counter))
		png.Encode(f, newFrame)
	}
}

func TestCancelingReservation(t *testing.T) {
	instance = nil
	proxyInstance = nil
	instanceCommunication = nil

	graph := generateSimpleRoad()
	c := util.Configuration{IntersectionPolicy: "fcfs"}
	p := AllVehiclesProxySingleton()
	nc := CommunicationLayerSingleton(p, c)

	im, _ := IntersectionManagerSingleton(graph, nc, c)
	//fcfs := im.policy.(*IntersectionPolicyFcfs)

	postRequest := func(sender types.VehicleId, ts types.Millisecond, wId types.WayId) {
		nc.SendDsrcV2R(DsrcV2RMessage{
			MsgType: AimProtocolMsgRequest,
			TsSent: 10,
			Sender: sender,
			VehicleX: 50,
			VehicleY: 50,
			VehicleSpeed: 10,
			VehicleMass: 1300,
			VehiclePower: 735 * 70,
			ApproachConflictZoneMinTs: ts,
			ApproachConflictZoneSpeed: 10,
			ConflictZoneNodeEnter: nil,
			ConflictZoneNodeExit: nil,
			MaxSpeedOnCurve: 10,
			IsTurning: false,
			EntryPointId: wId,
			ExitPointId: 1,
			Route: []types.Location{{X: 50, Y: 50}, {X: 60, Y: 60}},
		})
	}

	postRequest(1, 10, 1)
	im.Ping(10)
	msgs := nc.VehicleReceive(10, 1)
	if len(msgs) == 0 {
		t.Fatal("No reservation, but should be")
	}

	postRequest(1, 20, 1)
	im.Ping(20)
	msgs2 := nc.VehicleReceive(20, 1)
	if len(msgs2) > 0 {
		t.Fatal("This should be reserved, but got reservation")
	}

	nc.SendDsrcV2R(DsrcV2RMessage{Sender: 1, ReservationToCancelId: msgs[0].reservationId, MsgType: AimProtocolMsgReservationCancelation})
	postRequest(1, 30, 1)
	im.Ping(30)
	msgs3 := nc.VehicleReceive(30, 1)
	if len(msgs3) == 0 {
		t.Fatal("No reservation, but should be (because previous is cancelled)")
	}

}

func TestTrafficLightPolicy(t *testing.T) {
	graph := generateSimpleRoad()
	c := util.Configuration{IntersectionPolicy: "trafficLights"}
	p := AllVehiclesProxySingleton()
	nc := CommunicationLayerSingleton(p, c)

	instance = nil
	im, _ := IntersectionManagerSingleton(graph, nc, c)

	trafficLights := im.policy.(*IntersectionPolicyTrafficLights)

	postRequest := func(sender types.VehicleId, ts types.Millisecond, wId types.WayId) {
		nc.SendDsrcV2R(DsrcV2RMessage{
			MsgType: AimProtocolMsgRequest,
			TsSent: 10,
			Sender: sender,
			VehicleX: 50,
			VehicleY: 50,
			VehicleSpeed: 10,
			VehicleMass: 1300,
			VehiclePower: 735 * 70,
			ApproachConflictZoneMinTs: ts,
			ApproachConflictZoneSpeed: 10,
			ConflictZoneNodeEnter: nil,
			ConflictZoneNodeExit: nil,
			MaxSpeedOnCurve: 10,
			IsTurning: false,
			EntryPointId: wId,
			ExitPointId: 1,
			Route: []types.Location{{X: 50, Y: 50}, {X: 60, Y: 60}},
		})
	}

	for ts := types.Millisecond(10); ts < 60000; ts += constants.SimulationStepInterval {
		postRequest(1, ts, 1)
		postRequest(2, ts, 2)

		im.Ping(ts)
		trafficLights.GetReplies(ts)

		if len(nc.VehicleReceive(ts, 1)) > 0 {
			wayIdExpected := ts / (trafficLights.millisecondsFree + trafficLights.millisecondWait)
			wayIdExpected = wayIdExpected % 4

			if wayIdExpected != 0 {
				t.Fatal("Expected 0, got: " , wayIdExpected)
			}
		}
		if len(nc.VehicleReceive(ts, 2)) > 0 {
			wayIdExpected := ts / (trafficLights.millisecondsFree + trafficLights.millisecondWait)
			wayIdExpected = wayIdExpected % 4

			if wayIdExpected != 1 {
				t.Fatal("Expected 1, got: " , wayIdExpected)
			}
		}
	}
}
