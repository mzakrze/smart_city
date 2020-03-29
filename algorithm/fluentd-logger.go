package main

import (
	"fmt"
	"github.com/fluent/fluent-logger-golang/fluent"
	"encoding/json"
)

type FluentVehicleCurrentLocation struct {
	VehicleId int
	Timestamp int64
	Lat float64
	Lon float64
	Speed float64
} 

type FluentVehicleBucketLocation struct {
	VehicleId int 
	StartSecond int64
	Location []LocationStruct
	BboxNorth float64
	BboxSouth float64
	BboxEast float64
	BboxWest float64 
}

type FluentVehicleTrip struct {
	VehicleId int
	StartTs int64
	EndTs int64
	OriginLat float64
	OriginLon float64
	DestinationLat float64 
	DestinationLon float64
}

type FluentLogger struct {
	logger *fluent.Fluent
	// TODO - jakis loadbalancer moze
}

func NewFluentdLogger() *FluentLogger {
	var logger, err = fluent.New(fluent.Config{})
	if err != nil {
		panic(err)
	}
	return &FluentLogger{logger: logger}
}

// TODO - nie wiem gdzie indziej to dać, jakiś destruktor???
func (f *FluentLogger) Close() {
	f.logger.Close()
}

func (f *FluentLogger) VehicleCurrentLocation(data *FluentVehicleCurrentLocation) {
	location := fmt.Sprintf("{\"lat\": %f, \"lon\": %f}", data.Lat, data.Lon)
	msg := map[string]string{
		"vehicle_id": fmt.Sprintf("%d", data.VehicleId),
		"@timestamp": fmt.Sprintf("%d", data.Timestamp),
		"location": location,
		"speed": fmt.Sprintf("%f", data.Speed),
	}

	f.doSend("vehicle.log", msg)
}

func (f *FluentLogger) VehicleBucketLocation(data *FluentVehicleBucketLocation) {
	bytes, err := json.Marshal(data.Location)
	if err != nil {
		panic(err)
	}
	locationJson := string(bytes)

	msg := map[string]string {
		"vehicle_id": fmt.Sprintf("%d", data.VehicleId),
		"start_second": fmt.Sprintf("%d", data.StartSecond),
		"location_array": locationJson,
		"bbox_north": fmt.Sprintf("%f", data.BboxNorth),
		"bbox_south": fmt.Sprintf("%f", data.BboxSouth),
		"bbox_east": fmt.Sprintf("%f", data.BboxEast),
		"bbox_west": fmt.Sprintf("%f", data.BboxWest),
	}

	// fmt.Printf("Sending map, vehicle_id = %d\n", data.VehicleId)
	f.doSend("vehicle.map", msg)
}

func (f *FluentLogger) VehicleTrip(data *FluentVehicleTrip) {

	msg := map[string]string {
		"vehicle_id": fmt.Sprintf("%d", data.VehicleId),
		"start_ts": fmt.Sprintf("%d", data.StartTs),
		"end_ts": fmt.Sprintf("%d", data.EndTs),
		"origin_lat": fmt.Sprintf("%f", data.OriginLat),
		"origin_lon": fmt.Sprintf("%f", data.OriginLon),
		"destination_lat": fmt.Sprintf("%f", data.DestinationLat),
		"destination_lon": fmt.Sprintf("%f", data.DestinationLon),
	}

	f.doSend("vehicle.trip", msg)
}



/***************  private stuff  ************/

func (f *FluentLogger) doSend(tag string, msg map[string]string) {
	err := f.logger.Post(tag, msg)
	if err != nil {
		panic(err)
	}
}



