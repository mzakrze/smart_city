package logging

import (
	"algorithm2.0/types"
	"encoding/json"
	"fmt"
	"math"
	"time"
)

const mapTag = "map"
const infoTag = "info"
const vehicleTag = "vehicle"
const vehicleStepTag = "vehiclestep"
const intersectionTag = "intersection"

/*
Whole simulation operates on meters, but visualization operates on latitude and longitude.
Since logger here is an intermediate layer, translation x->lon and y-> lat must happen here.
Let's take an arbitrary coordinates to work with (maxLat and maxLon arise from mapWidth and mapHeight)
*/
const minLat = 52.219111
const minLon = 21.011711

type IResultLogger interface {
	Post(tag string, msg interface{}) error
}

func ResultsLoggerSingleton(logger IResultLogger, mapWidth, mapHeight types.Meter, simulationDurationSeconds float64) *ResultsLogger {
	const simulationStep = types.Millisecond(10)

	// Let's assume that simulation area is small enough to calculate as if Earth is flat
	maxLat, maxLon := calculateMaxLatMaxLon(mapWidth, mapHeight)
	yToLat := func(y types.YCoord) types.Latitude {
		return y / mapHeight * (maxLat - minLat) + minLat
	}
	xToLon := func(x types.XCoord) types.Longitude {
		return x / mapWidth * (maxLon - minLon) + minLon
	}

	if instance == nil {
		instance = &ResultsLogger{
			logger: logger,
			simulationStepInterval: simulationStep,
			yToLat: yToLat,
			xToLon: xToLon,
			bucketSize: int(1000 / simulationStep),
			currentSecond: 0,
			mapLogLocation: make(map[types.VehicleId][]types.LocationStruct),
			mapLogAlpha: make(map[types.VehicleId][]types.Angle),
			mapLogState: make(map[types.VehicleId][]int),
			vehicleLogSpeed: make(map[types.VehicleId][]types.MetersPerSecond),
			vehicleLogAcc: make(map[types.VehicleId][]types.MetersPerSecond2),
			vehicleLogArrive: make(map[types.VehicleId]types.Millisecond),
			intersectionLogVehiclesArrive: make(map[types.Second]int),
			intersectionLogVehiclesLeave: make(map[types.Second]int),
			simulationDurationSeconds: simulationDurationSeconds,
		}
	}
	return instance
}

func (f *ResultsLogger) SimulationStarted(simName string, startTime time.Time) {
	f.simulationName = simName
	f.simulationStartTime = startTime
}
func (f *ResultsLogger) SimulationFinished(finishTime time.Time) {
	f.simulationFinishTime = finishTime
	f.sendInfoLog()
	f.sendIntersectionLog()
}

func (f *ResultsLogger) VehicleStepReport(id types.VehicleId, ts types.Millisecond, x types.XCoord, y types.YCoord, alpha types.Angle, speed types.MetersPerSecond, acc types.MetersPerSecond2, state int) {
	f.maxTs = int(math.Max(float64(ts), float64(f.maxTs)))

	if ts % 100 == 0 {
		f.sendVehicleStepReport(id, ts, x, y, speed, acc)
	}

	var stepInSecond = int((ts % 1000) / f.simulationStepInterval)

	if ts % 1000 == 0 {
		f.currentSecond = types.Second(ts / 1000)
		f.intersectionLogVehiclesArrive[f.currentSecond] = 0
		f.intersectionLogVehiclesLeave[f.currentSecond] = 0
	}

	f.appendToMapLog(stepInSecond, id, x, y, alpha, state)
	f.appendToVehicleLog(stepInSecond, id, speed, acc)

	if ts % 1000 == 990 {
		f.sendMapLogAndFlush(id)
	}
}

func (f *ResultsLogger) VehicleFinished(id types.VehicleId, ts types.Millisecond) {
	if ts % 1000 != 990 {
		panic("Premise broken :( Vehicle should finish only on full second minus step interval")
	}
	f.intersectionLogVehiclesLeave[f.currentSecond] += 1
	f.sendVehicleLogAndFlush(id, ts)
	if ts < 1000 * types.Millisecond(f.simulationDurationSeconds) {
		f.vehiclesFinishedThroughput += 1
	}
}


var instance *ResultsLogger = nil
type ResultsLogger struct {
	logger IResultLogger

	simulationName       string
	simulationStartTime  time.Time
	simulationFinishTime time.Time
	maxTs                int

	simulationStepInterval types.Millisecond
	bucketSize             int

	currentSecond types.Second

	yToLat func(y types.YCoord) types.Latitude
	xToLon func(x types.XCoord) types.Longitude

	mapLogLocation map[types.VehicleId][]types.LocationStruct
	mapLogAlpha    map[types.VehicleId][]types.Angle
	mapLogState    map[types.VehicleId][]int

	vehicleLogSpeed  map[types.VehicleId][]types.MetersPerSecond
	vehicleLogAcc    map[types.VehicleId][]types.MetersPerSecond2
	vehicleLogArrive map[types.VehicleId]types.Millisecond

	intersectionLogVehiclesArrive map[types.Second]int
	intersectionLogVehiclesLeave  map[types.Second]int
	simulationDurationSeconds            float64
	vehiclesFinishedThroughput			int
}














func calculateMaxLatMaxLon(x types.Meter, y types.Meter) (float64, float64) {
	// https://stackoverflow.com/questions/7477003/calculating-new-longitude-latitude-from-old-n-meters
	const earth = 6378.137 //radius of the earth in kilometer
	const m = (1.0 / ((2.0 * math.Pi / 360.0) * earth)) / 1000

	maxLat := minLat + (y * m)
	maxLon := minLon + (x * m) / math.Cos(minLat * (math.Pi / 180.0))

	return maxLat, maxLon
}



