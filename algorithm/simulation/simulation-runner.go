package simulation

import (
	"fmt"
	"math"
	"math/rand"
	"mzakrze/smart_city/algorithm"
	"mzakrze/smart_city/util"
	"time"
)

const (
	MAX_LAT = 52.230000
	MAX_LON = 21.015000
	MIN_LAT = 52.227000
	MIN_LON = 21.000000

	STEP_INTERVAL_MS = 100
	SIMULATION_START_LAYOUT 	= "2020-01-01T20:00:00.000Z"
	SIMULATION_END_LAYOUT 		= "2020-01-01T21:00:00.000Z"

	// TODO - funkcja rozkladu prawdopodobienstwa:
	// 1 - pojawienia sie pojazdu od czasu (np 7-9 i 16-18 godziny szczytu itp.)
	// 2 - lokalizacji

	VEHICLES_NO = 1000

	// 
	STEPS_IN_SECOND = 1000 / STEP_INTERVAL_MS
)

type SimulationRunner struct {
	simulationStart int64
	simulationEnd int64

	logger *FluentLogger
	vehiclesControllers [VEHICLES_NO]VehicleController
}

func CreateSimulationRunner() *SimulationRunner {
	validateSettings()
	
	instance := SimulationRunner{}

	instance.initStartEnd()

	instance.initVehicleControllers()

	instance.initFluentdLogger()

	return &instance
}

func (r *SimulationRunner) initStartEnd() {
	start, err := time.Parse(time.RFC3339, SIMULATION_START_LAYOUT)
	if err != nil { panic(err) }
	end, err := time.Parse(time.RFC3339, SIMULATION_END_LAYOUT)
	if err != nil { panic(err) }
	var nanoToMs int64 = 10e6

	r.simulationStart = int64(start.UnixNano() / nanoToMs)
	r.simulationEnd = int64(end.UnixNano() / nanoToMs)
}

func (r *SimulationRunner) initVehicleControllers() {
	// r.vehiclesControllers = [VEHICLES_NO]VehicleController

	for vehicleId := 0; vehicleId < VEHICLES_NO; vehicleId += 1 {
		originLat 			:= rand.Float64() * (MAX_LAT - MIN_LAT) + MIN_LAT
		destinationLat 		:= rand.Float64() * (MAX_LAT - MIN_LAT) + MIN_LAT
		originLon 			:= rand.Float64() * (MAX_LON - MIN_LON) + MIN_LON
		destinationLon 		:= rand.Float64() * (MAX_LON - MIN_LON) + MIN_LON
	
		startTs	:= r.simulationStart // TODO losować
	
		vehicleController := VehicleController{
			vehicleId: vehicleId,
			isCurrentlyDriving: false,
			hasFinished: false,
			startTs: startTs,
			originLat: originLat,
			originLon: originLon,
			destinationLat: destinationLat,
			destinationLon: destinationLon,
		}

		// fmt.Printf("origin: %f, %f; dest: %f, %f\n", originLat, originLon, destinationLat, destinationLon)

		r.vehiclesControllers[vehicleId] = vehicleController 
	} 
}

func (r *SimulationRunner) initFluentdLogger() {
	// TODO - docelowo pewnie wielowątkowo
	r.logger = NewFluentdLogger()
}

type VehicleController struct {
	vehicleId int
	vehicleActor *algorithm.VehicleActor

	isCurrentlyDriving bool
	hasFinished bool

	startTs int64
	endTs int64

	originLat float64
	originLon float64
	destinationLat float64
	destinationLon float64

	locationReport VehicleLocationReport
	tripReportSent bool // TODO jakos to przepisac
}

type VehicleLocationReport struct {
	location [10]util.LocationStruct
	step int
	startSecond int64
	isEmpty bool // TODO - przepisac to
} 

func (r *SimulationRunner) reportCurrentLocation(v* VehicleController, ts int64) {
	data := &FluentVehicleCurrentLocation{
		VehicleId: v.vehicleId,
		Timestamp: ts,
		Lat: v.vehicleActor.Lat,
		Lon: v.vehicleActor.Lon,
		Speed: v.vehicleActor.Speed_mps,
	}
	r.logger.VehicleCurrentLocation(data)
}

func (r *SimulationRunner) appendLocationToBucket(v* VehicleController, ts int64) {
	if v.locationReport.isEmpty {
		v.locationReport.startSecond = ts / 1000
		v.locationReport.step = 0
		v.locationReport.isEmpty = false
	}
	
	v.locationReport.location[v.locationReport.step] = util.LocationStruct{
		Lat: v.vehicleActor.Lat,
		Lon: v.vehicleActor.Lon,
	}
	v.locationReport.step += 1
}

func (r *SimulationRunner) flushBucketIfFull(v *VehicleController, ts int64) {
	if v.locationReport.step == cap(v.locationReport.location) {
		var bboxNorth float64 = -math.MaxFloat64
		var bboxSouth float64 = math.MaxFloat64
		var bboxEast float64 = -math.MaxFloat64
		var bboxWest float64 = math.MaxFloat64

		for _, loc := range v.locationReport.location {
			if loc.Lat > bboxNorth {
				bboxNorth = loc.Lat
			}
			if loc.Lat < bboxSouth {
				bboxSouth = loc.Lat
			}
			if loc.Lon > bboxEast {
				bboxEast = loc.Lon
			}
			if loc.Lon < bboxWest {
				bboxWest = loc.Lon
			}
		}

		data := &FluentVehicleBucketLocation{
			VehicleId: v.vehicleId, 
			StartSecond: v.locationReport.startSecond,
			Location: v.locationReport.location[:],
			BboxNorth: bboxNorth,
			BboxSouth: bboxSouth,
			BboxEast: bboxEast,
			BboxWest: bboxWest,
		}

		r.logger.VehicleBucketLocation(data)
		
		v.locationReport.isEmpty = true
		v.locationReport.step = 0
	} 
}

