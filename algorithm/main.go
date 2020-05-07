package main

import (
	"mzakrze/smart_city/simulation"
	"net/http"
	"net/url"
)

// TODO - wydzielić tutaj ustawienia

// TODO - wszędzie pododawać aliasy do typów - timestampMs, Latitute, Longitude, itp.

func main() {

	pruneOldIndicesInElastic()

	runner := simulation.CreateSimulationRunner()

	runner.RunSimulation()

}

func pruneOldIndicesInElastic() {
	client := &http.Client{}

	urlLog := &url.URL{
		Scheme:  "http",
		Host: "localhost:9200",
		Path: "simulation-log-1",
	}

	urlMap := &url.URL{
		Scheme:  "http",
		Host: "localhost:9200",
		Path: "simulation-map-1",
	}

	urlTrip := &url.URL{
		Scheme:  "http",
		Host: "localhost:9200",
		Path: "simulation-trip-1",
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