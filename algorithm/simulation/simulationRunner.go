package simulation

import (
	"algorithm/constants"
	"algorithm/logging"
	"algorithm/types"
	"algorithm/util"
	"algorithm/vehicle"
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

	nextVehicleId            types.VehicleId
	vehicleIdToAppearTs      map[types.VehicleId]types.Millisecond
	vehicleIdToLeaveTs       map[types.VehicleId]types.Millisecond
	CurrentTs                types.Millisecond
	vehiclesSpawnedInAMinute int
	lastVehicleSpawnedSecond int
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
	VehicleAvegTravelTime  float64 // seconds
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
			r.activateNewVehicles(ts)
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
				r.resultsLogger.VehicleFinished(v.Id, ts, v.EntryPoint.WayId, v.ExitPoint.WayId)
			}
		}

		ts += constants.SimulationStepInterval

		r.CurrentTs = ts
	}

	r.resultsLogger.SimulationFinished(time.Now())

	return r.getStatitistics()
}

func (r *SimulationRunner) activateNewVehicles(ts  types.Millisecond) {
	for _, e  := range r.graph.Entrypoints {
		next := r.allVehiclesProxy.NextQueued(e.Id)
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
			initSpeed := 4.0
			next.Speed = initSpeed
			r.allVehiclesProxy.RegisterVehicle(next)
			next = r.allVehiclesProxy.NextQueued(e.Id)
		}
	}
}



func (r *SimulationRunner) spawnNewVehicles(ts types.Millisecond) {
	toSpawn := int(r.configuration.VehiclesPerMinute / 60)
	secondInMinute := float64(ts % 60000) / 1000.0
	if (float64(r.configuration.VehiclesPerMinute % 60) / 60.0) * secondInMinute > float64(r.vehiclesSpawnedInAMinute) {
		toSpawn += 1
		r.vehiclesSpawnedInAMinute += 1
	}

	if ts % 60000 == 0 {
		r.vehiclesSpawnedInAMinute = 0
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
		case len(r.graph.Entrypoints):
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
		return possibleExitpoints[rand.Intn(len(possibleExitpoints))]
	}

	entrypoint := r.graph.Entrypoints[rand.Intn(len(r.graph.Entrypoints))]
	exitpoint := getRandomCompatibleExitpoint(entrypoint)

	vId := r.nextVehicleId
	r.nextVehicleId += 1

	return vehicle.NewVehicleActor(vId, ts, entrypoint, exitpoint, r.graph, r.sensorLayer, r.communicationLayer, r.configuration)
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

	return minDist > 20.0
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

	avgTravelTime := float64(sumDelay) / float64(vehiclesCounter) / 1000
	if vehiclesCounter == 0 {
		avgTravelTime = 0
	}
	msgDelivered, msgLost, msgDelay := r.communicationLayer.GetStats()
	return ShortStats{
		IntersectionThroughput: intersectionThroughPut,
		VehicleAvegTravelTime:  avgTravelTime,
		MessagesLost:           msgLost,
		MessagesDelivered:      msgDelivered,
		MessagesAvgDelay:       msgDelay,
	}

}