func (f *ResultsLogger) appendToMapLog(step int, id types.VehicleId, x types.XCoord, y types.YCoord, alpha types.Angle, state int) {
	if _, exists := f.mapLogLocation[id]; exists == false {
		if step != 0 { panic("Illegal bucket state") }
		f.mapLogLocation[id] = make([]types.LocationStruct, f.bucketSize)
		f.mapLogAlpha[id] = make([]types.Angle, f.bucketSize)
		f.mapLogState[id] = make([]int, f.bucketSize)
	}
	f.mapLogLocation[id][step] = types.LocationStruct{Lon: f.xToLon(x), Lat: f.yToLat(y)}
	f.mapLogAlpha[id][step] = alpha
	f.mapLogState[id][step] = state
}
func (f *ResultsLogger) appendToVehicleLog(step int, id types.VehicleId, speed types.MetersPerSecond, acc types.MetersPerSecond2) {
	if _, exists := f.vehicleLogSpeed[id]; exists == false {
		if step != 0 { panic("Illegal bucket state") }
		f.vehicleLogSpeed[id] = make([]types.MetersPerSecond, f.bucketSize)
		f.vehicleLogAcc[id] = make([]types.MetersPerSecond2, f.bucketSize)
	}
	f.vehicleLogSpeed[id][step] = speed
	f.vehicleLogAcc[id][step] = acc
}

func (f *ResultsLogger) sendInfoLog() {
	throughput := int(float64(f.vehiclesFinishedThroughput) * 60 / f.simulationDurationSeconds) // per minute - hence "* 60"
	msg := map[string]string {
		"simulation_name": f.simulationName,
		"simulation_started_ts": fmt.Sprintf("%d", f.simulationStartTime.Second()),
		"simulation_finished_ts": fmt.Sprintf("%d", f.simulationFinishTime.Second()),
		"simulation_max_ts": fmt.Sprintf("%d", f.maxTs),
		"throughput": fmt.Sprintf("%d", throughput),
	}

	err := f.logger.Post(infoTag, msg); if err != nil { panic(err) }
}

func (f *ResultsLogger) sendMapLogAndFlush(vId types.VehicleId) {
	alpha := f.mapLogAlpha[vId]
	location := f.mapLogLocation[vId]
	state := f.mapLogState[vId]

	delete(f.mapLogAlpha, vId)
	delete(f.mapLogLocation, vId)
	delete(f.mapLogState, vId)

	bytes, err := json.Marshal(location); if err != nil { panic(err) }
	locationJson := string(bytes)
	bytesAlpha, err := json.Marshal(alpha); if err != nil { panic(err) }
	alphaJson := string(bytesAlpha)
	bytesState, err := json.Marshal(state); if err != nil { panic(err) }
	stateJson := string(bytesState)

	msg := map[string]string {
		"simulation_name": f.simulationName,
		"vehicle_id": fmt.Sprintf("%d", vId),
		"second": fmt.Sprintf("%d", f.currentSecond),
		"location_array": locationJson,
		"alpha_array": alphaJson,
		"state_array": stateJson,
	}

	err = f.logger.Post(mapTag, msg)
	if err != nil {
		panic(err)
	}
}

func (f *ResultsLogger) sendIntersectionLog() {
	for sec := types.Second(0); sec <= f.currentSecond; sec++ {
		msg := map[string]string {
			"simulation_name": f.simulationName,
			"second": fmt.Sprintf("%d", sec),
			"arrive_no": fmt.Sprintf("%d", f.intersectionLogVehiclesArrive[sec]),
			"leave_no": fmt.Sprintf("%d", f.intersectionLogVehiclesLeave[sec]),
		}

		err := f.logger.Post(intersectionTag, msg);
		if err != nil {
			panic(err)
		}
	}
}

func (f *ResultsLogger) sendVehicleLogAndFlush(id types.VehicleId, leaveTs types.Millisecond) {
	acc := f.vehicleLogAcc[id]
	speed := f.vehicleLogSpeed[id]
	arriveTs := f.vehicleLogArrive[id]
	duration := int((leaveTs - arriveTs) / 1000)

	bytesAcc, err := json.Marshal(acc); if err != nil { panic(err) }
	accJson := string(bytesAcc)

	bytesSpeed, err := json.Marshal(speed); if err != nil { panic(err) }
	speedJson := string(bytesSpeed)

	msg := map[string]string{
		"simulation_name": f.simulationName,
		"vehicle_id": fmt.Sprintf("%d", id),
		"start_time": fmt.Sprintf("%d", arriveTs),
		"finish_time": fmt.Sprintf("%d", leaveTs),
		"duration": fmt.Sprintf("%d", duration),
		"speed_array": speedJson,
		"acc_array": accJson,
	}

	err = f.logger.Post(vehicleTag, msg); if err != nil { panic(err) }

	delete(f.vehicleLogArrive, id)
	delete(f.vehicleLogSpeed, id)
	delete(f.vehicleLogArrive, id)
}

func (f *ResultsLogger) sendVehicleStepReport(id types.VehicleId, ts types.Millisecond, xCoord types.XCoord, yCoord types.YCoord, speed types.MetersPerSecond, acc types.MetersPerSecond2) {
	// FIXME - nie jest otestowane
	msg := map[string]string {
		"simulation_name": f.simulationName,
		"vehicle_id": fmt.Sprintf("%d", id),
		"speed": fmt.Sprintf("%f", speed),
		"acc": fmt.Sprintf("%f", acc),
		"ts": fmt.Sprintf("%d", ts),
	}

	err := f.logger.Post(vehicleStepTag, msg); if err != nil { panic(err) }

}

