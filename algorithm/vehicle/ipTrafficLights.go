package vehicle

import (
	"algorithm/constants"
	"algorithm/types"
	"algorithm/util"
	"math"
)



type IntersectionPolicyTrafficLights struct {
	//state:
	currentWayId          types.WayId
	currentWayStartedAt   types.Millisecond
	isWaitPhase           bool
	waitStartedAt         types.Millisecond
	nextReservationId     types.ReservationId
	replies               []DsrcR2VMessage
	prevApproachTs        types.Millisecond
	prevApproachPerLineTs map[int]types.Millisecond
	// configuration:
	millisecondsFree types.Millisecond
	millisecondWait  types.Millisecond
	nodeIdToLineId   map[types.NodeId]int
	nodeIdToWayId         map[types.NodeId]types.WayId
}

func CreateIntersectionPolicyTrafficLights(graph *util.Graph, configuration util.Configuration) *IntersectionPolicyTrafficLights {
	if configuration.DsrcMsgAvgDelay != 0 {
		panic("Illegal DsrcMsgAvgDelay for traffic lights (must be 0)")
	}
	if configuration.DsrcMsgLossProbability != 0 {
		panic("Illegal DsrcMsgLossProbability for traffic lights (must be 0)")
	}

	nodeIdToWayId := make(map[types.NodeId]types.WayId)
	nodeIdToLineId := make(map[types.NodeId]int)
	prevApproachPerLineTs := make(map[int]types.Millisecond)
	for i := range graph.Entrypoints {
		nodeIdToWayId[graph.Entrypoints[i].Id] = graph.Entrypoints[i].WayId
		nodeIdToLineId[graph.Entrypoints[i].Id] = graph.Entrypoints[i].EntryPointId
		prevApproachPerLineTs[graph.Entrypoints[i].EntryPointId] = types.Millisecond(0)
	}

	instance := IntersectionPolicyTrafficLights{
		currentWayId: 1,
		currentWayStartedAt: 0,
		isWaitPhase: false,
		waitStartedAt: 0,
		nodeIdToWayId: nodeIdToWayId,
		nodeIdToLineId: nodeIdToLineId,
		nextReservationId: 1,
		prevApproachPerLineTs: prevApproachPerLineTs,
	}

	instance.millisecondsFree = 8000
	instance.millisecondWait = 3000

	return &instance
}

func (ip * IntersectionPolicyTrafficLights) ProcessMsg(m DsrcV2RMessage) {
	if ip.isWaitPhase {
		return
	}
	if ip.currentWayId != ip.nodeIdToWayId[m.EntryPointId] {
		return
	}

	if ip.prevApproachPerLineTs[ip.nodeIdToLineId[m.EntryPointId]] + 1300 > m.ApproachConflictZoneMinTs {
		return
	}

	vMax := vehicleMaxSpeedOnConflictZone
	if m.IsTurning {
		vMax = m.MaxSpeedOnCurve
	}

	dist := 0.0
	for i := 1; i < len(m.Route); i++ {
		x := m.Route[i-1].X - m.Route[i].X
		y := m.Route[i-1].Y - m.Route[i].Y

		dist += math.Sqrt(x * x + y * y)
	}
	reservationTsToSpeed := make(map[types.Millisecond]types.MetersPerSecond)

	v := m.ApproachConflictZoneSpeed
	ts := m.ApproachConflictZoneMinTs
	approachConflictZoneTs := ts
	approachConflictZoneSpeed := v

	accelerateVehicle := func () {
		if v == vMax {
			return
		}

		diff := velocityDiffStepAccelerating(v)
		if v + diff >= vMax {
			v = vMax
		} else if v + diff < vMax {
			v += diff
		} else {
			panic("Oops")
		}
	}

	for ; dist > 0 ; ts += constants.SimulationStepInterval {
		reservationTsToSpeed[ts] = v
		accelerateVehicle()
		dist -= v * float64(constants.SimulationStepInterval) / 1000.0
	}

	reply := DsrcR2VMessage{
		reservationId: ip.nextReservationId,
		msgType: AimProtocolMsgAllow,
		receiver: m.Sender,
		reservationFromTs: approachConflictZoneTs,
		reservationToTs: ts,
		reservationDesiredSpeed: approachConflictZoneSpeed,
		reservationTsToSpeed: reservationTsToSpeed,
	}

	ip.prevApproachTs = approachConflictZoneTs
	ip.prevApproachPerLineTs[ip.nodeIdToLineId[m.EntryPointId]] = approachConflictZoneTs
	ip.nextReservationId += 1
	ip.replies = append(ip.replies, reply)

}

func (ip *IntersectionPolicyTrafficLights) GetReplies(ts types.Millisecond) []DsrcR2VMessage {
	if ip.isWaitPhase && ts - ip.waitStartedAt > ip.millisecondWait {
		ip.isWaitPhase = false
		ip.currentWayStartedAt = ts
		ip.currentWayId = (ip.currentWayId + 1) % 5
		if ip.currentWayId == 0 {
			ip.currentWayId = 1 // efekt indeksowania od 1, trzeba dodatkowy warunek
		}
	}
	if ip.isWaitPhase == false && ts - ip.currentWayStartedAt >= ip.millisecondsFree {
		ip.isWaitPhase = true
		ip.waitStartedAt = ts
	}

	res := ip.replies
	ip.replies = []DsrcR2VMessage{}
	return res
}