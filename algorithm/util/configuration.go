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
}

func ReadConfiguration(filePath string) (Configuration, error) {
	// TODO read from file
	return Configuration{
		VehiclesPerMinute: 120,
		IntersectionPolicy: "fcfs",
		SimulationDuration: 120,
		VehicleMaxSpeed: 10,
		RandomSeed: 42,
		VehiclePower: 130,
		VehicleBrakingForce: 3000.0,
		VehicleWeight: 1200.0,
		VehicleMaxAngularSpeed: 0.4,
		VehicleMaxSpeedOnConflictZone: 15.0,
	}, nil
}
