package vehicle

import (
	"algorithm2.0/types"
	"algorithm2.0/vehicle/intersection_policies"
	"algorithm2.0/vehicle_communication"
)

type IntersectionPolicy interface {

}

var instance *IntersectionManager = nil
type IntersectionManager struct {
	networkCard *vehicle_communication.CommunicationLayer
	policy *IntersectionPolicy
}

func IntersectionManagerSingleton(networkCard *vehicle_communication.CommunicationLayer, intersectionPolicyId string) (*IntersectionManager, error) {
	if instance == nil {
		var policy IntersectionPolicy
		// TODO - use reflection to create instance (add method to IntersectionPolicy returing code)
		switch intersectionPolicyId {
		case "sequential":
			policy = &intersection_policies.IntersectionPolicySequential{}
		}

		instance = &IntersectionManager{networkCard: networkCard, policy: &policy}
	}

	return instance, nil
}

func (im *IntersectionManager) Ping(ts types.Millisecond) {

}




