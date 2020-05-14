package logging

import (
	"algorithm2.0/types"
	"github.com/fluent/fluent-logger-golang/fluent"
	"time"
)

func ResultsLoggerSingleton(logger *fluent.Fluent, mapWidth, mapHeight types.Meter) *ResultsLogger {
	if instance == nil {
		instance = &ResultsLogger{}
	}
	return instance
}
func (f *ResultsLogger) SimulationStarted(simName string, startTime time.Time) {

}

func (f *ResultsLogger) SimulationEnded(simName string, finishTime time.Time) {

}

func (f*ResultsLogger) VehicleLocation(id types.VehicleId, ts types.Millisecond, x types.XCoord, y types.YCoord, alpha types.Angle) {

}

func (f*ResultsLogger) VehicleFinished(id types.VehicleId, ts types.Millisecond) {

}




var instance *ResultsLogger = nil
type ResultsLogger struct {

}



