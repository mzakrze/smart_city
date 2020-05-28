package main

import (
	"algorithm/logging"
	"algorithm/simulation"
	"algorithm/util"
	"algorithm/vehicle"
	"flag"
	"github.com/fluent/fluent-logger-golang/fluent"
	"math/rand"
	"net/http"
	"net/url"
)

func main() {
	dockerEnv := flag.Bool( "docker", false, "Is run in docker (effects host of elastic and fluent)")
	flag.Parse()

	var elasticHost, fluentHost string
	if *dockerEnv {
		elasticHost = "elastic"
		fluentHost = "fluentd"
	} else {
		elasticHost = "localhost"
		fluentHost = "localhost"
	}

	//rand.Seed(time.Now().Unix())
	rand.Seed(18)

	pruneOldIndicesInElastic(elasticHost)

	// create dependencies
	configuration, err := util.ReadConfiguration(); if err != nil { panic(err) }
	graph, err := util.ReadGraph(configuration.SimulationName, elasticHost); if err != nil { panic(err) }
	allVehiclesProxy := vehicle.AllVehiclesProxySingleton()
	communicationLayer := vehicle.CommunicationLayerSingleton(allVehiclesProxy)
	sensorLayer := vehicle.SensorLayerSingleton(allVehiclesProxy, graph)
	collisionDetector := vehicle.NewCollisionDetector(allVehiclesProxy)
	intersectionManager, err := vehicle.IntersectionManagerSingleton(graph, communicationLayer, configuration); if err != nil { panic(err) }
	fluentLogger, err := fluent.New(fluent.Config{FluentHost: fluentHost}); if err != nil { panic(err) }; defer fluentLogger.Close()
	resultLogger := logging.ResultsLoggerSingleton(fluentLogger, graph.MapWidth, graph.MapWidth, configuration.SimulationDuration.Seconds())

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



	go runProgressTracker(simulationRunner, &configuration)

	/* --------------------------- */
	/* ----- Run Simulation ------ */
	/* --------------------------- */
	simulationRunner.RunSimulation()




}


func runProgressTracker(s *simulation.SimulationRunner, c *util.Configuration) {

	//ctx := context.Background()
	//s := `Now that's what I call progress`
	//r := progress.NewReader(strings.NewReader("100"))
	//go func() {
	//	progressChan := progress.NewTicker(ctx, r, 100, 1*time.Second)
	//	<-progressChan
	//	for p := <-progressChan; ; {
	//		fmt.Printf("\r%v remaining...",
	//			p.Remaining().Round(time.Second))
	//	}
	//	fmt.Println("\rdownload is completed")
	//}()

	//bar := pb.StartNew(100)
	//
	//for true {
	//	time.Sleep(time.Millisecond * 500)
	//	p := int64(float64(s.CurrentTs) / (c.SimulationDuration.Seconds() * 1000) * 100)
	//	if p > bar.Current() {
	//		bar.Increment()
	//	}
	//	if p == 100 {
	//		break
	//	}
	//}
	//bar.Finish()

}


func pruneOldIndicesInElastic(host string) {
	client := &http.Client{}

	urlLog := &url.URL{
		Scheme:  "http",
		Host: host + ":9200",
		Path: "simulation-map",
	}

	urlMap := &url.URL{
		Scheme:  "http",
		Host: host + ":9200",
		Path: "simulation-vehicle",
	}

	urlTrip := &url.URL{
		Scheme:  "http",
		Host: host + ":9200",
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
