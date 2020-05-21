package vehicle

import (
	"algorithm2.0/types"
	"algorithm2.0/util"
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

func CommunicationLayerSingleton(proxy *AllVehicleProxy) *CommunicationLayer {
	if instanceCommunication == nil {
		instanceCommunication = &CommunicationLayer{
			proxy: proxy,
			vehicleToReceive: make(map[types.VehicleId][]*DsrcR2VMessage),
			vehicleToReceiveV2V: make(map[types.VehicleId][]*DsrcV2VMessage),
		}
	}
	return instanceCommunication
}

func (c *CommunicationLayer) SendDsrcV2V(m *DsrcV2VMessage) {
	for _, v := range c.proxy.GetAllVehicles() {
		if v.Id == m.sender {
			continue
		}
		c.vehicleToReceiveV2V[v.Id] = append(c.vehicleToReceiveV2V[v.Id], m)
	}
}

func (c *CommunicationLayer) SendDsrcV2R(m *DsrcV2RMessage) {
	// FIXME - dodać opóźnienie + gubienie
	c.imToReceive = append(c.imToReceive, m)
}

func (c *CommunicationLayer) SendDsrcR2V(m *DsrcR2VMessage) {
	// FIXME - dodać opóźnienie + gubienie
	if _, exists := c.vehicleToReceive[m.receiver]; exists == false {
		c.vehicleToReceive[m.receiver] = []*DsrcR2VMessage{}
	}
	c.vehicleToReceive[m.receiver] = append(c.vehicleToReceive[m.receiver], m)
}



func (c *CommunicationLayer) VehicleReceive(id types.VehicleId) []*DsrcR2VMessage {
	queue := c.vehicleToReceive[id]
	delete(c.vehicleToReceive, id)
	return queue
}

func (c *CommunicationLayer) VehicleReceiveV2V(id types.VehicleId) []*DsrcV2VMessage {
	queue := c.vehicleToReceiveV2V[id]
	delete(c.vehicleToReceiveV2V, id)
	return queue
}

func (c *CommunicationLayer) IntersectionManagerReceive() []*DsrcV2RMessage {
	queue := c.imToReceive
	rand.Shuffle(len(queue), func(i, j int) {
		el := queue[i]
		queue[i] = queue[j]
		queue[j] = el

	})
	c.imToReceive = []*DsrcV2RMessage{}
	return queue
}


var instanceCommunication *CommunicationLayer = nil
type CommunicationLayer struct {
	proxy *AllVehicleProxy
	vehicleToReceive map[types.VehicleId][]*DsrcR2VMessage
	vehicleToReceiveV2V map[types.VehicleId][]*DsrcV2VMessage
	imToReceive []*DsrcV2RMessage
}




