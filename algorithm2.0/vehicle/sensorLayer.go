package vehicle

import (
	"algorithm2.0/constants"
	"math"
)

const MaxDistanceMeasurment = 100.0

func SensorLayerSingleton(proxy *AllVehicleProxy) *SensorLayer {
	if instanceSensor == nil {
		instanceSensor = &SensorLayer{proxy: proxy}
	}
	return instanceSensor
}


func (sensor *SensorLayer) ScanVehiclesAhead (vehicle *VehicleActor) float64 {
	// aktualna implementacja zaklada że pojazd porusza się pod kątem 0, albo 90, albo 180, albo 270

	var getDist func (actor *VehicleActor) float64 = nil

	//switch vehicle.Alpha {
	//case math.Pi / 2: // riding down
	//	getDist = func(other *VehicleActor) float64 {
	//		if vehicle.X == other.X && other.Y < vehicle.Y {
	//			return vehicle.Y - other.Y
	//		}
	//		return MaxDistanceMeasurment
	//	}
	//case -math.Pi / 2: // riding up
	//	getDist = func(other *VehicleActor) float64 {
	//		if vehicle.X == other.X && other.Y > vehicle.Y {
	//			return other.Y - vehicle.Y
	//		}
	//		return MaxDistanceMeasurment
	//	}
	//case 0: // riding right
	//	getDist = func(other *VehicleActor) float64 {
	//		if vehicle.Y == other.Y && other.X > vehicle.X {
	//			return other.X - vehicle.X
	//		}
	//		return MaxDistanceMeasurment
	//	}
	//case -math.Pi: // riding left
	//	getDist = func(other *VehicleActor) float64 {
	//		if vehicle.Y == other.Y && other.X < vehicle.X {
	//			return vehicle.X - other.X
	//		}
	//		return MaxDistanceMeasurment
	//	}
	//default:
	//	// FIXME
	//	fmt.Printf("Warning: scaning doesnt work: %f\n", vehicle.Alpha)
	//	return MaxDistanceMeasurment
	//}

	x := vehicle.X - vehicle.EdgeAt.To.X
	y := vehicle.Y - vehicle.EdgeAt.To.Y

	d := math.Sqrt(x * x + y * y)

	getDist = func (actor *VehicleActor) float64 {
		if vehicle.EdgeAt.Id != actor.EdgeAt.Id {
			return MaxDistanceMeasurment
		}
		x1 := actor.X - actor.EdgeAt.To.X
		y1 := actor.Y- actor.EdgeAt.To.Y
		d1 := math.Sqrt(x1 * x1 + y1 * y1)
		if d < d1 {
			return MaxDistanceMeasurment
		}
		return d - d1 - constants.VehicleLength
	}

	closest := MaxDistanceMeasurment
	for _, v := range sensor.proxy.GetAllVehicles() {
		if v.Id != vehicle.Id {
			closest = math.Min(getDist(v), closest)
		}
	}
	return closest
}


var instanceSensor *SensorLayer = nil
type SensorLayer struct {
	proxy *AllVehicleProxy
}
