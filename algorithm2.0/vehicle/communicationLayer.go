package vehicle

import (
	"algorithm2.0/types"
)

type AimProtocolMessageType = int32
const (
	AimProtocolMsgRequest = iota
	AimProtocolMsgAllow
)

type DsrcV2RMessage struct {
	MsgType AimProtocolMessageType
	TsSent types.Millisecond
	Sender types.VehicleId
	IsFirstOnIntersection bool
	VehicleInFrontId types.VehicleId
	X types.XCoord
	Y types.YCoord
	Speed types.MetersPerSecond
	Acc types.MetersPerSecond2
}

type DsrcR2VMessage struct {
	msgType AimProtocolMessageType
	tsSent types.Millisecond
	receiver types.VehicleId
}

func CommunicationLayerSingleton(proxy *AllVehicleProxy) *CommunicationLayer {
	if instanceCommunication == nil {
		instanceCommunication = &CommunicationLayer{proxy: proxy}
	}
	return instanceCommunication
}

func (c *CommunicationLayer) SendDsrcV2R(m DsrcV2RMessage) {
	// FIXME - dodać opóźnienie + gubienie
	c.imToReceive = append(c.imToReceive, m)
}

func (c *CommunicationLayer) SendDsrcR2V(m DsrcR2VMessage) {
	// FIXME - dodać opóźnienie + gubienie
	if _, exists := c.vehicleToReceive[m.receiver]; exists == false {
		c.vehicleToReceive[m.receiver] = []DsrcR2VMessage{}
	}
	c.vehicleToReceive[m.receiver] = append(c.vehicleToReceive[m.receiver], m)
}

func (c *CommunicationLayer) VehicleReceive(id types.VehicleId) []DsrcR2VMessage {
	queue := c.vehicleToReceive[id]
	delete(c.vehicleToReceive, id)
	return queue
}

func (c *CommunicationLayer) IntersectionManagerReceive() []DsrcV2RMessage {
	queue := c.imToReceive
	c.imToReceive = []DsrcV2RMessage{}
	return queue
}


var instanceCommunication *CommunicationLayer = nil
type CommunicationLayer struct {
	proxy *AllVehicleProxy
	vehicleToReceive map[types.VehicleId][]DsrcR2VMessage
	imToReceive []DsrcV2RMessage
}




