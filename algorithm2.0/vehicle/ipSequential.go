package vehicle

import (
	"algorithm2.0/constants"
	"algorithm2.0/types"
	"algorithm2.0/util"
	"fmt"
	"math"
)

type IntersectionPolicySequential struct {
	winningRequest  *DsrcV2RMessage
	vehicleToFirstRequestTs map[types.VehicleId]types.Millisecond
	graph                   *util.Graph
	nextReservationId		types.ReservationId
	reservations			[]ipReservation

}


type ipReservation struct {
	from types.Millisecond
	to types.Millisecond
	id types.ReservationId
}

type turnDirection int
const (
	left turnDirection = iota
	right
	straight
)

func CreateIntersectionPolicySequential(graph *util.Graph) *IntersectionPolicySequential {
	return &IntersectionPolicySequential{
		winningRequest: nil,
		graph: graph,
		nextReservationId: 1,
		vehicleToFirstRequestTs: make(map[types.VehicleId]types.Millisecond),
		reservations: make([]ipReservation, 0),
	}
}

func (ip * IntersectionPolicySequential) ProcessMsg(m DsrcV2RMessage) {

	if m.MsgType == AimProtocolMsgReservationCancelation {
		index := math.MinInt32
		for i := range ip.reservations {
			if ip.reservations[i].id == m.ReservationToCancelId {
				index = i
			}
		}
		if index < 0 {
			panic("Oops")
		}
		ip.reservations = append(ip.reservations[:index], ip.reservations[index + 1 : ]...)

		return
	}

	if _, exists := ip.vehicleToFirstRequestTs[m.Sender]; exists == false {
		ip.vehicleToFirstRequestTs[m.Sender] = m.TsSent
	}

	arriveTs, leaveTs, _, _ := ip.calculateRouteForRequest(&m)

	if ip.isReserved(arriveTs, leaveTs) {
		return
	}

	if ip.winningRequest != nil && ip.scoreRequest(ip.winningRequest) < ip.scoreRequest(&m) {

		if m.PlatooningReservationId > 0 {
			fmt.Println("PlatooningReservationId rejected, because there is better offer")
		}

		return
	}

	ip.winningRequest = &m
}


func (ip *IntersectionPolicySequential) GetReplies(ts types.Millisecond) []*DsrcR2VMessage {
	//if ip.nextAvailableTs - ts > 0 {
	//	// we still have time to make a decision
	//	return []*DsrcR2VMessage{}
	//}
	if ip.winningRequest == nil {
		return []*DsrcR2VMessage{}
	}
	req := ip.winningRequest
	ip.winningRequest = nil

	arriveTs, leaveTs, arriveSpeed, tsToSpeed := ip.calculateRouteForRequest(req)

	if ip.isReserved(arriveTs, leaveTs) {
		panic("Sth not ok")
	}

	ip.reservations = append(ip.reservations, ipReservation{
		from: arriveTs,
		to: leaveTs,
		id: ip.nextReservationId,
	})

	ip.assertReservationsDontOverlap()

	offer := &DsrcR2VMessage{
		msgType: AimProtocolMsgAllow,
		receiver: req.Sender,
		reservationFromTs: arriveTs,
		reservationToTs: leaveTs,
		reservationDesiredSpeed: arriveSpeed,
		reservationTsToSpeed: tsToSpeed,
		reservationId: ip.nextReservationId,
	}
	ip.nextReservationId += 1

	res := []*DsrcR2VMessage{offer}
	return res
}


func (ip *IntersectionPolicySequential) getTurnDirection(message *DsrcV2RMessage) turnDirection {
	entry, exit := ip.getPointerByIdEntryExit(message.EntryPointId, message.ExitPointId)
	diff := entry.WayId - exit.WayId

	if diff == -1 || diff == 3 {
		return left
	} else if diff == 1 || diff == -3 {
		return right
	} else if diff == -2 || diff == 2 {
		return straight
	} else {
		panic("Illegal way")
	}
}

