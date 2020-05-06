package simulation

import (
	"fmt"
	"math/rand"
	"mzakrze/smart_city/algorithm"
	"mzakrze/smart_city/types"
	"mzakrze/smart_city/util"
	"time"
)

const (

	STEP_INTERVAL_MS = 100
	SIMULATION_START_LAYOUT 	= "2020-01-01T20:00:00.000Z"
	SIMULATION_END_LAYOUT 		= "2020-01-01T21:00:00.000Z"

	// TODO - funkcja rozkladu prawdopodobienstwa:
	// 1 - pojawienia sie pojazdu od czasu (np 7-9 i 16-18 godziny szczytu itp.)
	// 2 - lokalizacji

	VEHICLES_NO = 20

	// 
	STEPS_IN_SECOND = 1000 / STEP_INTERVAL_MS
)

type SimulationRunner struct {
	simulationStart int64
	simulationEnd int64

	logger *FluentLogger
	RoadsGraph *types.Graph
	vehiclesControllers [VEHICLES_NO]VehicleController
}

func CreateSimulationRunner() *SimulationRunner {
	validateSettings()

	roadsGraph := util.ReadGraph()

	instance := SimulationRunner{RoadsGraph: roadsGraph}

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
	const nanoToMs int64 = 10e6
	const hackFIXME int64 = 10 // TODO - pobieram nanosekundy, zamieniam na ms, a bazie jest 10 razy za mała wartość. Pewnie na innej warstwie sie psuje, ale taki hack na szybko

	r.simulationStart = int64(start.UnixNano() / nanoToMs) * hackFIXME
	r.simulationEnd = int64(end.UnixNano() / nanoToMs) * hackFIXME
}

func (r *SimulationRunner) initVehicleControllers() {

	rand.Seed(time.Now().Unix())

	destinations := []types.DestinationPoint{}
	for i := range r.RoadsGraph.AllNodes {
		nodeFrom := r.RoadsGraph.AllNodes[i]
		for _, e := range nodeFrom.Edges {
			if e.IsInternal {
				x := (nodeFrom.X + e.To.X) / 2
				y := (nodeFrom.Y + e.To.Y) / 2

				d := types.DestinationPoint{
					NodeFirst: nodeFrom.Id,
					NodeSecond: e.To.Id,
					X: x,
					Y: y,
				}

				destinations = append(destinations,  d)
			}
		}
	}

	randomStartStopLocation := func() (types.DestinationPoint, types.DestinationPoint) {
		const maxTries = 100
		guard := 0
		for true {
			i1 := rand.Intn(len(destinations))
			i2 := rand.Intn(len(destinations))
			if i1 != i2 {
				return destinations[i1], destinations[i2]
			}
			guard++
			if guard >= maxTries {
				break
			}
			// try again
		}
		panic(fmt.Sprintf("Error, could find random destination after %d tries.", maxTries))
	}

	for vehicleId := 0; vehicleId < VEHICLES_NO; vehicleId += 1 {
		origin, destination := randomStartStopLocation()

		startTs	:= r.simulationStart // TODO losować

		vehicleController := VehicleController{
			vehicleId: int32(vehicleId),
			VehicleState: types.VEHICLE_NOT_STARTED,
			startTs: startTs,
			origin: origin,
			destination: destination,
			RoadsGraph: r.RoadsGraph,
		}

		r.vehiclesControllers[vehicleId] = vehicleController
	} 
}

func (r *SimulationRunner) initFluentdLogger() {
	// TODO - docelowo pewnie wielowątkowo

	// TODO - dziala tylko dla 1szej cwiartki
	n := r.RoadsGraph.MapBBox.North
	s := r.RoadsGraph.MapBBox.South
	w := r.RoadsGraph.MapBBox.West
	e := r.RoadsGraph.MapBBox.East
	width := r.RoadsGraph.MapBBox.Width
	height := r.RoadsGraph.MapBBox.Height
	latToY := func(lat types.Latitude) types.YCoord {
		return lat / (n - s) * height
	}
	lonToX := func(lon types.Longitude) types.XCoord {
		return lon / (w - e) * width
	}
	yToLat := func(y types.YCoord) types.Latitude {
		return y / height* (n - s) + s
	}
	xToLon := func(x types.XCoord) types.Longitude {
		return x / width * (e - w) + w
	}

	r.logger = NewFluentdLogger(lonToX, latToY, yToLat, xToLon)
}

type VehicleController struct {
	vehicleId types.VehicleId
	vehicleActor *algorithm.VehicleActor

	VehicleState types.VehicleState

	startTs int64
	endTs int64

	origin types.DestinationPoint
	destination types.DestinationPoint
	RoadsGraph *types.Graph
}


func (v* VehicleController) ping(ts types.Timestamp) {
	switch v.VehicleState {
	case types.VEHICLE_FINISHED:
		// do nothing
	case types.VEHICLE_DRIVING:
		v.vehicleActor.Ping(ts)
		if v.vehicleActor.HasFinished {
			v.VehicleState = types.VEHICLE_FINISHED
		}
	case types.VEHICLE_NOT_STARTED:
		if v.startTs >= ts {
			// initiate vehicleActor
			v.vehicleActor = algorithm.InitVehicleActor(v.origin, v.destination, v.RoadsGraph)
			v.VehicleState = types.VEHICLE_DRIVING
		}
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
		if v.VehicleState != types.VEHICLE_FINISHED {
			return false
		}
	}
	return true
}

func (r *SimulationRunner) RunSimulation() {

	// TODO - coś do śledzenia progresu, np: https://github.com/machinebox/progress
	ts := r.simulationStart
	for ; ts <= r.simulationEnd || ! r.haveAllVehiclesFinished(); ts += STEP_INTERVAL_MS {

		for i := 0; i < len(r.vehiclesControllers); i++ {
			r.vehiclesControllers[i].ping(ts)
			r.logger.ReportVehicle(ts, &r.vehiclesControllers[i])
		}

	}

	fmt.Println("Done.")
}

