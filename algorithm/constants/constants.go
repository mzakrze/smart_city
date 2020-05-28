package constants

import "algorithm/types"

// TODO - stała SimulationStepInterval w sekundach - bo bardzo dużo jest castowania na floaat + dzielenia przez 10000
const SimulationStepInterval = types.Millisecond(10)
const SimulationStepIntervalSeconds = float64(SimulationStepInterval) / 1000.0
const VehicleLength = types.Meter(4.636)
const VehicleWidth = types.Meter(1.981)

const SimulationName = "simulation"

const VehiclePower = 120.0 * 745.699872 // [ horse power -> watt ]
const VehicleBrakingForce = 3000.0 // N
const Vehicleweight = 1200.0 // kilogram
const VehicleMaxAngularSpeed = 0.4
const VehicleMaxSpeedOnConflictZone = 15.0