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
	for n := 0; n < 100; n += 1{
		doTestFollowsReservationIngoreFirstNRequests(n, t)
	}
}

func doTestFollowsReservationIngoreFirstNRequests(n int, t *testing.T) {
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
	counter := 0
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
			if counter == n {
				response = handleRequest(ts, requests[0])
				nc.SendDsrcR2V(response)
				reservationSent = true
			} else {
				counter += 1
			}
		}

		if ts % 500 == 0 { // some debug msg:
			//fmt.Println("Speed:", v.Speed)
		}
	}
}
