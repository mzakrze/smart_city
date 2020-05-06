package simulation

import (
	"encoding/json"
	"fmt"
	"github.com/fluent/fluent-logger-golang/fluent"
	"mzakrze/smart_city/types"
)

// TODO - nie exportować tych typów poza pakiet

type FluentVehicleCurrentLocation struct {
	VehicleId types.VehicleId
	Timestamp types.Timestamp
	Lat       types.Latitude
	Lon       types.Longitude
	Alpha     float64
	Speed     float64
	MoveX     float64
	MoveY     float64
}

type FluentVehicleBucketLocation struct {
	VehicleId types.VehicleId
	StartSecond types.Timestamp
	Location []types.LocationStruct
	Alpha []float64
}

type FluentVehicleTrip struct {
	VehicleId types.VehicleId
	StartTs types.Timestamp
	EndTs types.Timestamp
	Width types.Meter
	Length types.Meter
	//OriginLat float64
	//OriginLon float64
	//DestinationLat float64
	//DestinationLon float64
}

type FluentLogger struct {
	logger *fluent.Fluent
	latToY func(types.Latitude) types.YCoord
	lonToX func(types.Longitude) types.XCoord
	yToLat func(coord types.YCoord) types.Latitude
	xToLon func(coord types.XCoord) types.Longitude

	tripReportSent map[types.VehicleId]bool
	locationReport map[types.VehicleId]*VehicleLocationReport


	// TODO - jakis loadbalancer moze
}

type VehicleLocationReport struct {
	location    [STEPS_IN_SECOND]types.LocationStruct
	// aktualnie wysylanie 1 raz na sekunde
	alpha       [STEPS_IN_SECOND]float64
	step        int
	startSecond int64
	isEmpty bool // TODO - przepisac to
}

func NewFluentdLogger(latToY func(types.Latitude) types.YCoord,
		lonToX func(types.Longitude) types.XCoord,
		yToLat func(coord types.YCoord) types.Latitude,
		xToLon func(coord types.XCoord) types.Longitude) *FluentLogger {
	var logger, err = fluent.New(fluent.Config{})
	if err != nil {
		panic(err)
	}
	locationReport := make(map[types.VehicleId]*VehicleLocationReport)
	return &FluentLogger{logger: logger,
		latToY: latToY, lonToX: lonToX,
		yToLat: yToLat, xToLon: xToLon,
		locationReport: locationReport,
		tripReportSent: make(map[types.VehicleId]bool)}
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
		"alpha": fmt.Sprintf("%f", data.Alpha),
		"movex": fmt.Sprintf("%f", data.MoveX),
		"movey": fmt.Sprintf("%f", data.MoveY),
		"speed": fmt.Sprintf("%f", data.Speed),
	}

	f.doSend("vehicle.log", msg)
}

func (f *FluentLogger) VehicleBucketLocation(data *FluentVehicleBucketLocation) {
	bytes, err := json.Marshal(data.Location)
	if err != nil { panic(err) }
	locationJson := string(bytes)

	bytesAlpha, err := json.Marshal(data.Alpha)
	if err != nil { panic(err) }
	alphaJson := string(bytesAlpha)

	msg := map[string]string {
		"vehicle_id": fmt.Sprintf("%d", data.VehicleId),
		"start_second": fmt.Sprintf("%d", data.StartSecond),
		"location_array": locationJson,
		"alpha_array": alphaJson,
	}

	// fmt.Printf("Sending map, vehicle_id = %d\n", data.VehicleId)
	f.doSend("vehicle.map", msg)
}

