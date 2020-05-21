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
	VehicleMaxAcc types.MetersPerSecond2
	VehicleMaxDecel types.MetersPerSecond2
	VehicleMaxSpeed types.MetersPerSecond
}



func ReadConfiguration() (Configuration, error) {
	// TODO mocked
	d, _ := time.ParseDuration("60s")
	return Configuration{
		SimulationName: "test1",
		VehiclesPerMinute: 60,
		IntersectionPolicy: "fcfs",
		SimulationDuration: d,
		VehicleMaxSpeed: 10,
		VehicleMaxAcc: 2.0,
		VehicleMaxDecel: 3.5,
	}, nil
}
