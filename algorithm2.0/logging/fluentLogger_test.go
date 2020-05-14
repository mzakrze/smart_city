package logging

import (
	"algorithm2.0/types"
	"fmt"
	"testing"
	"time"
)

type resultLoggerMock struct {
	err string
	t *testing.T

	msgCounterInfo    int
	msgCounterMap     int
	msgCounterVehicle int
	msgCounterIntersection int
}

func (m *resultLoggerMock) Post(tag string, message interface{}) error {

	var expectedFields map[string]bool
	switch tag {
	case "map":
		m.msgCounterMap += 1
		expectedFields = map[string]bool{
			"simulation_name": false,
			"vehicle_id": false,
			"second": false,
			"location_array": false,
			"alpha_array": false,
		}
	case "vehicle":
		m.msgCounterVehicle += 1
		expectedFields = map[string]bool{
			"simulation_name": false,
			"vehicle_id": false,
			"start_time": false,
			"finish_time": false,
			"duration": false,
			"speed_array": false,
			"acc_array": false,
		}
	case "info":
		m.msgCounterInfo += 1
		expectedFields = map[string]bool {
			"simulation_name": false,
			"simulation_started_ts": false,
			"simulation_finished_ts": false,
		}
	case "intersection":
		m.msgCounterIntersection += 1
		expectedFields = map[string]bool{
			"simulation_name": false,
			"second": false,
			"arrive_no": false,
			"leave_no": false,
		}
	default:
		m.t.Error("Illegal tag provided: " + tag + ". ")
	}

	for k := range message.(map[string]string) {
		_, exists := expectedFields[k]
		if exists {
			expectedFields[k] = true
		} else {
			m.t.Error("Unexpected field in message (tag = " + tag + "): " + k + ". ")
		}
	}

	for k, v := range expectedFields {
		if v == false {
			m.t.Error("Field not provided (tag = " + tag + "): " + k + ". ")
		}
	}

	return nil
}



func TestFluentLoggerUnit(t *testing.T) {

	const simName = "whatever"
	const mapWidth = 500
	const mapHeight = 500
	startTime := time.Now()
	loggerMock := &resultLoggerMock{t: t}

	// given
	fluentLogger := ResultsLoggerSingleton(loggerMock, mapWidth, mapHeight, 10)

	// when
	fluentLogger.SimulationStarted(simName, startTime)
	x, y, alpha := 0.0, 0.0, 0.0
	speed := 10.0
	durationMs := types.Millisecond(mapWidth / speed * 1000)
	acc := 0.0
	vehicleId := types.VehicleId(0)
	for ts := types.Millisecond(0); ts < durationMs; ts += 10 {
		fluentLogger.VehicleStepReport(vehicleId, ts, x, y, alpha, speed, acc)
		x += speed * 10.0 / 1000.0
	}
	fluentLogger.VehicleFinished(vehicleId, durationMs)
	fluentLogger.SimulationFinished(startTime.Add(time.Second * 2))

	// then
	if expected := 1; loggerMock.msgCounterVehicle != expected {
		// 1 per vehicle
		t.Error(fmt.Printf("Number of messages sent by fluent is: %d. should be %d\n", loggerMock.msgCounterVehicle, expected))
	}
	if expected := 1; loggerMock.msgCounterInfo != expected {
		// 1 per simulation
		t.Error(fmt.Printf("Number of messages sent by fluent is: %d. should be %d\n", loggerMock.msgCounterInfo, expected))
	}
	// FIXME - 49
	if expected := 49; loggerMock.msgCounterMap != expected {
		// 1 per vehicle * seconds
		t.Error(fmt.Printf("Number of messages sent by fluent is: %d. should be %d\n", loggerMock.msgCounterMap, expected))
	}
	if expected := 50; loggerMock.msgCounterIntersection != expected {
		// 1 per second
		t.Error(fmt.Printf("Number of messages sent by fluent is: %d. should be %d\n", loggerMock.msgCounterIntersection, expected))
	}

}

// It' semi-automatic test, also it's more an integration test
func TestFluentElasticKibanaIntegration(t *testing.T) {

	// given


	// when

	// then


	// !!! Manual Test Required !!!
	// (1) go to localhost:3000 and check if vehicles ride
	// (2) go to localhost:5601 and check if Kibana has nice dashboard
}