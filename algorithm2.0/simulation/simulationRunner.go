package simulation

import (
	"algorithm2.0/constants"
	"algorithm2.0/logging"
	"algorithm2.0/types"
	"algorithm2.0/util"
	"algorithm2.0/vehicle"
	"math"
	"math/rand"
	"time"
)

type SimulationRunner struct {
	configuration util.Configuration
	graph *util.Graph
	im *vehicle.IntersectionManager
	allVehiclesProxy *vehicle.AllVehicleProxy
	communicationLayer *vehicle.CommunicationLayer
	sensorLayer *vehicle.SensorLayer
	resultsLogger *logging.ResultsLogger
	collisionDetector *vehicle.CollisionDetector

	nextVehicleId types.VehicleId
	CurrentTs types.Millisecond
}


var instance *SimulationRunner = nil
func SimulationRunnerSingleton(
		configuration util.Configuration,
		graph *util.Graph,
		intersectionManager *vehicle.IntersectionManager,
		allVehiclesProxy *vehicle.AllVehicleProxy,
		communicationLayer *vehicle.CommunicationLayer,
		sensorLayer *vehicle.SensorLayer,
		resultsLogger *logging.ResultsLogger,
		collisionDetector *vehicle.CollisionDetector,
	) *SimulationRunner {

	if instance == nil {
		instance = &SimulationRunner{
			configuration: configuration,
			graph: graph,
			im: intersectionManager,
			allVehiclesProxy: allVehiclesProxy,
			communicationLayer: communicationLayer,
			sensorLayer: sensorLayer,
			resultsLogger: resultsLogger,
			collisionDetector: collisionDetector,
		}
	}

	return instance
}


func (r *SimulationRunner) RunSimulation() {

	ts := types.Millisecond(0)

	isFullSecond := func() bool {
		return ts % 1000 == 0
	}
	isFullSecondMinusStep := func() bool {
		return (ts + constants.SimulationStepInterval) % 1000 == 0
	}
	simulationDurationElapsed := func() bool {
		return ts >= types.Millisecond(r.configuration.SimulationDuration.Seconds() * 1000)
	}
	simulationFinished := func() bool {
		return simulationDurationElapsed() &&
			//r.allVehiclesProxy.AllVehiclesDone() &&
			isFullSecond()
	}

	r.resultsLogger.SimulationStarted(r.configuration.SimulationName, time.Now())

	for simulationFinished() == false {

		// Premise: vehicle must appear only on full second
		if simulationDurationElapsed() == false && isFullSecond() {
			r.spawnNewVehicles(ts)
		}

		for _, v := range r.allVehiclesProxy.GetAllVehicles() {
			v.Ping(ts)
			r.resultsLogger.VehicleStepReport(v.Id, ts, v.X, v.Y, v.Alpha, v.Speed, v.Acc)
		}

		r.im.Ping(ts)

		// Premise: vehicle must disappear only on full second - step interval
		if isFullSecondMinusStep() {
			prunedVehicles := r.cleanUpVehicles()
			for _, v := range prunedVehicles {
				r.resultsLogger.VehicleFinished(v.Id, ts)
			}
		}

		ts += constants.SimulationStepInterval

		r.CurrentTs = ts
	}

	r.resultsLogger.SimulationFinished(time.Now())

}

func (r *SimulationRunner) spawnNewVehicles(ts types.Millisecond) []*vehicle.VehicleActor {
	spawned := []*vehicle.VehicleActor{}

	toSpawn, _ := int(r.configuration.VehiclesPerMinute / 60), r.configuration.VehiclesPerMinute % 60
	//if int(remainder) < rand.Intn(60) { // TODO
	//	toSpawn += 1
	//}

	//if toSpawn > 0 {
	//	toSpawn = 1
	//	r.configuration.VehiclesPerMinute = 0
	//}

	//if len(r.allVehiclesProxy.GetAllVehicles()) == 0 {
	//	toSpawn = 1
	//} else {
	//	toSpawn = 0
	//}

	for i := 0; i < toSpawn; i++ {
		newVehicle := r.createRandomVehicleIfEntryPointAvailable(ts)
		if newVehicle == nil {
			break
		}

		r.allVehiclesProxy.RegisterVehicle(newVehicle)
		spawned = append(spawned, newVehicle)
	}

	return spawned
}


func (r *SimulationRunner) createRandomVehicleIfEntryPointAvailable(ts types.Millisecond) *vehicle.VehicleActor {
	getRandomAvailableEntrypoint := func() (*util.Node, types.Meter) {
		availableEntrypoints := []*util.Node{}
		spareDistance := make(map[types.NodeId]types.Meter)

		for _, entry := range r.graph.Entrypoints {
			edge := entry.EdgesFrom[0] // entryPoint always has exactly 1 edge
			minDist := 100.0
			for _, v := range r.allVehiclesProxy.GetAllVehicles() {
				if edge.Id == v.EdgeAt.Id {
					x := entry.X - v.X
					y := entry.Y - v.Y
					d := math.Sqrt(x * x + y * y)
					minDist = math.Min(minDist, d)
				}
			}

			if minDist > 10 {
				spareDistance[entry.Id] = minDist
				availableEntrypoints = append(availableEntrypoints, entry)
			}
		}

		if len(availableEntrypoints) == 0 {
			return nil, 0
		}

		entrypoint := availableEntrypoints[rand.Intn(len(availableEntrypoints))]
		return entrypoint, spareDistance[entrypoint.Id]
	}

	getRandomCompatibleExitpoint := func(entrypoint *util.Node) *util.Node {
		possibleExitpoints := []*util.Node{}
		for _, e := range r.graph.Exitpoints {
			if e.ExitPointId == entrypoint.EntryPointId && e.WayId != entrypoint.WayId && entrypoint.EntryPointId == 0 { // && math.Abs(float64(e.WayId - entrypoint.WayId)) == 1
				possibleExitpoints = append(possibleExitpoints, e)
			}
		}
		return possibleExitpoints[rand.Intn(len(possibleExitpoints))]
	}

	entrypoint, distance := getRandomAvailableEntrypoint()
	if entrypoint == nil {
		return nil // no available entrypoint - cannot generate new vehicle
	}
	// given a (deceleration) and s (braking distance), what is max v (speed)?
	initSpeed := math.Min(r.configuration.VehicleMaxSpeed, math.Sqrt(2.0 * distance * r.configuration.VehicleMaxDecel)) * 0.8

	exitpoint := getRandomCompatibleExitpoint(entrypoint)

	vId := r.nextVehicleId
	r.nextVehicleId += 1

	return vehicle.NewVehicleActor(vId, ts, entrypoint, exitpoint, initSpeed, r.graph, r.sensorLayer, r.communicationLayer)
}


func (r *SimulationRunner) cleanUpVehicles() []*vehicle.VehicleActor {
	pruned := []*vehicle.VehicleActor{}

	allVehicles := r.allVehiclesProxy.GetAllVehicles()
	for i := range allVehicles {

		v := allVehicles[i]
		if v.HasFinished {
			pruned = append(pruned, v)
		}
	}
	for _, v := range pruned {
		r.allVehiclesProxy.UnregisterVehicle(v)
	}

	return pruned
}
