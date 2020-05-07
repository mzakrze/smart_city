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
	SIMULATION_END_LAYOUT 		= "2020-01-01T20:01:00.000Z"

	//
	STEPS_IN_SECOND = 1000 / STEP_INTERVAL_MS
)

type SimulationRunner struct {
	simulationStart types.Milisecond
	simulationEnd   types.Milisecond
	nextVehicleId 	types.VehicleId

	logger              *FluentLogger
	vehicleAheadSensor  *algorithm.VehicleAheadSensor
	RoadsGraph          *types.Graph
	VehiclesControllers []*VehicleController
	entrypoints         []types.DestinationPoint // TODO - różne typy, bo sie kiedyś pomiesza in z out
	exitpoints          []types.DestinationPoint
}

func CreateSimulationRunner() *SimulationRunner {
	validateSettings()

	roadsGraph := util.ReadGraph()
	entrypoints, exitpoints := parseRoadsGraph(roadsGraph)

	instance := SimulationRunner{RoadsGraph: roadsGraph, entrypoints: entrypoints, exitpoints: exitpoints}
	instance.vehicleAheadSensor = &algorithm.VehicleAheadSensor{AllVehicles: []*algorithm.VehicleActor{}}

	instance.initStartEnd()


	instance.initFluentdLogger()

	return &instance
}

func parseRoadsGraph(graph *types.Graph) ([]types.DestinationPoint, []types.DestinationPoint) {
	entrypoints := []types.DestinationPoint{}
	exitpoints := []types.DestinationPoint{}
	for _, n := range graph.AllNodes {
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

	return entrypoints, exitpoints
}

func (r *SimulationRunner) initStartEnd() {
	start, err := time.Parse(time.RFC3339, SIMULATION_START_LAYOUT)
	if err != nil { panic(err) }
	end, err := time.Parse(time.RFC3339, SIMULATION_END_LAYOUT)
	if err != nil { panic(err) }
	const nanoToMs int64 = 10e6
	const hackFIXME int = 10 // TODO - pobieram nanosekundy, zamieniam na ms, a bazie jest 10 razy za mała wartość. Pewnie na innej warstwie sie psuje, ale taki hack na szybko

	r.simulationStart = int(start.UnixNano() / nanoToMs) * hackFIXME
	r.simulationEnd = int(end.UnixNano() / nanoToMs) * hackFIXME
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
	VehicleId    types.VehicleId
	VehicleActor *algorithm.VehicleActor

	VehicleState types.VehicleState

	startTs types.Milisecond
	endTs   types.Milisecond

	origin           types.DestinationPoint
	destination      types.DestinationPoint
	RoadsGraph       *types.Graph
	simulationRunner *SimulationRunner
}


func (v* VehicleController) ping(ts types.Milisecond) {
	switch v.VehicleState {
	case types.VEHICLE_FINISHED:
		// do nothing
	case types.VEHICLE_DRIVING:
		v.VehicleActor.Ping(ts)
		if v.VehicleActor.HasFinished {
			v.VehicleState = types.VEHICLE_FINISHED
		}
	case types.VEHICLE_NOT_STARTED:
		// TODO
	}
}

func validateSettings() {
	
	if 1000 % STEP_INTERVAL_MS != 0 {
		panic(fmt.Sprintf("variable STEP_INTERVAL_MS has to divisible by 1000, but is: %d", STEP_INTERVAL_MS))
	}

	// TODO - czy nie za duzo 

	// TODO - czy min/max lat/lon ma sens
}


func (r *SimulationRunner) RunSimulation() {

	//rand.Seed(time.Now().Unix())

	rand.Seed(1237)
	// TODO - coś do śledzenia progresu, np: https://github.com/machinebox/progress

	const simulationTimeMs types.Milisecond = 60 * 1000
	for ts := 0; ts <= simulationTimeMs; ts += STEP_INTERVAL_MS {

		if ts % 3000 == 0 {
			r.updateVehicleControllers(ts)
		}

		for i := 0; i < len(r.VehiclesControllers); i++ {
			r.VehiclesControllers[i].ping(ts)
			r.logger.ReportVehicle(ts, r.VehiclesControllers[i])
		}

	}

	fmt.Println("Done.")
}

func (r *SimulationRunner) updateVehicleControllers(ts types.Milisecond) {

	const VEHICLES_NO = 8

	// 1. usun tych którzy dojechali (i wyślij do ES)
	// FIXME - not implemented

	// 2. dodaj jakiś jeśli jest miejsce
	//for v := len(r.VehiclesControllers); v < VEHICLES_NO; v++ {
		// TODO - narazie zakładam, że to nie bedzie problem, ale sprawdzac czy pojazd ma szanse się zmieścic na lini (wyhamować itp.)
		newVehicle := r.generateNewVehicleController(ts)
		// FIXME - wykrywać czy wygenerowany pojazd ma w ogóle szanse wyhamować (jeśli nie - generować gdzie indziej)
		r.VehiclesControllers = append(r.VehiclesControllers, newVehicle)

		// TODO - to jest troche krzywe, bo trzeba o tym pamiętać
		r.vehicleAheadSensor.AllVehicles = append(r.vehicleAheadSensor.AllVehicles, newVehicle.VehicleActor)
	//}

}
func (r *SimulationRunner) generateNewVehicleController(ts types.Milisecond) *VehicleController {
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

	origin := r.entrypoints[rand.Intn(len(r.entrypoints))]
	destination := getRandomWithDifferentCluster(r.exitpoints, origin.Cluster)

	id := r.nextVehicleId
	r.nextVehicleId += 1

	actor := algorithm.InitVehicleActor(id, origin, destination, ts, r.RoadsGraph, r.vehicleAheadSensor)
	vehicleController := VehicleController{
		VehicleId:        id,
		VehicleState:     types.VEHICLE_DRIVING,
		startTs:          ts,
		origin:           origin,
		destination:      destination,
		RoadsGraph:       r.RoadsGraph,
		simulationRunner: r,
		VehicleActor:	  actor,
	}

	return &vehicleController
}