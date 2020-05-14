package simulation

import (
	"algorithm2.0/logging"
	"algorithm2.0/util"
	"algorithm2.0/vehicle"
	"algorithm2.0/vehicle_communication"
	"algorithm2.0/vehicle_sensor"
)

type SimulationRunner struct {
	configuration util.Configuration
	graph *util.Graph
	im *vehicle.IntersectionManager
	allVehiclesProxy *util.AllVehicleProxy
	communicationLayer *vehicle_communication.CommunicationLayer
	sensorLayer *vehicle_sensor.SensorLayer
	resultsLogger *logging.ResultsLogger
	collisionDetector *util.CollisionDetector
}


var instance *SimulationRunner = nil
func SimulationRunnerSingleton(
		configuration util.Configuration,
		graph *util.Graph,
		intersectionManager *vehicle.IntersectionManager,
		allVehiclesProxy *util.AllVehicleProxy,
		communicationLayer *vehicle_communication.CommunicationLayer,
		sensorLayer *vehicle_sensor.SensorLayer,
		resultsLogger *logging.ResultsLogger,
		collisionDetector *util.CollisionDetector,
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

	// Premise: vehicle must appear and disappear only on full second (where millisecond part is 000 + next step)





}
