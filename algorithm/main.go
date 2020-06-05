package main

import (
	"algorithm/logging"
	"algorithm/simulation"
	"algorithm/util"
	"algorithm/vehicle"
	"flag"
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"github.com/fluent/fluent-logger-golang/fluent"
	"math/rand"
	"time"
)

func main() {
	dockerEnv := flag.Bool( "docker", false, "Is run in docker (effects host of elastic and fluent)")
	quiet := flag.Bool( "quiet", false, "Quiet progress bar")
	logsOff := flag.Bool( "logsOff", false, "Safe vehicles location for visualization (store results in Elasticsearch)")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Invalid use. Specify configuration file")
	}
	confFilePath := flag.Args()[0]

	var elasticHost, fluentHost string
	if *dockerEnv {
		elasticHost = "elastic"
		fluentHost = "fluentd"
	} else {
		elasticHost = "localhost"
		fluentHost = "localhost"
	}

	configuration, err := util.ReadConfiguration(confFilePath); if err != nil { panic(err) }
	if configuration.RandomSeed != 0 {
		rand.Seed(configuration.RandomSeed)
	} else {
		rand.Seed(time.Now().Unix())
	}

	util.ClearOldIndicesInElastic(elasticHost)

	// create dependencies
	graph, err := util.ReadGraph(elasticHost); if err != nil { panic(err) }
	vehicle.Initiate(configuration)
	allVehiclesProxy := vehicle.AllVehiclesProxySingleton()
	communicationLayer := vehicle.CommunicationLayerSingleton(allVehiclesProxy, configuration)
	sensorLayer := vehicle.SensorLayerSingleton(allVehiclesProxy, graph)
	collisionDetector := vehicle.NewCollisionDetector(allVehiclesProxy)
	intersectionManager, err := vehicle.IntersectionManagerSingleton(graph, communicationLayer, configuration); if err != nil { panic(err) }
	resultLogger := logging.NewNoOpLogger()
	if *logsOff == false {
		fluentLogger, err := fluent.New(fluent.Config{FluentHost: fluentHost}); if err != nil { panic(err) }; defer fluentLogger.Close()
		resultLogger = logging.ResultsLoggerSingleton(fluentLogger, graph.MapWidth, graph.MapWidth, configuration.SimulationDuration)
	}

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

	if ! *quiet {
		go runProgressTracker(simulationRunner, &configuration)
	}


	/* --------------------------- */
	/* ----- Run Simulation ------ */
	/* --------------------------- */
	stats := simulationRunner.RunSimulation()


	fmt.Println("Simulation finished")
	fmt.Println("Intersection throughput:",stats.IntersectionThroughput)
	fmt.Println("Vehicle delay:",stats.VehicleAverageDelay)
	fmt.Println("Messages delivered no:",stats.MessagesDelivered)
	fmt.Println("Messages lost no:",stats.MessagesLost)
	fmt.Println("Messages average delay:",stats.MessagesAvgDelay)
}


func runProgressTracker(s *simulation.SimulationRunner, c *util.Configuration) {

	seconds := c.SimulationDuration + c.SimulationWarmUp

	bar := pb.StartNew(seconds)

	for int(s.CurrentTs) < seconds * 1000 {
		time.Sleep(time.Millisecond * 500)

		for bar.Current() <= int64(s.CurrentTs / 1000) {
			bar.Increment()
		}

	}

	bar.Finish()
}

