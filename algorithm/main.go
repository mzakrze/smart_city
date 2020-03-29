package main

import (
	"github.com/fluent/fluent-logger-golang/fluent"
	"fmt"
	"time"
	"math/rand"
	"math"
	"encoding/json"
)

const (
	maxLat = 52.258071
	maxLon = 21.067436
	minLat = 52.212656
	minLon = 20.937030

	steps_no = 2000
	step_interval_ms = 100 // TODO - wymusic podzielne przez 1000

	vehicles_no = 500
)

type Vehicle struct {
	carId int
	lat float64
	lon float64
	speedMPerS float64
	acc float64
	dirX float64 // TODO - rename dirLat, dirLon
	dirY float64

	locationReport *VehicleLocationReport
}

type VehicleLocationReport struct {
	carId int
	location []LocationStruct
	startSecond int64
} 

type LocationStruct struct {
	Lat float64
	Lon float64
}

func (v* Vehicle) move() {
	latDiffMeters := v.speedMPerS * v.dirX / 10.0
	lonDiffMeters := v.speedMPerS * v.dirY / 10.0

	/*
	Length in meters of 1° of latitude = always 111.32 km
	Length in meters of 1° of longitude = 40075 km * cos( latitude ) / 360
	*/

	// 1 lat = 1m * 111320
	// 1m 
	// 1 lon = 1m * 40075000.0 * cos (lat) / 360.0

	// 1m = 1 / 40075000.0 / cos(lat) * 3600

	latDiff := latDiffMeters / 111320.0
	lonDiff := lonDiffMeters * 3600.0 / 40075000.0 / math.Cos(v.lat) 

	v.lat += latDiff;
	v.lon += lonDiff;

	if v.lat > maxLat || v.lat < minLat {
		v.dirX *= -1
	}
	if v.lon > maxLon || v.lat < minLon {
		v.dirY *= -1
	}
}

func (v* Vehicle) reportLocation() {
	var ts int64 = current_time.UnixNano() / 1000000
	data := struct {
		Speed float64 //3.14 //v.speed
		Lat float64
		Lon float64
		Car_no int 
		Timestamp int64
	} {
		v.speedMPerS,
		v.lat,
		v.lon,
		v.carId,
		ts,
	}

	error := logger.Post("vehicle.log", data)
	if error != nil {
	  panic(error)
	}

	if v.locationReport == nil {
		v.locationReport = &VehicleLocationReport{
			carId: v.carId,
			location: make([]LocationStruct, 0),
			startSecond: ts / 1000,
		}
	}

	v.locationReport.location = append(v.locationReport.location, LocationStruct{
		Lat: v.lat,
		Lon: v.lon,
	})

	// TODO - tutaj wyliczać ile próbek w sekundzie
	if len(v.locationReport.location) == 1000 / step_interval_ms {

		locationJSON, err := json.Marshal(v.locationReport.location)
		if err != nil {
			panic(err)
		}

		var bboxNorth float64 = -1000
		var bboxSouth float64 = 1000
		var bboxEast float64 = -1000
		var bboxWest float64 = 1000

		for _, loc := range v.locationReport.location {
			if loc.Lat > bboxNorth {
				bboxNorth = loc.Lat
			}
			if loc.Lat < bboxSouth {
				bboxSouth = loc.Lat
			}
			if loc.Lon > bboxEast {
				bboxEast = loc.Lon
			}
			if loc.Lon < bboxWest {
				bboxWest = loc.Lon
			}
		}

		data2 := struct {
			CarId int 
			StartSecond int64
			LocationJSON []byte
			BboxNorth float64
			BboxSouth float64
			BboxEast float64
			BboxWest float64 
		} {
			v.carId,
			v.locationReport.startSecond,
			locationJSON,
			bboxNorth,
			bboxSouth,
			bboxEast,
			bboxWest,
		}

		error := logger.Post("vehicle.map", data2)
		if error != nil {
			panic(error)
		}

		v.locationReport = nil
	} 
}

func RandomVehicle(carId int) *Vehicle {
	var dir_coordinate float64 = rand.Float64() * 10
	return &Vehicle{
		carId: carId,
		lat: rand.Float64() * (maxLat - minLat) + minLat,
		lon: rand.Float64() * (maxLon - minLon) + minLon,
		speedMPerS: rand.Float64() * 1,
		acc: 0.0,
		dirX: dir_coordinate,
		dirY: 10.0 - dir_coordinate,
	}
}

var current_time = time.Now()
var logger, err = fluent.New(fluent.Config{})


func main() {
	fmt.Println("Hello")
	defer logger.Close()
	// logger, err := fluent.New(fluent.Config{})
	// if err != nil {
	//   fmt.Println(err)
	// }
	

	step_duration , _ := time.ParseDuration(fmt.Sprintf("%dms", step_interval_ms))

	vehicles := [vehicles_no]*Vehicle{}
	for i:=0; i<vehicles_no; i += 1 {
		vehicles[i] = RandomVehicle(i)
	}

	for step := 1; step <= steps_no; step += 1 {
		for i:=0; i<vehicles_no; i += 1 {
			vehicles[i].move()
			vehicles[i].reportLocation()
		}

		current_time = current_time.Add(step_duration)
	}

}