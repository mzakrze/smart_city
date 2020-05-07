package algorithm

import (
	"fmt"
	"math"
)

type VehicleAheadSensor struct {
	AllVehicles []*VehicleActor
}

func (sensor *VehicleAheadSensor) ScanVehiclesAhead (vehicle *VehicleActor) float64 {
	// aktualna implementacja zaklada że pojazd porusza się pod kątem 0, albo 90, albo 180, albo 270

	var getDist func (actor *VehicleActor) float64 = nil

	switch vehicle.Alpha {
	case math.Pi / 2: // riding down
		getDist = func(other*VehicleActor) float64 {
			if vehicle.X == other.X && other.Y < vehicle.Y {
				return vehicle.Y - other.Y
			}
			return 100.0
		}
	case -math.Pi / 2: // riding up
		getDist = func(other*VehicleActor) float64 {
			if vehicle.X == other.X && other.Y > vehicle.Y {
				return other.Y - vehicle.Y
			}
			return 100.0
		}
	case 0: // riding right
		getDist = func(other*VehicleActor) float64 {
			if vehicle.Y == other.Y && other.X > vehicle.X {
				return other.X - vehicle.X
			}
			return 100.0
		}
	case -math.Pi: // riding left
		getDist = func(other*VehicleActor) float64 {
			if vehicle.Y == other.Y && other.X < vehicle.X {
				return vehicle.X - other.X
			}
			return 100.0
		}
	default:
		fmt.Printf("Warning: scaning doesnt work: %f\n", vehicle.Alpha)
		return 100.0
	}

	closest := 100.0
	for _, v := range sensor.AllVehicles {
		if d := getDist(v); d < closest {
			closest = d
		}
	}
	return closest




	//closest := 100.0
	//
	//for _, v := range sensor.AllVehicles  {
	//	if v.Id == vehicle.Id {
	//		continue
	//	}
	//	if v.HasFinished {
	//		continue
	//	}
	//
	//	if v.Route[0].Node.Id == vehicleNodeDest.Id {
	//		// oba pojazdy jadą do tego samego node'a
	//		vSlope := (vehicleNodeDest.X - v.X) / (vehicleNodeDest.Y - v.Y)
	//		if math.Abs(vSlope - vehicleSlope) < 0.00001 {
	//			// oba pojazdy jadą w tym samym kierunku
	//
	//			x1 := vehicle.X - vehicleNodeDest.X
	//			y1 := vehicle.Y - vehicleNodeDest.Y
	//
	//			x2 := v.X - vehicleNodeDest.X
	//			y2 := v.Y - vehicleNodeDest.Y
	//
	//			if math.Sqrt(x1 * x1 + y1 * y1) > math.Sqrt(x2 * x2 + y2 * y2) {
	//				// vehicle jest za v <=> vehicle ma szanse wjechać w v
	//
	//				x := v.X - vehicle.X
	//				y := v.Y - vehicle.Y
	//
	//				distance := math.Sqrt(x * x + y * y) - v.Length / 2 - vehicle.Length / 2 + 5
	//
	//				closest = math.Min(distance, closest)
	//			}
	//		}
	//	}
	//}
	//
	//
	//return closest
}




















