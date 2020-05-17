package logging

import (
	"algorithm2.0/types"
	"fmt"
	"github.com/fluent/fluent-logger-golang/fluent"
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

const simName = "whatever"
const mapWidth = 500
const mapHeight = 500

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
			"simulation_max_ts": false,
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
	startTime := time.Now()
	loggerMock := &resultLoggerMock{t: t}

	// given
	fluentLogger := ResultsLoggerSingleton(loggerMock, mapWidth, mapHeight)

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
	fluentLogger.VehicleFinished(vehicleId, durationMs - 10)
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
	if expected := 50; loggerMock.msgCounterMap != expected {
		// 1 per vehicle * seconds
		t.Error(fmt.Printf("Number of messages sent by fluent is: %d. should be %d\n", loggerMock.msgCounterMap, expected))
	}
	if expected := 50; loggerMock.msgCounterIntersection != expected {
		// 1 per second
		t.Error(fmt.Printf("Number of messages sent by fluent is: %d. should be %d\n", loggerMock.msgCounterIntersection, expected))
	}

}

// It' semi-automatic test, also it's more an integration test
// required EFK stack running
func TestFluentElasticVisualizationIntegration(t *testing.T) {

	const simName = "whatever"
	const mapWidth = 500
	const mapHeight = 500
	startTime := time.Now()

	// given
	trueFluentLogger , err := fluent.New(fluent.Config{}); if err != nil { t.Fatal(err) }
	fluentLogger := ResultsLoggerSingleton(trueFluentLogger, mapWidth, mapHeight, 10)

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
	fluentLogger.VehicleFinished(vehicleId, durationMs - 10)
	fluentLogger.SimulationFinished(startTime.Add(time.Second * 2))

	// then
	// !!! Manual Test Required !!!
	// (1) check if there are any error in fluent console
	// (2) go to localhost:3000 and check if vehicles ride

}


// It' semi-automatic test, also it's more an integration test
// required EFK stack running
func TestFluentElasticVisualizationKibanaIntegration(t *testing.T) {

	startTime := time.Now()

	// given
	trueFluentLogger , err := fluent.New(fluent.Config{}); if err != nil { t.Fatal(err) }
	fluentLogger := ResultsLoggerSingleton(trueFluentLogger, mapWidth, mapHeight, 10)

	// when
	fluentLogger.SimulationStarted(simName, startTime)
	durationMs := types.Millisecond(120 * 1000)
	vehiclesIds, vehiclesSteps, vehiclesFinishTs := mockMultipleVehicles(durationMs)
	for ts := types.Millisecond(0); ts < durationMs; ts += 10 {
		for _, v := range vehiclesIds {
			if ts  <= vehiclesFinishTs[v]  {
				r := vehiclesSteps[v][ts / 10]
				fluentLogger.VehicleStepReport(v, ts, r.x, r.y, r.alpha, r.speed, r.acc)

				if vehiclesFinishTs[v] == ts {
					fluentLogger.VehicleFinished(v, ts)
				}
			}
		}
	}

	fluentLogger.SimulationFinished(startTime.Add(time.Second * 2))

	// then
	/// !!! Manual test
	// (1) check if there are any error in fluent console
	// (2) go to localhost:3000 and check if vehicles ride
	// (3) go to localhost:5601 and check if Kibana has nice dashboard

}

type vehicleStepReport struct {
	x     types.XCoord
	y     types.YCoord
	alpha types.Angle
	speed types.MetersPerSecond
	acc   types.MetersPerSecond2
}

func mockMultipleVehicles(maxTs types.Millisecond) ([]types.VehicleId, map[types.VehicleId][]vehicleStepReport, map[types.VehicleId]types.Millisecond) {
	// generates multiple vehicles mock - starting at the same ms
	vehicleIds := []types.VehicleId{}
	vehicleStepsMap := make(map[types.VehicleId][]vehicleStepReport)
	vehicleFinishMap := make(map[types.VehicleId]types.Millisecond)

	minSpeed := float64(1000.0 * mapWidth / maxTs)
	maxSpeed := 20.0

	vId := 0

	// 10 vehicles with different speeds
	for s := minSpeed; s < maxSpeed; s += (maxSpeed - minSpeed) / 10 {
		vehicleSteps := []vehicleStepReport{}

		x := 0.0
		y := 0.0
		for t := types.Millisecond(0); t < maxTs; t += 10 {
			r := vehicleStepReport{
				x: x,
				y: y,
				alpha: 0.0,
				speed: s,
				acc: 0.0,
			}

			vehicleSteps = append(vehicleSteps, r)
			x += s * 10.0 / 1000.0

			if x >= mapWidth && t % 1000 == 990 {
				vehicleFinishMap[types.VehicleId(vId)] = t
				vehicleStepsMap[types.VehicleId(vId)] = vehicleSteps
				vehicleIds = append(vehicleIds, types.VehicleId(vId))
				break
			}
		}

		vId += 1
	}

	return vehicleIds, vehicleStepsMap, vehicleFinishMap
}