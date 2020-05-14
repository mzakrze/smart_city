package main

import (
	"algorithm2.0/logging"
	"algorithm2.0/simulation"
	"algorithm2.0/util"
	"algorithm2.0/vehicle"
	"algorithm2.0/vehicle_communication"
	"algorithm2.0/vehicle_sensor"
	"github.com/fluent/fluent-logger-golang/fluent"
)

func main() {

	// create dependencies
	configuration, err := util.ReadConfiguration(); if err != nil { panic(err) }
	graph, err := util.ReadGraph(configuration.SimulationName); if err != nil { panic(err) }
	allVehiclesProxy := util.AllVehiclesProxySingleton()
	communicationLayer := vehicle_communication.CommunicationLayerSingleton(allVehiclesProxy)
	sensorLayer := vehicle_sensor.SensorLayerSingleton(allVehiclesProxy)
	collisionDetector := util.NewCollisionDetector(allVehiclesProxy)
	intersectionManager, err := vehicle.IntersectionManagerSingleton(communicationLayer, configuration.IntersectionPolicy); if err != nil { panic(err) }
	fluentLogger, err := fluent.New(fluent.Config{}); if err != nil { panic(err) }; defer fluentLogger.Close()
	resultLogger := logging.ResultsLoggerSingleton(fluentLogger, graph.MapWidth, graph.MapWidth)

	// create runner
	simulationRunner := simulation.SimulationRunnerSingleton(
		configuration,
		graph,
		intersectionManager,
		allVehiclesProxy,
		communicationLayer,
		sensorLayer,
		resultLogger,
		collisionDetector)



	/* --------------------------- */
	/* ----- Run Simulation ------ */
	/* --------------------------- */
	simulationRunner.RunSimulation()



}
