package constants

import "algorithm/types"

// TODO - stała SimulationStepInterval w sekundach - bo bardzo dużo jest castowania na floaat + dzielenia przez 10000
const SimulationStepInterval = types.Millisecond(10)
const SimulationStepIntervalSeconds = float64(SimulationStepInterval) / 1000.0
const VehicleLength = types.Meter(4.636)
const VehicleWidth = types.Meter(1.981)



