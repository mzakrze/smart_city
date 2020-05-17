package vehicle

import (
	"algorithm2.0/types"
	"algorithm2.0/vehicle/intersection_policies"
)

type IntersectionPolicy interface {

}

var instance *IntersectionManager = nil
type IntersectionManager struct {
	networkCard *CommunicationLayer
	policy *IntersectionPolicy
}

func IntersectionManagerSingleton(networkCard *CommunicationLayer, intersectionPolicyId string) (*IntersectionManager, error) {
	if instance == nil {
		var policy IntersectionPolicy
		// TODO - use reflection to create instance (add method to IntersectionPolicy returing code)
		switch intersectionPolicyId {
		case "sequential":
			policy = &intersection_policies.IntersectionPolicySequential{}
		default:
			panic("Illegal name of intersection policy")
		}

		instance = &IntersectionManager{networkCard: networkCard, policy: &policy}
	}

	return instance, nil
}

func (im *IntersectionManager) Ping(ts types.Millisecond) {
	im.networkCard.IntersectionManagerReceive()
}




