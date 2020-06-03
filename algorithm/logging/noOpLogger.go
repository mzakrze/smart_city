package logging

import (
	"algorithm/types"
	"time"
)

type NoOpLogger struct {

}

func NewNoOpLogger() IResultsLogger {
	n := NoOpLogger{}
	return n
}

func (n NoOpLogger) SimulationStarted(startTime time.Time) {
}

func (n NoOpLogger) SimulationFinished(finishTime time.Time) {
}

func (n NoOpLogger) VehicleStepReport(types.VehicleId, types.Millisecond, types.XCoord, types.YCoord, types.Angle, types.MetersPerSecond, types.MetersPerSecond2, int) {
}

func (n NoOpLogger) VehicleFinished(types.VehicleId, types.Millisecond) {
}