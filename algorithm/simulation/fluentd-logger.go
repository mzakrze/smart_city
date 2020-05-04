package simulation

import (
	"encoding/json"
	"fmt"
	"github.com/fluent/fluent-logger-golang/fluent"
	"math"
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
}

type FluentVehicleBucketLocation struct {
	VehicleId types.VehicleId
	StartSecond types.Timestamp
	Location []types.LocationStruct
	Alpha []float64
	BboxNorth types.Latitude
	BboxSouth types.Latitude
	BboxEast types.Longitude
	BboxWest types.Longitude
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
		var bboxNorth float64 = -math.MaxFloat64
		var bboxSouth float64 = math.MaxFloat64
		var bboxEast float64 = -math.MaxFloat64
		var bboxWest float64 = math.MaxFloat64

		for _, loc := range locationReport.location {
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

		data := &FluentVehicleBucketLocation{
			VehicleId: v.vehicleId,
			StartSecond:locationReport.startSecond,
			Location: locationReport.location[:],
			Alpha: locationReport.alpha[:],
			BboxNorth: bboxNorth,
			BboxSouth: bboxSouth,
			BboxEast: bboxEast,
			BboxWest: bboxWest,
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
				BboxNorth: 1000,
				BboxSouth: -1000,
				BboxEast: -1000,
				BboxWest: 1000,
				// FIXME - te boundy wyliczyć
			}

			f.VehicleBucketLocation(data)
		}
	}
}
