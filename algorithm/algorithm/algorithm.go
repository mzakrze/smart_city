package algorithm

import (
	"mzakrze/smart_city/types"
)

const (
	VEHICLE_MASS = 1500
	VEHICLE_POWER = 150
	VEHICLE_BREAKING_POWER = 400
)

// TODO - to docelowo oddzielny pakiet, żeby na pewno nie widziało simulationRunnera'a 
// albo nawet lepiej - jeśli się da - oddzielna binarka

type VehicleActor struct {
	Origin types.DestinationPoint
	Destination types.DestinationPoint

	X types.XCoord
	Y types.YCoord

	Speed_mps float64
}

func InitVehicleActor(origin, destination types.DestinationPoint) *VehicleActor {
	// TODO - te wartosci gdzies wyniesc do configa
	return &VehicleActor{
		Origin: origin,
		Destination: destination,
		X: origin.X,
		Y: origin.Y,
		Speed_mps: 10,
	}

	// TODO - dijkistra
	// TODO - wlaczanie sie do ruchu
}

// mocked algorithm
func (v *VehicleActor) Ping(ts types.Timestamp) {

	if v.Destination.X - v.X < 0 {
		v.X -= 1
	} else {
		v.X += 1
	}

	if v.Destination.Y - v.Y < 0 {
		v.Y -= 1
	} else {
		v.Y += 1
	}


}

