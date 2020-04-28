package main

import (
	"mzakrze/smart_city/simulation"
)

// TODO - wydzielić tutaj ustawienia

// TODO - wszędzie pododawać aliasy do typów - timestampMs, Latitute, Longitude, itp.

func main() {

	runner := simulation.CreateSimulationRunner()

	runner.RunSimulation()

}