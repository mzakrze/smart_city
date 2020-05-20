package vehicle

import (
	"algorithm2.0/types"
)

type IntersectionPolicySequential struct {
	nextAvailableTs types.Millisecond
	winningRequest *DsrcV2RMessage

	vehicleToFirstRequestTs map[types.VehicleId]types.Millisecond
}

func CreateIntersectionPolicySequential() *IntersectionPolicySequential {
	return &IntersectionPolicySequential{
		nextAvailableTs: 0,
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

	if ip.winningRequest != nil && ip.scoreRequest(ip.winningRequest) > ip.scoreRequest(&m) {
		return
	}

	ip.winningRequest = &m
}

func (ip *IntersectionPolicySequential) GetReplies(ts types.Millisecond) []*DsrcR2VMessage {
	if ip.nextAvailableTs - ts > 0 {
		// we still have time to make a decision
		return []*DsrcR2VMessage{}
	}
	if ip.winningRequest == nil {
		return []*DsrcR2VMessage{}
	}

	req := ip.winningRequest
	offer := &DsrcR2VMessage{
		msgType: AimProtocolMsgAllow,
		receiver: req.Sender,
		reservationFromTs: req.ApproachConflictZoneTs,
		reservationToTs: req.LeaveConflictZoneTs,
		reservationDesiredSpeed: req.ApproachConflictZoneSpeedMax,
	}

	ip.nextAvailableTs = req.LeaveConflictZoneTs

	res := []*DsrcR2VMessage{offer}
	ip.winningRequest = nil
	return res
}

/**
The higher, the better (IP will more likely grant reservation)
 */
func (ip *IntersectionPolicySequential) scoreRequest(message *DsrcV2RMessage) float64 {
	res := 0.0

	// punishment for late arrival
	res -= float64(message.ApproachConflictZoneTs)

	//res += float64(ip.vehicleToFirstRequestTs[message.Sender])

	return res
}