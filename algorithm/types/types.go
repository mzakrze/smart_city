package types

type XCoord = float64
type YCoord = float64

type Latitude = float64
type Longitude = float64

type NodeId = int32
type VehicleId = int32
type ImID = int32

type Meter = float64
type Milisecond = int

type VehicleState = int

type RoadCluster = int

const (
	VEHICLE_NOT_STARTED VehicleState = 1 + iota
	VEHICLE_DRIVING
	VEHICLE_FINISHED
)

// represents point on street
type DestinationPoint struct {
	NodeFirst NodeId
	NodeSecond NodeId
	X XCoord
	Y YCoord
	Cluster RoadCluster
}

type LocationStruct struct {
	Lat Latitude
	Lon Longitude
}

type Graph struct {
	AllNodes []Node // FIXME - zamienic na slownik id->Node
	MapBBox MapBBox
	IntersectionManager IntersectionManager

	//StartNodes []*Node // TODO - narazie raczej nie potrzebne
	//EndNodes []*Node
}

type MapBBox struct {
	Width Meter
	Height Meter
	North Latitude
	South Latitude
	West Longitude
	East Longitude
}

type Node struct {
	Id           NodeId
	X            XCoord
	Y            YCoord
	Edges        []Edge
	IsEntrypoint RoadCluster
	IsExitpoint  RoadCluster
}

type Edge struct {
	// TODO - w przyszlosci wynik dzialania IM: latency, delay
	To         *Node
	Distance   Meter
	IsArc      bool
}

type IntersectionManager struct {
	// Intersection Manager's jurisdiction area
	BboxUp Meter
	BboxDown Meter
	BboxLeft Meter
	BboxRight Meter
}