func (ip *IntersectionPolicySequential) getPointerByIdEntryExit (entryId, exitId types.NodeId) (*util.Node, *util.Node) {
	var entry *util.Node = nil
	var exit *util.Node = nil
	for i, e := range ip.graph.Entrypoints {
		if e.Id == entryId {
			entry = ip.graph.Entrypoints[i]
		}
	}
	for i, e := range ip.graph.Exitpoints {
		if e.Id == exitId {
			exit = ip.graph.Exitpoints[i]
		}
	}
	if entry == nil { panic("oops") }
	if exit == nil { panic("oops") }

	return entry, exit
}

func (ip *IntersectionPolicySequential) getTurnRadius(msg *DsrcV2RMessage) types.Meter {
	x := math.Abs(msg.ConflictZoneNodeEnter.X - msg.ConflictZoneNodeExit.X)
	return x
}

/**
The higher, the better (IP will more likely grant reservation)
 */
func (ip *IntersectionPolicySequential) scoreRequest(message *DsrcV2RMessage) float64 {
	res := 0.0

	res = -float64(message.ApproachConflictZoneMinTs)

	return res
}

func (ip *IntersectionPolicySequential) distanceOnConflictZone(msg *DsrcV2RMessage) types.Meter {
	switch ip.getTurnDirection(msg) {
	case straight:
		x := msg.ConflictZoneNodeEnter.X - msg.ConflictZoneNodeExit.X
		y := msg.ConflictZoneNodeEnter.Y - msg.ConflictZoneNodeExit.Y

		return math.Sqrt(x * x + y * y)
	case left:
		fallthrough
	case right:
		return 2.0 * math.Pi * ip.getTurnRadius(msg) / 4.0
	default:
		panic("oops")
	}
}

func (ip *IntersectionPolicySequential) calculateRouteForRequest(req *DsrcV2RMessage) (types.Millisecond, types.Millisecond, types.MetersPerSecond, map[types.Millisecond]types.MetersPerSecond) {
	approachConflictZoneSpeed := req.ApproachConflictZoneSpeedMax
	approachConflictZoneTs := req.ApproachConflictZoneMinTs + constants.SimulationStepInterval

	reservationTsToSpeed := make(map[types.Millisecond]types.MetersPerSecond)
	tsOnIntersection := approachConflictZoneTs
	speed := approachConflictZoneSpeed
	accelerateTo := func(limit types.MetersPerSecond) {
		if speed < limit {
			speed += maxAcc * (float64(constants.SimulationStepInterval) / 1000.0)
			if speed > limit {
				speed = limit
			}
		}
	}
	distanceInConflictZone := ip.distanceOnConflictZone(req)
	for d := types.Meter(0); d < distanceInConflictZone; tsOnIntersection += constants.SimulationStepInterval {
		reservationTsToSpeed[tsOnIntersection] = speed

		d += speed * (float64(constants.SimulationStepInterval) / 1000.0)

		if req.IsTurning {
			accelerateTo(req.MaxSpeedOnCurve)
		} else {
			accelerateTo(maxSpeed)
		}
	}

	return approachConflictZoneTs, tsOnIntersection, approachConflictZoneSpeed, reservationTsToSpeed
}

func (ip *IntersectionPolicySequential) isReserved(from types.Millisecond, to types.Millisecond) bool {
	for _, r := range ip.reservations {
		if r.from <= from && from <= r.to {
			return true
		}
		if r.from <= to && to <= r.to {
			return true
		}
		if from <= r.from && r.to <= to {
			return true
		}
	}

	return false
}

func (ip *IntersectionPolicySequential) assertReservationsDontOverlap() {

	uniqueId := make(map[types.ReservationId]bool)
	for _, r := range ip.reservations {
		_, exists := uniqueId[r.id]
		if exists {
			panic("Id is not unique")
		}
		uniqueId[r.id] = true
	}

	for _, r1 := range ip.reservations {
		for _, r2 := range ip.reservations {
			if r1.id == r2.id {
				continue
			}

			if r1.from <= r2.from && r2.from <= r1.to {
				panic("Overlap")
			}
			if r1.from <= r2.to && r2.to <= r1.to {
				panic("Overlap")
			}
		}
	}
}