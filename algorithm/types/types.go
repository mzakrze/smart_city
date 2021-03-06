package types

type XCoord = float64
type YCoord = float64

type Latitude = float64
type Longitude = float64

type Millisecond = int32
type Second = int32
type Meter = float64

type VehicleId = int32
type ReservationId = int32
type Angle = float64
type NodeId = int32
type WayId = int32
type EdgeId = int32
type MetersPerSecond = float64
type MetersPerSecond2 = float64

type LocationStruct struct {
	Lat Latitude
	Lon Longitude
}

type Location struct {
	X XCoord
	Y YCoord
}