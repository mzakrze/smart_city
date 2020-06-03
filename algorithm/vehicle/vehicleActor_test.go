package vehicle

import (
	"algorithm/constants"
	"algorithm/types"
	"algorithm/util"
	"fmt"
	"math"
	"testing"
)

func TestCalculateApproachConflictZoneTimeSpeed(t *testing.T) {

	cases := []struct {
		v0 float64
		s float64
		vMax float64
	}{
		{0.0, 2.2980223533085677, 5.764170763289394},
		{0.0, 1.6755673168268572, 5.99577919604539},
		{0.0, 1.9659350727295077, 5.764170763289394},
		{0.0, 1.8495424474140592, 5.99577919604539},
		{1.6693282522641928, 5.434425318736443, 10.37550737392091},
		{10.549999999999988, 34.246, 15.0},
		{19, 40, 20},
		{2, 2, 4},
		{5, 2, 7},
	}

	for _, c := range cases {
		ts, v2, plan, success := calculateApproachConflictZoneTimeSpeed(c.v0, c.s, c.vMax)

		if success == false {
			t.FailNow()
		}

		fmt.Println("ts:", ts, ", v2:", v2)

		_ = v2
		_ = plan
		_ = ts
	}
}

func generateSimpleRoad() *util.Graph {
	n1 := util.Node{
		Id: 1,
		X: 50,
		Y: 0,
		EdgesFrom: []*util.Edge{},
		IsEntryPoint: true,
		WayId: 1,
	}
	n2 := util.Node{
		Id: 2,
		X: 50,
		Y: 40,
		EdgesFrom: []*util.Edge{},
	}
	n3 := util.Node{
		Id: 3,
		X: 50,
		Y: 60,
		EdgesFrom: []*util.Edge{},
	}
	n4 := util.Node{
		Id: 4,
		X: 50,
		Y: 100,
		EdgesFrom: []*util.Edge{},
		IsExitPoint: true,
		WayId: 3,
	}
	n1.EdgesFrom = append(n1.EdgesFrom, &util.Edge{
		Id: 1,
		From: &n1,
		To: &n2,
	})
	n2.EdgesFrom = append(n2.EdgesFrom, &util.Edge{
		Id: 2,
		From: &n2,
		To: &n3,
	})
	n3.EdgesFrom = append(n3.EdgesFrom, &util.Edge{
		Id: 3,
		From: &n3,
		To: &n4,
	})

	roadGraph := &util.Graph{
		AllNodes: []util.Node{n1,n2,n3,n4},
		ConflictZone: util.ConflictZone{
			MinX: 40,
			MinY: 40,
			MaxX: 60,
			MaxY: 60,
		},
	}

	return roadGraph
}

func TestStopsBeforeIntersection(t *testing.T) {
	roadGraph := generateSimpleRoad()
	p := AllVehiclesProxySingleton()
	s := SensorLayerSingleton(p, roadGraph)
	nc := CommunicationLayerSingleton(p, util.Configuration{})

	v := NewVehicleActor(1, 0, &roadGraph.AllNodes[0], &roadGraph.AllNodes[3], 10, roadGraph, s, nc)

	if len(v.route) == 0 {
		t.FailNow()
	}

	for ts := types.Millisecond(0); ts < 2 * 10e3; ts += constants.SimulationStepInterval {
		v.Ping(ts)
		if ts % 500 == 0 { // some debug msg:
			//fmt.Println("Speed:", v.Speed)
		}
	}

	if msgSentNo := len(nc.IntersectionManagerReceive(2 * 10e3)); msgSentNo == 0 {
		t.Error("Vehicle sent no requests")
	}

	if v.Speed != 0 {
		t.Error("Vehicle should have stopped before intersection")
	}
}


func TestFollowsReservation(t *testing.T) {
	//for n := 0; n < 100; n += 1{
	//	fmt.Println("n=", n)
		doTestFollowsReservationIngoreFirstNRequests(0, 2, t)
	//}
}

