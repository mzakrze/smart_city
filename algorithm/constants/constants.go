package constants

import "algorithm/types"

const SimulationStepInterval = types.Millisecond(10)
const SimulationStepIntervalSeconds = float64(SimulationStepInterval) / 1000.0
const VehicleLength = types.Meter(4.636)
const VehicleWidth = types.Meter(1.981)

const SimulationName = "simulation"

const WarmupSeconds = 30
