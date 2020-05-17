package vehicle

import (
	"algorithm2.0/types"
)

type IntersectionPolicySequential struct {
	nextAvailableTs types.Millisecond
	bestOffer *DsrcR2VMessage
	winningRequest *DsrcV2RMessage

	vehicleToFirstRequestTs map[types.VehicleId]types.Millisecond
}

func CreateIntersectionPolicySequential() *IntersectionPolicySequential {
	return &IntersectionPolicySequential{
		nextAvailableTs: 0,
		bestOffer: nil,
		winningRequest: nil,
		vehicleToFirstRequestTs: make(map[types.VehicleId]types.Millisecond),
	}
}

func (ip * IntersectionPolicySequential) ProcessMsg(m DsrcV2RMessage) {
	if _, exists := ip.vehicleToFirstRequestTs[m.Sender]; exists == false {
		ip.vehicleToFirstRequestTs[m.Sender] = m.TsSent
	}

	if ip.nextAvailableTs >= m.ApproachConflictZoneTs {
		return
	}

	from := m.ApproachConflictZoneTs
	to := m.LeaveConflictZoneTs

	if ip.bestOffer != nil && ip.evaluateRequest(ip.winningRequest) > ip.evaluateRequest(&m) {
		return
	}

	ip.nextAvailableTs = to

	reply := DsrcR2VMessage{
		msgType: AimProtocolMsgAllow,
		receiver: m.Sender,
		reservationFromTs: from,
		reservationToTs: to,
		reservationDesiredSpeed: m.ApproachConflictZoneSpeedMax,
	}

	ip.bestOffer = &reply
	ip.winningRequest = &m
}

func (ip *IntersectionPolicySequential) GetReplies() []*DsrcR2VMessage {
	if ip.bestOffer == nil {
		return []*DsrcR2VMessage{}
	}
	res := []*DsrcR2VMessage{ip.bestOffer}
	ip.bestOffer = nil
	return res
}

func (ip *IntersectionPolicySequential) evaluateRequest(message *DsrcV2RMessage) float64 {
	res := 0.0

	res -= 10.0 * float64(message.ApproachConflictZoneTs)

	//res += float64(ip.vehicleToFirstRequestTs[message.Sender])

	return res
}