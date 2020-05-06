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

	VEHICLES_NO = 50

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

	entrypoints := []types.DestinationPoint{}
	exitpoints := []types.DestinationPoint{}
	for _, n := range r.RoadsGraph.AllNodes {
		if n.IsEntrypoint != 0 {
			// z kazdego entrypointu zawsze dokladnie 1 edge
			nodeTo := n.Edges[0].To
			d := types.DestinationPoint{
				NodeFirst: n.Id,
				NodeSecond: nodeTo.Id,
				X: n.X,
				Y: n.Y,
				Cluster: n.IsEntrypoint,
			}

			entrypoints = append(entrypoints, d)
		} else {
			for _, e := range n.Edges {
				if e.To.IsExitpoint != 0 {
					nodeTo := e.To
					d := types.DestinationPoint{
						NodeFirst: n.Id,
						NodeSecond: nodeTo.Id,
						X: nodeTo.X,
						Y: nodeTo.Y,
						Cluster: e.To.IsExitpoint,
					}

					exitpoints = append(exitpoints, d)
				}
			}
		}
	}

	getRandomWithDifferentCluster := func (points []types.DestinationPoint, cluster types.RoadCluster) types.DestinationPoint {
		maxTries := 100
		for i := 0; i < maxTries; i++ {
			r := points[rand.Intn(len(points))]
			if r.Cluster != cluster {
				return  r
			}
		}
		panic("Fail after 100 tries")
	}

	for vehicleId := 0; vehicleId < VEHICLES_NO; vehicleId += 1 {
		origin := entrypoints[rand.Intn(len(entrypoints))]
		destination := getRandomWithDifferentCluster(exitpoints, origin.Cluster)

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

