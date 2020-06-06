package vehicle

import (
	"algorithm/types"
	"algorithm/util"
	"math/rand"
)

type AimProtocolMessageType = int32
const (
	AimProtocolMsgRequest = iota
	AimProtocolMsgReservationCancelation
	AimProtocolMsgAllow
	AimProtocolMsgReservationInfo
)

type DsrcV2RMessage struct {
	MsgType                   AimProtocolMessageType
	ReservationToCancelId     types.ReservationId
	PlatooningReservationId   types.ReservationId
	TsSent                    types.Millisecond
	Sender                    types.VehicleId
	VehicleX                  types.XCoord
	VehicleY                  types.YCoord
	VehicleSpeed              types.MetersPerSecond
	ApproachConflictZoneMinTs types.Millisecond
	ApproachConflictZoneSpeed types.MetersPerSecond
	ConflictZoneNodeEnter     *util.Node
	ConflictZoneNodeExit      *util.Node
	MaxSpeedOnCurve           types.MetersPerSecond
	IsTurning                 bool
	EntryPointId              types.NodeId
	ExitPointId               types.NodeId
	Route                     []types.Location
}

type DsrcR2VMessage struct {
	msgType                 AimProtocolMessageType
	reservationId           types.ReservationId
	tsSent                  types.Millisecond
	receiver                types.VehicleId
	reservationFromTs       types.Millisecond
	reservationToTs         types.Millisecond
	reservationDesiredSpeed types.MetersPerSecond
	reservationTsToSpeed    map[types.Millisecond]types.MetersPerSecond
}

type DsrcV2VMessage struct {
	msgType                  AimProtocolMessageType
	sender                   types.VehicleId
	reservationId            types.ReservationId
	arriveConflictZoneTs     types.Millisecond
	arriveConflictZoneSpeed  types.MetersPerSecond
	approachConflictZonePlan map[types.Millisecond]types.MetersPerSecond
	tsSent                   types.Millisecond
	x                        types.XCoord
	y                        types.YCoord
}

type message struct {
	msgLayer string
	deliveryTs types.Millisecond
	vehicleReceiverId types.VehicleId
	msgv2r * DsrcV2RMessage
	msgr2v * DsrcR2VMessage
	msgv2v * DsrcV2VMessage
}


func CommunicationLayerSingleton(proxy *AllVehicleProxy, configuration util.Configuration) *CommunicationLayer {
	if instanceCommunication == nil {
		instanceCommunication = &CommunicationLayer{
			proxy:                   proxy,
			messages:                []message{},
			deliveryLossProbability: configuration.DsrcMsgLossProbability,
			deliveryAvgDelay:        types.Millisecond(configuration.DsrcMsgAvgDelay),
			statsLost:               0,
			statsDelivered:          0,
			statsSumDelay:           0,
		}
	}
	return instanceCommunication
}

func (c *CommunicationLayer) SendDsrcV2V(msg DsrcV2VMessage) {
	for _, v := range c.proxy.GetAllActiveVehicles() {
		if v.Id == msg.sender {
			continue
		}

		delay, deliver := c.randLDelayAndDeliveryUpdateStats()
		if deliver == false {
			return
		}
		m := message{msgv2v: &msg, msgLayer: "v2v", deliveryTs: msg.tsSent + delay, vehicleReceiverId: v.Id}
		c.messages = append(c.messages, m)
	}
}

func (c *CommunicationLayer) SendDsrcV2R(msg DsrcV2RMessage) {
	delay, deliver := c.randLDelayAndDeliveryUpdateStats()
	if deliver == false {
		return
	}
	m := message{msgv2r: &msg, msgLayer: "v2r", deliveryTs: msg.TsSent + delay}
	c.messages = append(c.messages, m)
}

func (c *CommunicationLayer) SendDsrcR2V(msg DsrcR2VMessage) {
	delay, deliver := c.randLDelayAndDeliveryUpdateStats()
	if deliver == false {
		return
	}
	m := message{msgr2v: &msg, msgLayer: "r2v", deliveryTs: msg.tsSent + delay, vehicleReceiverId: msg.receiver}
	c.messages = append(c.messages, m)
}

