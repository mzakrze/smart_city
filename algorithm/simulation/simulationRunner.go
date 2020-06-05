package simulation

import (
	"algorithm/constants"
	"algorithm/logging"
	"algorithm/types"
	"algorithm/util"
	"algorithm/vehicle"
	"fmt"
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
	resultsLogger logging.IResultsLogger
	collisionDetector *vehicle.CollisionDetector

	nextVehicleId types.VehicleId
	vehicleIdToAppearTs map[types.VehicleId]types.Millisecond
	vehicleIdToLeaveTs map[types.VehicleId]types.Millisecond
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
		resultsLogger logging.IResultsLogger,
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
			vehicleIdToLeaveTs: make(map[types.VehicleId]types.Millisecond),
			vehicleIdToAppearTs: make(map[types.VehicleId]types.Millisecond),
			nextVehicleId: 1,
		}
	}

	return instance
}

type ShortStats struct {
	Throughput int // vehicles per minute
	AverageDelay int // seconds (rounded)
}

func (r *SimulationRunner) RunSimulation() ShortStats {

	ts := types.Millisecond(0)

	isFullSecond := func() bool {
		return ts % 1000 == 0
	}
	isFullSecondMinusStep := func() bool {
		return (ts + constants.SimulationStepInterval) % 1000 == 0
	}
	simulationDurationElapsed := func() bool {
		return ts >= types.Millisecond((constants.WarmupSeconds + r.configuration.SimulationDuration) * 1000)
	}
	simulationFinished := func() bool {
		return simulationDurationElapsed() && isFullSecond()
	}

	r.resultsLogger.SimulationStarted(time.Now())

	for simulationFinished() == false {

		// Premise: vehicle must appear only on full second
		if simulationDurationElapsed() == false && isFullSecond() {
			r.spawnNewVehicles(ts)
		}

		for _, v := range r.allVehiclesProxy.GetAllVehicles() {
			v.Ping(ts)
			r.resultsLogger.VehicleStepReport(v.Id, ts, v.X, v.Y, v.Alpha, v.Speed, v.Acc, v.State)
		}

		r.im.Ping(ts)

		// Premise: vehicle must disappear only on full second - step interval
		if isFullSecondMinusStep() {
			prunedVehicles := r.cleanUpVehicles(ts)
			for _, v := range prunedVehicles {
				r.resultsLogger.VehicleFinished(v.Id, ts)
			}
		}

		ts += constants.SimulationStepInterval

		r.CurrentTs = ts
	}

	r.resultsLogger.SimulationFinished(time.Now())

	return r.getStatitistics()
}



func (r *SimulationRunner) spawnNewVehicles(ts types.Millisecond) []*vehicle.VehicleActor {
	spawned := []*vehicle.VehicleActor{}

	toSpawn, remainder := int(r.configuration.VehiclesPerMinute / 60), r.configuration.VehiclesPerMinute % 60
	if int(remainder) > rand.Intn(60) {
		toSpawn += 1
	}

	for i := 0; i < toSpawn; i++ {
		newVehicle := r.createRandomVehicleIfEntryPointAvailable(ts)
		if newVehicle == nil {
			break
		}

		r.vehicleIdToAppearTs[newVehicle.Id] = ts
		r.allVehiclesProxy.RegisterVehicle(newVehicle)
		spawned = append(spawned, newVehicle)
	}

	return spawned
}


func (r *SimulationRunner) createRandomVehicleIfEntryPointAvailable(ts types.Millisecond) *vehicle.VehicleActor {
	possibleLaneExitpoints := make(map[int][]int)

	legalWays := func(lane int, way types.WayId) []types.WayId {
		// TODO - narazie obsługa tylko 2 pasów
		if lane == 0 {
			way -= 1
			first := (way + 1) % 4 + 1 //left
			second := (way + 2) % 4 + 1 //straight
			return []types.WayId{first, second}
		}
		if lane == 1 {
			way -= 1
			first := (way + 3) % 4 + 1 //left
			second := (way + 2) % 4 + 1 //straight
			return []types.WayId{first, second}
		}
		panic("Oops")
	}

	possibleLaneExitpoints[0] = []int{1}
	possibleLaneExitpoints[1] = []int{}
	possibleLaneExitpoints[2] = []int{1,2}
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
			if e.ExitPointId == entrypoint.EntryPointId && e.WayId != entrypoint.WayId {

				for _, v := range legalWays(entrypoint.EntryPointId, entrypoint.WayId) {
					if v == e.WayId {
						possibleExitpoints = append(possibleExitpoints, e)
					}
				}

			}
		}
		return possibleExitpoints[rand.Intn(len(possibleExitpoints))]
	}

	entrypoint, distance := getRandomAvailableEntrypoint()
	if entrypoint == nil {
		return nil // no available entrypoint - cannot generate new vehicle
	}
	// given a (deceleration) and s (braking distance), what is max v (speed)?
	initSpeed := math.Min(10.0, math.Sqrt(2 * distance * r.configuration.VehicleBrakingForce / r.configuration.VehicleWeight) * 0.5) * 0.8
	exitpoint := getRandomCompatibleExitpoint(entrypoint)

	vId := r.nextVehicleId
	r.nextVehicleId += 1

	return vehicle.NewVehicleActor(vId, ts, entrypoint, exitpoint, initSpeed, r.graph, r.sensorLayer, r.communicationLayer)
}


func (r *SimulationRunner) cleanUpVehicles(ts types.Millisecond) []*vehicle.VehicleActor {
	pruned := []*vehicle.VehicleActor{}

	allVehicles := r.allVehiclesProxy.GetAllVehicles()
	for i := range allVehicles {

		v := allVehicles[i]
		if v.HasFinished {
			pruned = append(pruned, v)
		}
	}
	for _, v := range pruned {
		r.vehicleIdToLeaveTs[v.Id] = ts
		r.allVehiclesProxy.UnregisterVehicle(v)
	}

	return pruned
}

func (r *SimulationRunner) getStatitistics() ShortStats {
	sent, lost, d := r.communicationLayer.GetStats()
	fmt.Println("Sent:", sent, ", lost:", lost, ", avgDelay:", d)

	timeToIgnore := types.Millisecond(constants.WarmupSeconds * 1000)

	vehiclesCounter := 0
	sumDelay := types.Millisecond(0)

	for vId, tsAppear := range r.vehicleIdToAppearTs {
		if tsAppear >= timeToIgnore {
			tsLeft, e := r.vehicleIdToLeaveTs[vId]
			if e {
				sumDelay += tsLeft - tsAppear
				vehiclesCounter += 1
			}
		}
	}

	avgDelay := int(float64(sumDelay) / float64(vehiclesCounter) / 1000)
	if vehiclesCounter == 0 {
		avgDelay = 0
	}
	return ShortStats{
		Throughput: vehiclesCounter * 60 / r.configuration.SimulationDuration,
		AverageDelay: avgDelay,
	}

}