func (f *FluentLogger) VehicleTrip(data *FluentVehicleTrip) {

	msg := map[string]string {
		"vehicle_id": fmt.Sprintf("%d", data.VehicleId),
		"start_ts": fmt.Sprintf("%d", data.StartTs),
		"end_ts": fmt.Sprintf("%d", data.EndTs),
		"vehicle_width": fmt.Sprintf("%f", data.Width),
		"vehicle_length": fmt.Sprintf("%f", data.Length),
		//"origin_lat": fmt.Sprintf("%f", data.OriginLat),
		//"origin_lon": fmt.Sprintf("%f", data.OriginLon),
		//"destination_lat": fmt.Sprintf("%f", data.DestinationLat),
		//"destination_lon": fmt.Sprintf("%f", data.DestinationLon),
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

func (f *FluentLogger) ReportVehicle(ts types.Timestamp, controller *VehicleController) {
	if controller.VehicleState == types.VEHICLE_NOT_STARTED {
		return
	}
	if f.tripReportSent[controller.vehicleId] {
		return
	}

	if controller.VehicleState == types.VEHICLE_FINISHED && f.tripReportSent[controller.vehicleId] == false {
		data := &FluentVehicleTrip{
			VehicleId: controller.vehicleId,
			StartTs: controller.startTs,
			EndTs: controller.endTs,
			Width: controller.vehicleActor.Width,
			Length: controller.vehicleActor.Length,
			//OriginLat: v.originLat,
			//OriginLon: v.originLon,
			//DestinationLat: v.destinationLat,
			//DestinationLon: v.destinationLon,
		}

		f.VehicleTrip(data)
		f.tripReportSent[controller.vehicleId] = true
	} else {
		f.reportCurrentLocation(controller, ts)
		f.appendLocationToBucket(controller, ts)
		f.flushBucketIfFull(controller, ts)
	}

}


func (f *FluentLogger) reportCurrentLocation(v* VehicleController, ts types.Timestamp) {
	data := &FluentVehicleCurrentLocation{
		VehicleId: v.vehicleId,
		Timestamp: ts,
		Lat: f.yToLat(v.vehicleActor.Y),
		Lon: f.xToLon(v.vehicleActor.X),
		Speed: v.vehicleActor.Speed_mps,
	}
	f.VehicleCurrentLocation(data)
}


func (f *FluentLogger) flushBucketIfFull(v *VehicleController, ts int64) {
	locationReport := f.locationReport[v.vehicleId]
	if locationReport.step == cap(locationReport.location) {

		data := &FluentVehicleBucketLocation{
			VehicleId: v.vehicleId,
			StartSecond:locationReport.startSecond,
			Location: locationReport.location[:],
			Alpha: locationReport.alpha[:],
		}

		f.VehicleBucketLocation(data)

		locationReport.isEmpty = true
		locationReport.step = 0
	}
}

	func (f *FluentLogger) appendLocationToBucket(v* VehicleController, ts types.Timestamp) {
	locationReport := f.locationReport[v.vehicleId]
	if locationReport == nil {
		locationReport = &VehicleLocationReport{
			location: [STEPS_IN_SECOND]types.LocationStruct{},
			alpha: [STEPS_IN_SECOND]float64{},
			isEmpty: true,
		}
		f.locationReport[v.vehicleId] = locationReport
	}
	if locationReport.isEmpty {
		locationReport.startSecond = ts / 1000
		locationReport.step = 0
		locationReport.isEmpty = false
	}

	locationReport.location[locationReport.step] = types.LocationStruct{
		Lat: f.yToLat(v.vehicleActor.Y),
		Lon: f.xToLon(v.vehicleActor.X),
	}
	locationReport.alpha[locationReport.step] = v.vehicleActor.Alpha
	locationReport.step += 1
}

func (f *FluentLogger) flushAllReports(ts int64) {
	//locationReport := f.locationReport[v.vehicleId]
	for vehicleId, locationReport := range f.locationReport {
		if locationReport.isEmpty == false {
			data := &FluentVehicleBucketLocation{
				VehicleId: vehicleId,
				StartSecond: locationReport.startSecond,
				Location: locationReport.location[:],
				Alpha: locationReport.alpha[:],
			}

			f.VehicleBucketLocation(data)
		}
	}
}
