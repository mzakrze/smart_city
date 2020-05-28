package util

import (
	"algorithm/types"
	"time"
)

const filePath = "../simulation.conf"

type Configuration struct {
	VehiclesPerMinute int32
	IntersectionPolicy string
	SimulationDuration time.Duration
	VehicleMaxSpeed types.MetersPerSecond
	RandomSeed int64
}



func ReadConfiguration() (Configuration, error) {
	// TODO read from file
	d, err := time.ParseDuration("120s")
	if err != nil {
		panic(err)
	}
	return Configuration{
		VehiclesPerMinute: 120,
		IntersectionPolicy: "fcfs",
		SimulationDuration: d,
		VehicleMaxSpeed: 10,
		RandomSeed: 42,
	}, nil
}
