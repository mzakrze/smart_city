package algorithm

import "math"

// TODO - to docelowo oddzielny pakiet, żeby na pewno nie widziało simulationRunnera'a 
// albo nawet lepiej - jeśli się da - oddzielna binarka

type VehicleActor struct {
	Lat float64
	Lon float64

	// też OriginLat/Lon

	DestinationLat float64
	DestinationLon float64

	Speed_mps float64
}

// mocked algorithm
func (v *VehicleActor) Ping(ts int64) {
	if v.Speed_mps == 0.0 {
		v.Speed_mps = 10.0
	}

	// x := v.Lat - v.DestinationLat
	// y := v.Lon - v.DestinationLon

	// latDiffMeters := 0.0 	
	// lonDiffMeters := 0.0
	
	// if math.Abs(x) > math.Abs(y) {
	// 	latDiffMeters = v.Speed_mps / 10.0
	// } else {
	// 	lonDiffMeters = v.Speed_mps / 10.0
	// }

	/*
	Length in meters of 1° of latitude = always 111.32 km
	Length in meters of 1° of longitude = 40075 km * cos( latitude ) / 360
	*/
	// latDiff := latDiffMeters / 111320.0
	// lonDiff := math.Abs(lonDiffMeters * 3600.0 / 40075000.0 / math.Cos(v.Lat))

	latDiff := 0.00005
	lonDiff := 0.00005

	if latDiff < 0 {
		panic("lat diff must be positive")
	}
	if lonDiff < 0 {
		panic("lon diff must be positive")
	}

	if math.Abs(v.Lat - v.DestinationLat) > latDiff {
		if v.Lat < v.DestinationLat {
			v.Lat += latDiff
		} else {
			v.Lat -= latDiff
		}
	} 

	if math.Abs(v.Lon - v.DestinationLon) > lonDiff {
		if v.Lon < v.DestinationLon {
			v.Lon += lonDiff
		} else {
			v.Lon -= lonDiff
		}
	}


}

