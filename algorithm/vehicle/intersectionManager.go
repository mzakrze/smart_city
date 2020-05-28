package vehicle

import (
	"algorithm/types"
	"algorithm/util"
)

type IntersectionPolicy interface {
	ProcessMsg(message *DsrcV2RMessage)
	GetReplies(millisecond types.Millisecond) []*DsrcR2VMessage
}

var instance *IntersectionManager = nil
type IntersectionManager struct {
	networkCard *CommunicationLayer
	policy IntersectionPolicy
	nextAvailableTs types.Millisecond
}

func IntersectionManagerSingleton(graph *util.Graph, networkCard *CommunicationLayer, configuration util.Configuration) (*IntersectionManager, error) {
	if instance == nil {
		var policy IntersectionPolicy
		// TODO - use reflection to create instance (add method to IntersectionPolicy returing code)
		switch configuration.IntersectionPolicy {
		case "sequential":
			policy = CreateIntersectionPolicySequential(graph)
		case "fcfs":
			policy = CreateIntersectionPolicyFcfs(graph, configuration)
		default:
			panic("Illegal name of intersection policy")
		}

		instance = &IntersectionManager{networkCard: networkCard, policy: policy, nextAvailableTs: 0}
	}

	return instance, nil
}

func (im *IntersectionManager) Ping(ts types.Millisecond) {
	messages :=im.networkCard.IntersectionManagerReceive()

	if im.nextAvailableTs >= ts {
		return
	}

	for _, m := range messages {

		im.policy .ProcessMsg(m)

	}

	replies := im.policy.GetReplies(ts)
	for _, r := range replies {
		r.tsSent = ts
		im.networkCard.SendDsrcR2V(r)
	}

}




