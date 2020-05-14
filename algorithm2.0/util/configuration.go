package util

import (
	"algorithm2.0/types"
	"time"
)

const filePath = "../simulation.conf"

type Configuration struct {
	SimulationName string
	VehiclesPerMinute int32
	IntersectionPolicy string
	SimulationDuration time.Duration
	StepInterval types.Millisecond
	VehicleMaxAcc types.MetersPerSecond2
	VehicleMaxDecel types.MetersPerSecond2
	VehicleMaxSpeed types.MetersPerSecond
}



func ReadConfiguration() (Configuration, error) {
	return Configuration{}, nil
}