func doTestFollowsReservationIngoreFirstNRequests(n int, distToConflictZone types.Meter, t *testing.T) {
	roadGraph := generateSimpleRoad()
	p := AllVehiclesProxySingleton()
	s := SensorLayerSingleton(p, roadGraph)
	nc := CommunicationLayerSingleton(p, util.Configuration{})

	v := NewVehicleActor(1, 0, &roadGraph.AllNodes[0], &roadGraph.AllNodes[3], 10, roadGraph, s, nc)

	handleRequest := func (ts types.Millisecond, req DsrcV2RMessage) DsrcR2VMessage {
		const distOnConflictZone = 20.0
		tTotal := req.ApproachConflictZoneMinTs + types.Millisecond(distOnConflictZone / req.ApproachConflictZoneSpeed * 1000)
		if tTotal % 10 != 0 {
			tTotal = tTotal - tTotal % 10 + 10
		}

		plan := make(map[types.Millisecond]types.MetersPerSecond)
		for t := types.Millisecond(req.ApproachConflictZoneMinTs); t <= tTotal; t += constants.SimulationStepInterval {
			plan[t] = req.ApproachConflictZoneSpeed
		}

		response := DsrcR2VMessage{
			msgType:                 AimProtocolMsgAllow,
			reservationId:           1,
			tsSent:                  ts,
			receiver:                1,
			reservationFromTs:       req.ApproachConflictZoneMinTs,
			reservationToTs:         tTotal,
			reservationDesiredSpeed: req.ApproachConflictZoneSpeed,
			reservationTsToSpeed:    plan,
		}

		return response
	}

	reservationSent := false
	var response DsrcR2VMessage
	//counter := 0
	for ts := types.Millisecond(0); ts < 2 * 10e3; ts += constants.SimulationStepInterval {
		v.Ping(ts)

		if reservationSent {
			if v.State == beforeIntersectionNotAllowed {
				t.Fatal("Vehicle should have received and accepted reservation")
			}
		}

		if v.State == atIntersection {
			if math.Abs(v.Speed - response.reservationTsToSpeed[ts]) > 0.15 {
				if ts < response.reservationFromTs {
					if response.reservationFromTs - ts > 50 {
						t.Errorf("Vehicle arrived %d milliseconds to early (n = %d)", response.reservationFromTs - ts, n)
					}
				} else if ts > response.reservationToTs {
					if ts - response.reservationToTs > 50 {
						t.Errorf("Vehicle arrived %d milliseconds to early (n = %d)", response.reservationFromTs - ts, n)
					}
				} else {
					t.Error(fmt.Printf("v.Speed: %f, should be: %f, Reservation from: %d, to: %d, ts: %d, (n = %d)\n", v.Speed, response.reservationTsToSpeed[ts], response.reservationFromTs, response.reservationToTs, ts, n))
				}
			}
		}

		requests := nc.IntersectionManagerReceive(ts)

		if len(requests) > 0 {
			if v.Speed == 0 {
				response = handleRequest(ts, requests[0])
				nc.SendDsrcR2V(response)
				reservationSent = true
			}
			//if counter == n {
			//	response = handleRequest(ts, requests[0])
			//	nc.SendDsrcR2V(response)
			//	reservationSent = true
			//} else {
			//	counter += 1
			//}
		}

		if ts % 500 == 0 { // some debug msg:
			//fmt.Println("Speed:", v.Speed)
		}
	}
}



