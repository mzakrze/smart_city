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
	configuration      util.Configuration
	graph              *util.Graph
	im                 *vehicle.IntersectionManager
	allVehiclesProxy   *vehicle.AllVehicleProxy
	communicationLayer *vehicle.CommunicationLayer
	sensorLayer        *vehicle.SensorLayer
	resultsLogger      logging.IResultsLogger
	collisionDetector  *vehicle.CollisionDetector

	nextVehicleId       types.VehicleId
	vehicleIdToAppearTs map[types.VehicleId]types.Millisecond
	vehicleIdToLeaveTs  map[types.VehicleId]types.Millisecond
	CurrentTs           types.Millisecond
	vehiclesSpawnedInAMinute int
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
	IntersectionThroughput float64 // vehicles per minute
	VehicleAverageDelay    float64 // seconds
	MessagesDelivered      int
	MessagesLost           int
	MessagesAvgDelay       types.Millisecond
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
		return ts >= types.Millisecond((r.configuration.SimulationWarmUp + r.configuration.SimulationDuration) * 1000)
	}
	simulationFinished := func() bool {
		return simulationDurationElapsed() && isFullSecond()
	}

	r.resultsLogger.SimulationStarted(time.Now())

	for simulationFinished() == false {

		// Premise: vehicle must appear only on full second
		if simulationDurationElapsed() == false && isFullSecond() {
			r.spawnNewVehicles(ts)
			r.activeNewVehicles(ts) // TODO - to przeniesc żeby sie wykonywalo zawsze (trzeba zmienić w logowaniu wyników)
		}

		for _, v := range r.allVehiclesProxy.GetAllActiveVehicles() {
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

func (r *SimulationRunner) activeNewVehicles(ts  types.Millisecond) {
	for _, wId := range []types.WayId{1,2,3,4} {
		next := r.allVehiclesProxy.NextQueued(wId)
		for next != nil && r.spaceAvailable(next) {
			distance := 100.0
			for _, v := range r.allVehiclesProxy.GetAllActiveVehicles() {
				if next.EdgeAt.Id == v.EdgeAt.Id {
					x := next.EntryPoint.X - v.X
					y := next.EntryPoint.Y - v.Y
					d := math.Sqrt(x * x + y * y)
					distance = math.Min(distance, d)
				}
			}
			initSpeed := math.Min(10.0, math.Sqrt(2 * distance * r.configuration.VehicleBrakingForce / r.configuration.VehicleWeight) * 0.5) * 0.8
			next.Speed = initSpeed
			r.allVehiclesProxy.RegisterVehicle(next)
			next = r.allVehiclesProxy.NextQueued(wId)
		}
	}
}

func (r *SimulationRunner) spawnNewVehicles(ts types.Millisecond) {
	toSpawn, remainder := int(r.configuration.VehiclesPerMinute / 60), r.configuration.VehiclesPerMinute % 60
	if r.vehiclesSpawnedInAMinute < int(ts % 60000) * remainder {
		toSpawn += 1
		r.vehiclesSpawnedInAMinute += 1
	}

	for i := 0; i < toSpawn; i++ {

		newVehicle := r.createRandomVehicle(ts)

		r.allVehiclesProxy.Enqueue(newVehicle)
		r.vehicleIdToAppearTs[newVehicle.Id] = ts
	}
}

func (r *SimulationRunner) createRandomVehicle(ts types.Millisecond) *vehicle.VehicleActor {
	legalWays := func(lane int, way types.WayId) []types.WayId {
		way -= 1
		left := (way + 1) % 4 + 1
		straight := (way + 2) % 4 + 1
		right := (way + 3) % 4 + 1
		way += 1
		switch lane {
		case 0:
			return []types.WayId{left, straight}
		case r.configuration.MapLanes - 1:
			return []types.WayId{right, straight}
		default:
			return []types.WayId{straight}
		}
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
		if len(possibleExitpoints) == 0 {
			fmt.Println("not ok")
		}
		return possibleExitpoints[rand.Intn(len(possibleExitpoints))]
	}

	entrypoint := r.graph.Entrypoints[rand.Intn(len(r.graph.Entrypoints))]
	exitpoint := getRandomCompatibleExitpoint(entrypoint)

	vId := r.nextVehicleId
	r.nextVehicleId += 1

	return vehicle.NewVehicleActor(vId, ts, entrypoint, exitpoint, r.graph, r.sensorLayer, r.communicationLayer)
}


func (r *SimulationRunner) spaceAvailable(v *vehicle.VehicleActor) bool {
	edge := v.EntryPoint.EdgesFrom[0] // entryPoint always has exactly 1 edge
	minDist := 100.0
	for _, v := range r.allVehiclesProxy.GetAllActiveVehicles() {
		if edge.Id == v.EdgeAt.Id {
			x := v.EntryPoint.X - v.X
			y := v.EntryPoint.Y - v.Y
			d := math.Sqrt(x * x + y * y)
			minDist = math.Min(minDist, d)
		}
	}

	return minDist > 10
}

func (r *SimulationRunner) cleanUpVehicles(ts types.Millisecond) []*vehicle.VehicleActor {
	pruned := []*vehicle.VehicleActor{}

	allVehicles := r.allVehiclesProxy.GetAllActiveVehicles()
	for i := range allVehicles {

		v := allVehicles[i]
		if v.HasFinished {
			pruned = append(pruned, v)
		}
	}
	for _, v := range pruned {
		r.vehicleIdToLeaveTs[v.Id] = ts
		r.allVehiclesProxy.VehicleFinished(v)
	}

	return pruned
}

func (r *SimulationRunner) getStatitistics() ShortStats {
	timeToIgnore := types.Millisecond(r.configuration.SimulationWarmUp * 1000)
	timeMax := types.Millisecond((r.configuration.SimulationWarmUp + r.configuration.SimulationDuration) * 1000)

	intersectionThroughPut := 0.0
	for _, v := range r.allVehiclesProxy.GetAllVehiclesIntroduced() {
		if v.EnteredIntersection > timeToIgnore && v.ExitedIntersection <  timeMax {
			intersectionThroughPut += 1
		}
	}
	intersectionThroughPut = intersectionThroughPut * 60.0 * 1000 / float64(timeMax - timeToIgnore)

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

	avgDelay := float64(sumDelay) / float64(vehiclesCounter) / 1000
	if vehiclesCounter == 0 {
		avgDelay = 0
	}
	msgDelivered, msgLost, msgDelay := r.communicationLayer.GetStats()
	return ShortStats{
		IntersectionThroughput: intersectionThroughPut,
		VehicleAverageDelay:    avgDelay,
		MessagesLost:           msgLost,
		MessagesDelivered:      msgDelivered,
		MessagesAvgDelay:       msgDelay,
	}

}
