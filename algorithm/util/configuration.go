package util

import (
	"algorithm/types"
)

type Configuration struct {
	VehiclesPerMinute             int32
	IntersectionPolicy            string
	SimulationDuration            int
	VehicleMaxSpeed               types.MetersPerSecond
	RandomSeed                    int64
	VehiclePower                  float64
	VehicleBrakingForce           float64
	VehicleWeight                float64
	VehicleMaxAngularSpeed        float64
	VehicleMaxSpeedOnConflictZone float64
	DsrcMsgLossProbability		  float64
	DsrcMsgAvgDelay				  int
}

func ReadConfiguration(filePath string) (Configuration, error) {
	// TODO read from file
	c := Configuration{
		VehiclesPerMinute: 120,
		IntersectionPolicy: "trafficLights",
		SimulationDuration: 60,
		VehicleMaxSpeed: 10,
		RandomSeed: 42,
		VehiclePower: 80,
		VehicleBrakingForce: 4000.0,
		VehicleWeight: 1500.0,
		VehicleMaxAngularSpeed: 0.4,
		VehicleMaxSpeedOnConflictZone: 15.0,
		DsrcMsgLossProbability: 0.0,
		DsrcMsgAvgDelay: 0,
	}

	if c.DsrcMsgAvgDelay != 0 {
		panic("Illegal DsrcMsgAvgDelay for traffic lights (must be 0)")
	}
	if c.DsrcMsgLossProbability != 0 {
		panic("Illegal DsrcMsgLossProbability for traffic lights (must be 0)")
	}


	return c, nil

}