func (v *VehicleController) checkIfVehicleArrived() bool {
	// FIXME
	if math.Abs(v.vehicleActor.Lat - v.vehicleActor.DestinationLat) < 0.00005 &&
		math.Abs(v.vehicleActor.Lon - v.vehicleActor.DestinationLon) < 0.00005 {

		// fmt.Println("Vehicle arrived!")

		return true

	} else {
		return false
	}
}

func (r *SimulationRunner) reportIfReachedDestination(v* VehicleController, ts int64) {

	if v.hasFinished && v.tripReportSent == false {

		data := &FluentVehicleTrip{
			VehicleId: v.vehicleId,
			StartTs: v.startTs,
			EndTs: v.endTs,
			OriginLat: v.originLat,
			OriginLon: v.originLon,
			DestinationLat: v.destinationLat, 
			DestinationLon: v.destinationLon,
		}

		r.logger.VehicleTrip(data)
		v.tripReportSent = true
	}
}

func (r *SimulationRunner) report(ts int64) {

	// FIXME - gdy jest robione tak jak niżej zakomentowane to cos sie psuje z przekazywaniem przez wartosc/wskaznik
	// przez co nie agreguje się do locationReport

	for i := 0; i < VEHICLES_NO; i++ {
		if r.vehiclesControllers[i].isCurrentlyDriving {
			r.reportCurrentLocation(&r.vehiclesControllers[i], ts)
			r.appendLocationToBucket(&r.vehiclesControllers[i], ts)
			r.flushBucketIfFull(&r.vehiclesControllers[i], ts)
			r.reportIfReachedDestination(&r.vehiclesControllers[i], ts)
		}
	}

	// for _, v := range r.vehiclesControllers {
	// 	r.reportCurrentLocation(&v, ts)
	// 	r.appendLocationToBucket(&v, ts)
	// 	r.flushBucketIfFull(&v, ts)
	// 	r.reportIfReachedDestination(&v, ts)
	// }

}

func (r *SimulationRunner) flushAllReports(ts int64) {

	for _, v := range r.vehiclesControllers {
		if v.locationReport.isEmpty == false {
			data := &FluentVehicleBucketLocation{
				VehicleId: v.vehicleId, 
				StartSecond: v.locationReport.startSecond,
				Location: v.locationReport.location[:],
				BboxNorth: 1000,
				BboxSouth: -1000,
				BboxEast: -1000,
				BboxWest: 1000,
				// FIXME - te boundy wyliczyć
			}
	
			r.logger.VehicleBucketLocation(data)
		}
	}

}

func (v* VehicleController) ping(ts int64) {
	if v.hasFinished {
		return
	}

	if v.startTs >= ts && v.isCurrentlyDriving == false {
		// initiate vehicleActor
		v.vehicleActor = &algorithm.VehicleActor{
			Lat: v.originLat,
			Lon: v.originLon,
			DestinationLat: v.destinationLat,
			DestinationLon: v.destinationLon,
		}

		v.isCurrentlyDriving = true
		v.locationReport = VehicleLocationReport{
			isEmpty: true,
		}
	}

	if v.isCurrentlyDriving {
		v.vehicleActor.Ping(ts)
		
		// fmt.Printf("lat: %f, destLat: %f, diff: %f \n", v.vehicleActor.Lat, v.destinationLat, math.Abs(v.destinationLat - v.vehicleActor.Lat))
		// fmt.Printf("lon: %f, destLon: %f, diff: %f \n", v.vehicleActor.Lon, v.destinationLon, math.Abs(v.destinationLon - v.vehicleActor.Lon))
	}

	if v.checkIfVehicleArrived() {
		v.isCurrentlyDriving = false
		v.endTs = ts
		v.hasFinished = true
	}

}

func validateSettings() {
	
	if 1000 % STEP_INTERVAL_MS != 0 {
		panic(fmt.Sprintf("variable STEP_INTERVAL_MS has to divisible by 1000, but is: %d", STEP_INTERVAL_MS))
	}

	// TODO - czy nie za duzo 

	// TODO - czy min/max lat/lon ma sens
}

func (r *SimulationRunner) haveAllVehiclesFinished() bool {
	for _, v := range r.vehiclesControllers {
		if v.hasFinished == false {
			return false
		}
	}
	return true
}

func (r *SimulationRunner) RunSimulation() {

	ts := r.simulationStart
	for ; ts <= r.simulationEnd || ! r.haveAllVehiclesFinished(); ts += STEP_INTERVAL_MS {
		
		for i := 0; i < len(r.vehiclesControllers); i++ {
			r.vehiclesControllers[i].ping(ts)
		}

		r.report(ts)

	}

	// r.flushAllReports(ts)

	fmt.Println("Done.")
}

