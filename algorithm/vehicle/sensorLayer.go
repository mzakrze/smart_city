package vehicle

import (
	"algorithm/constants"
	"algorithm/types"
	"algorithm/util"
	"math"
)

const MaxDistanceMeasurement = 50.0

func SensorLayerSingleton(proxy *AllVehicleProxy, graph *util.Graph) *SensorLayer {

	if instanceSensor == nil {
		instanceSensor = &SensorLayer{proxy: proxy, graph: graph}
	}
	return instanceSensor
}

func (sensor *SensorLayer) ScanVehiclesAhead (vehicle *VehicleActor) (float64, types.VehicleId) {
	const measurementPrecision = 0.1

	stepX := 0.0
	stepY := 0.0
	offsetX := 0.0
	offsetY := 0.0
	switch vehicle.Alpha {
	case -math.Pi / 2: // up
		stepY = measurementPrecision
		offsetX = constants.VehicleWidth
		offsetY = constants.VehicleLength
	case math.Pi / 2: // down
		stepY = -measurementPrecision
		offsetX = constants.VehicleWidth
		offsetY = constants.VehicleLength
	case 0: // right
		stepX = measurementPrecision
		offsetX = constants.VehicleLength
		offsetY = constants.VehicleWidth
	case math.Pi: // left
		stepX = -measurementPrecision
		offsetX = constants.VehicleLength
		offsetY = constants.VehicleWidth
	default:
		//return -1, -1
		panic("Illegal sensor query")
	}

	var collidedWith *VehicleActor = nil
	collides := func (x, y types.Meter) bool {
		vehicles := sensor.proxy.GetAllActiveVehicles()
		for i, v := range vehicles {
			if v.Id == vehicle.Id {
				continue
			}
			if v.X - offsetX <= x && x <= v.X + offsetX &&
				v.Y - offsetY <= y && y <= v.Y + offsetY {
				collidedWith = vehicles[i]
				return true
			}
		}
		return false
	}

	x := vehicle.X
	y := vehicle.Y
	for step := 1; step < MaxDistanceMeasurement / measurementPrecision ; step += 1 {
		x += stepX
		y += stepY
		if collides(x, y) {
			res := math.Sqrt((vehicle.X - x) * (vehicle.X - x) + (vehicle.Y - y) * (vehicle.Y - y))
			res -= constants.VehicleLength / 2
			res -= math.Sin(collidedWith.Alpha) * constants.VehicleWidth / 2

			if res < 0.0 {
				return 0.0, collidedWith.Id // FIXME - lekki hack - w powyższym jest drobny błąd - źle wyliczana jest odległość w zależności od kąta pojazdu, z którym sie zderza
				//panic("Oops")
			}
			if collidedWith.Id > vehicle.Id {
				return 10, -1
			}
			return res, collidedWith.Id
		}
	}
	return MaxDistanceMeasurement, -1
}


var instanceSensor *SensorLayer = nil
type SensorLayer struct {
	proxy *AllVehicleProxy
	graph *util.Graph
}

