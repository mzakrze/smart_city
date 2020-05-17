package main

import (
	"algorithm2.0/logging"
	"algorithm2.0/simulation"
	"algorithm2.0/util"
	"algorithm2.0/vehicle"
	"github.com/fluent/fluent-logger-golang/fluent"
	"net/http"
	"net/url"
)

func main() {

	pruneOldIndicesInElastic()

	// create dependencies
	configuration, err := util.ReadConfiguration(); if err != nil { panic(err) }
	graph, err := util.ReadGraph(configuration.SimulationName); if err != nil { panic(err) }
	allVehiclesProxy := vehicle.AllVehiclesProxySingleton()
	communicationLayer := vehicle.CommunicationLayerSingleton(allVehiclesProxy)
	sensorLayer := vehicle.SensorLayerSingleton(allVehiclesProxy)
	collisionDetector := vehicle.NewCollisionDetector(allVehiclesProxy)
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


func pruneOldIndicesInElastic() {
	client := &http.Client{}

	urlLog := &url.URL{
		Scheme:  "http",
		Host: "localhost:9200",
		Path: "simulation-map",
	}

	urlMap := &url.URL{
		Scheme:  "http",
		Host: "localhost:9200",
		Path: "simulation-vehicle",
	}

	urlTrip := &url.URL{
		Scheme:  "http",
		Host: "localhost:9200",
		Path: "simulation-intersection",
	}

	_, err := client.Do(&http.Request{
		Method: http.MethodDelete,
		URL: urlLog,
	})
	if err != nil {
		panic(err)
	}

	_, err = client.Do(&http.Request{
		Method: http.MethodDelete,
		URL: urlMap,
	})
	if err != nil {
		panic(err)
	}

	_, err = client.Do(&http.Request{
		Method: http.MethodDelete,
		URL: urlTrip,
	})
	if err != nil {
		panic(err)
	}
}