func TestPlatooningReservation(t *testing.T) {
	const v2AppearsTs = 1000
	const v1PlatooningReservationId = 42
	roadGraph := generateSimpleRoad()
	p := AllVehiclesProxySingleton()
	s := SensorLayerSingleton(p, roadGraph)
	nc := CommunicationLayerSingleton(p, util.Configuration{})

	v1 := NewVehicleActor(1, 0, &roadGraph.AllNodes[0], &roadGraph.AllNodes[3], 10, roadGraph, s, nc)
	v2 := NewVehicleActor(2, v2AppearsTs, &roadGraph.AllNodes[0], &roadGraph.AllNodes[3], 10, roadGraph, s, nc)

	p.RegisterVehicle(v1)
	p.RegisterVehicle(v2)

	handleRequest := func (vId types.VehicleId, ts types.Millisecond, req DsrcV2RMessage) DsrcR2VMessage {
		const distOnConflictZone = 20.0
		tTotal := req.ApproachConflictZoneMinTs + types.Millisecond(distOnConflictZone / req.ApproachConflictZoneSpeed * 1000)
		if tTotal % 10 != 0 {
			tTotal = tTotal - tTotal % 10 + 10
		}

		plan := make(map[types.Millisecond]types.MetersPerSecond)
		for t := types.Millisecond(req.ApproachConflictZoneMinTs); t <= tTotal; t += constants.SimulationStepInterval {
			plan[t] = req.ApproachConflictZoneSpeed
		}

		response := DsrcR2VMessage{
			msgType:                 AimProtocolMsgAllow,
			reservationId:           v1PlatooningReservationId,
			tsSent:                  ts,
			receiver:                vId,
			reservationFromTs:       req.ApproachConflictZoneMinTs,
			reservationToTs:         tTotal,
			reservationDesiredSpeed: req.ApproachConflictZoneSpeed,
			reservationTsToSpeed:    plan,
		}

		return response
	}

	v1ReponseReceived := false
	v2SentPlatooningRequest := false
	for ts := types.Millisecond(0); ts < 20 * 10e3; ts += constants.SimulationStepInterval {
		v1.Ping(ts)
		if ts >= v2AppearsTs {
			v2.Ping(ts)
		}

		if v1.State != beforeIntersectionNotAllowed {
			v1ReponseReceived = true
		}

		requests := nc.IntersectionManagerReceive(ts)
		for i := range requests {
			r := requests[i]
			if r.Sender == 1 {
				if r.MsgType == AimProtocolMsgReservationInfo {
					if v1ReponseReceived == false {
						t.Fatal("V1 sent broadcast before got reservation")
					}
				} else if r.MsgType == AimProtocolMsgRequest {
					if v1.Speed < 100 {
						v1Response := handleRequest(v1.Id, ts, r)
						nc.SendDsrcR2V(v1Response)
					}
				}
			}
			if r.Sender == 2 {
				if r.MsgType == AimProtocolMsgRequest {
					if (v1.State == beforeIntersectionNotAllowed || v1.State == beforeIntersectionHasReservation) && v1ReponseReceived == false {
						t.Fatal("V1 cannot send request without v2 xbroadcasting")
					}
					if v2SentPlatooningRequest && r.PlatooningReservationId != v1PlatooningReservationId {
						t.Error("Vehicle 2 should have send platooning reservation request")
					}
					v2SentPlatooningRequest = true

					v2Response := handleRequest(v2.Id, ts, r)
					nc.SendDsrcR2V(v2Response)
				}
			}
		}
	}

	if v1.State  != afterIntersection {
		t.Error(fmt.Sprintf("v1.State is: %d, should be: %d\n", v1.State, afterIntersection))
	}
	if v2.State != afterIntersection {
		t.Error(fmt.Sprintf("v2.State is: %d, should be: %d\n", v2.State, afterIntersection))
	}
	if v1ReponseReceived == false {
		t.Error("v1ReponseReceived is false")
	}
	if v2SentPlatooningRequest== false {
		t.Error("v2SentPlatooningRequest is false")
	}
}

func TestCalculateApproachConflictZonePlan(t *testing.T) {
	cases := []struct {
		t0 types.Millisecond
		v0 float64
		t2 types.Millisecond
		v2 float64
		s float64
		successExpected bool
	}{
		{0, 9, 5000, 11, 50,true},
		{0, 9, 5000, 11, 51,true},
		{0, 10, 4700, 10, 50,true},


		{0, 10, 5100, 10, 50,true},
		{0, 10, 5200, 10, 50,true},
	}

	for _, c := range cases {
		plan, success := calculateApproachConflictZonePlan_fixme(c.t0, c.v0, c.t2, c.v2, c.s)

		if success != c.successExpected {
			t.Error("Test case failed:", c)
		}

		if success {
			for ts := c.t0; ts < c.t2; ts += constants.SimulationStepInterval {

				if ts != c.t0 {
					before := plan[ts- constants.SimulationStepInterval]
					after := plan[ts]
					if after == before {

					} else if after > before {
						ok := before + velocityDiffStepAccelerating(before) >= after
						if ok == false {
							t.Error("Impossible acceleration (from:", before, ", to:", after, ")")
						}
					} else {
						ok := before - velocityDiffStepBraking(before) <= after
						if ok == false {
							t.Error("Impossible decelartion (from:", before, ", to:", after, ")")
						}
					}

					//fmt.Println(plan[ts])
				}
			}
		}

		_ = plan
	}
}
