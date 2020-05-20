package vehicle

import (
	"algorithm2.0/constants"
	"algorithm2.0/types"
	"algorithm2.0/util"
	"fmt"
	"math"
)

type IntersectionPolicySequential struct {
	nextAvailableTs types.Millisecond
	winningRequest  *DsrcV2RMessage

	vehicleToFirstRequestTs map[types.VehicleId]types.Millisecond
	graph                   *util.Graph
}

type turnDirection int
const (
	left turnDirection = iota
	right
	straight
)

func CreateIntersectionPolicySequential(graph *util.Graph) *IntersectionPolicySequential {
	return &IntersectionPolicySequential{
		nextAvailableTs: 0,
		winningRequest: nil,
		graph: graph,
		vehicleToFirstRequestTs: make(map[types.VehicleId]types.Millisecond),
	}
}

func (ip * IntersectionPolicySequential) ProcessMsg(m DsrcV2RMessage) {
	if _, exists := ip.vehicleToFirstRequestTs[m.Sender]; exists == false {
		ip.vehicleToFirstRequestTs[m.Sender] = m.TsSent
	}

	if ip.nextAvailableTs >= m.ApproachConflictZoneMinTs {
		return
	}

	if ip.winningRequest != nil && ip.scoreRequest(ip.winningRequest) > ip.scoreRequest(&m) {
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

	approachConflictZoneSpeed := req.ApproachConflictZoneSpeedMax
	approachConflictZoneTs := req.ApproachConflictZoneMinTs

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
	distanceInConflictZone := ip.distanceOnConflictZone(req) + constants.VehicleLength
	for d := types.Meter(0); d < distanceInConflictZone; tsOnIntersection += constants.SimulationStepInterval {
		reservationTsToSpeed[tsOnIntersection] = speed

		d += speed * (float64(constants.SimulationStepInterval) / 1000.0)

		if req.IsTurning {
			accelerateTo(req.MaxSpeedOnCurve)
		} else {
			accelerateTo(maxSpeed)
		}
	}
	ip.nextAvailableTs = tsOnIntersection
	fmt.Println(ip.nextAvailableTs)

	offer := &DsrcR2VMessage{
		msgType: AimProtocolMsgAllow,
		receiver: req.Sender,
		reservationFromTs: approachConflictZoneTs,
		reservationDesiredSpeed: approachConflictZoneSpeed,
		reservationTsToSpeed: reservationTsToSpeed,
	}

	res := []*DsrcR2VMessage{offer}
	ip.winningRequest = nil
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

	// punishment for late arrival
	res -= float64(message.ApproachConflictZoneMinTs)

	//res += float64(ip.vehicleToFirstRequestTs[message.Sender])

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