func (c *CommunicationLayer) GetStats() (int, int, types.Millisecond) {
	avgDelay := types.Millisecond(float64(c.statsSumDelay) / float64(c.statsDelivered))
	return c.statsDelivered, c.statsLost, avgDelay
}

func (c *CommunicationLayer) randLDelayAndDeliveryUpdateStats() (types.Millisecond, bool) {
	if float64(rand.Intn(100)) < c.deliveryLossProbability {
		c.statsLost += 1
		return 0, false
	}
	delay := types.Millisecond(rand.Float64() * 2.0 * float64(c.deliveryAvgDelay))
	c.statsDelivered += 1
	c.statsSumDelay += delay
	return delay, true
}

func (c *CommunicationLayer) VehicleReceive(ts types.Millisecond, id types.VehicleId) []DsrcR2VMessage {
	res := []DsrcR2VMessage{}
	tmp := c.messages[:0]
	for i := range c.messages {
		if c.messages[i].deliveryTs <= ts && c.messages[i].msgLayer == "r2v" && c.messages[i].vehicleReceiverId == id {
			res = append(res, *c.messages[i].msgr2v)
		} else {
			tmp = append(tmp, c.messages[i])
		}
	}
	c.messages = tmp
	return res
}

func (c *CommunicationLayer) VehicleReceiveV2V(ts types.Millisecond, id types.VehicleId) []DsrcV2VMessage {
	res := []DsrcV2VMessage{}
	//fmt.Println("Before: len(c.messages):", len(c.messages))
	tmp := c.messages[:0]
	for i := range c.messages {
		if c.messages[i].deliveryTs <= ts && c.messages[i].msgLayer == "v2v" && c.messages[i].vehicleReceiverId == id {
			res = append(res, *c.messages[i].msgv2v)
		} else {
			tmp = append(tmp, c.messages[i])
		}
	}
	c.messages = tmp
	//fmt.Println("After: len(c.messages):", len(c.messages))
	return res
}

func (c *CommunicationLayer) IntersectionManagerReceive(ts types.Millisecond) []DsrcV2RMessage {
	res := []DsrcV2RMessage{}
	tmp := c.messages[:0]
	for i := range c.messages {
		if c.messages[i].deliveryTs <= ts && c.messages[i].msgLayer == "v2r" {
			res = append(res, *c.messages[i].msgv2r)
		} else {
			tmp = append(tmp, c.messages[i])
		}
	}
	c.messages = tmp
	return res
}


var instanceCommunication *CommunicationLayer = nil
type CommunicationLayer struct {
	proxy                   *AllVehicleProxy
	messages                []message
	deliveryLossProbability float64
	deliveryAvgDelay        types.Millisecond
	statsLost               int
	statsDelivered          int
	statsSumDelay           types.Millisecond
}


//func deliveryRatioV2VByDistance(dist types.Meter) int {
//	if dist < 25 {
//		return 100
//	} else if dist < 450 {
//		// delivery ratio drops linearly from 1 to 0.8 on distance from 25m to 450m
//		return int(math.Round(100 - 20 * (450  - dist) / 450))
//	} else {
//		// then, delivery ratio drops to 0 at around 700 m
//		return int(math.Max(0, 80 - 80 * (1- (700 - dist) / 700)))
//	}
//}
//
//func latencyV2VByDistance(dist types.Meter) types.Millisecond {
//	const min = 15
//	const max = 25
//
//	if dist > 700 {
//		return max
//	}
//
//	return types.Millisecond(min + (max - min) * (700 - dist) / 700)
//}
//
//func randV2VLatencyAndDrop(dist types.Meter) (types.Millisecond, bool) {
//	if dist < 0 { panic("illegal argument") }
//
//	dropRatio := deliveryRatioV2VByDistance(dist)
//
//	if rand.Intn(100) > dropRatio {
//		// this means we loose this message :(
//		return 0, false
//	}
//
//	latency := latencyV2VByDistance(dist)
//
//	return latency, false
//}

