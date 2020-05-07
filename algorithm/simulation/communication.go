package simulation

import (
	"math"
	"math/rand"
	"mzakrze/smart_city/types"
)

func deliveryRatioV2VByDistance(dist types.Meter) int {
	if dist < 25 {
		return 100
	} else if dist < 450 {
		// delivery ratio drops linearly from 1 to 0.8 on distance from 25m to 450m
		return int(math.Round(100 - 20 * (450 - dist) / 450))
	} else {
		// then, delivery ratio drops to 0 at around 700 m
		return int(math.Max(0, 80 - 80 * (1- (700 - dist) / 700)))
	}
}

func latencyV2VByDistance(dist types.Meter) types.Milisecond {
	const min = 15
	const max = 25

	if dist > 700 {
		return max
	}

	return int(min + (max - min) * (700 - dist) / 700)
}

func randV2VLatencyAndDrop(dist types.Meter) (types.Milisecond, bool) {
	if dist < 0 { panic("illegal argument") }

	dropRatio := deliveryRatioV2VByDistance(dist)

	if rand.Intn(100) > dropRatio {
		// this means we loose this message :(
		return 0, false
	}

	latency := latencyV2VByDistance(dist)

	return latency, false
}
