package algorithm

import (
	"fmt"
	"math"
	"mzakrze/smart_city/types"
)

const (
	VEHICLE_MASS = 1500
	VEHICLE_POWER = 150
	VEHICLE_BREAKING_POWER = 400
)

// FIXME - to jest skopiowane z simulationrunner, docelowo trzeba jakoś przepychać
const STEP_INTERVAL_MS = 100


// TODO - to docelowo oddzielny pakiet, żeby na pewno nie widziało simulationRunnera'a
// albo nawet lepiej - jeśli się da - oddzielna binarka

type VehicleActor struct {
	Origin      types.DestinationPoint
	Destination types.DestinationPoint

	// represents the center of the vehicle
	X types.XCoord
	Y types.YCoord
	Speed_mps float64
	Alpha float64
	Width types.Meter
	Length types.Meter

	HasFinished bool

	roadGraph *types.Graph
	route     []*types.Node
}


func InitVehicleActor(origin, destination types.DestinationPoint, roadGraph *types.Graph) *VehicleActor {
	res := &VehicleActor{
		Origin: origin,
		Destination: destination,
		X: origin.X,
		Y: origin.Y,
		Speed_mps: 5,
		Width: 1.981,
		Length: 4.636,
		roadGraph: roadGraph,
	}

	res.planRoute()

	return res
	// TODO - wlaczanie sie do ruchu
}

func (v *VehicleActor) move(distSpare types.Meter) {

	dest := v.route[0]
	beforeX := v.X
	beforeY := v.Y

	xDiff := dest.X - v.X
	yDiff := dest.Y - v.Y
	d := math.Sqrt(xDiff * xDiff + yDiff * yDiff)

	if d <= distSpare {
		distSpare -= d

		v.X = dest.X
		v.Y = dest.Y

		if len(v.route) == 1 {
			v.HasFinished = true
			return
		}

		v.route = v.route[1:]
		if distSpare > 0 {
			v.move(distSpare)
		}

		return
	}

	moveX := distSpare * xDiff / d
	moveY := distSpare * yDiff / d

	v.X += moveX
	v.Y += moveY
	v.Alpha = math.Acos(-xDiff / d)

	improvementX := math.Abs(dest.X - beforeX) - math.Abs(dest.X - v.X)
	if improvementX < 0 {
		panic(fmt.Sprintf("improvement of vehicle location (x coord) is negative"))
	}

	improvementY := math.Abs(dest.Y - beforeY) - math.Abs(dest.Y - v.Y)
	if improvementY < 0 {
		panic(fmt.Sprintf("improvement of vehicle location (y coord) is negative"))
	}

	if improvementX == 0 && improvementY == 0 {
		panic("0 improvement of vehicle position")
	}
}

func (v *VehicleActor) Ping(ts types.Timestamp) {
	if v.HasFinished {
		return
	}

	dist := v.Speed_mps * STEP_INTERVAL_MS / 1000

	v.move(dist)
}


func (v *VehicleActor) planRoute() {

	var nextNode, nodeFrom, nodeTo *types.Node
	for i, n := range v.roadGraph.AllNodes {
		switch n.Id {
		case v.Origin.NodeFirst:
			nextNode = &v.roadGraph.AllNodes[i]
		case v.Origin.NodeSecond:
			nodeFrom = &v.roadGraph.AllNodes[i]
		case v.Destination.NodeFirst:
			nodeTo = &v.roadGraph.AllNodes[i]
		}
	}
	if nextNode == nil || nodeFrom == nil || nodeTo == nil { panic("Not found") }

	path := findShortestPathFromTo(nodeFrom, nodeTo, v.roadGraph)
	path = append([]*types.Node{nextNode}, path...)
	v.route = path

}

