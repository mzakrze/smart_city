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
	return Configuration{
		VehiclesPerMinute: 120,
		IntersectionPolicy: "fcfs",
		SimulationDuration: 30,
		VehicleMaxSpeed: 10,
		RandomSeed: 42,
		VehiclePower: 100,
		VehicleBrakingForce: 4000.0,
		VehicleWeight: 1500.0,
		VehicleMaxAngularSpeed: 0.4,
		VehicleMaxSpeedOnConflictZone: 15.0,
		DsrcMsgLossProbability: 0.0,
		DsrcMsgAvgDelay: 0,
	}, nil
}
