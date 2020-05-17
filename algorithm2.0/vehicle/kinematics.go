package vehicle

import "algorithm2.0/types"

/**
Find deceleration to brake from initial speed v to 0 in given time t
 */
func brakingDecel(v types.MetersPerSecond, t types.Millisecond) types.MetersPerSecond2 {
	return -float64(v) / (float64(t) / 1000.0)
}
