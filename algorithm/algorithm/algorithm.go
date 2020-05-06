package algorithm

import (
	"math"
	"mzakrze/smart_city/types"
)

const (
	VEHICLE_MASS = 1500
	VEHICLE_POWER = 150
	VEHICLE_BRAKING_DECELERATION = 3.5 // m/s^2

	VEHICLE_MAX_SPEED = 15
)

type AccelerateState = int
const (
	ACCELERATING AccelerateState = 1 + iota
	DECELERATING
	SUSTAIN
)


// FIXME - to jest skopiowane z simulationrunner, docelowo trzeba jakoś przepychać
const STEP_INTERVAL_MS = 100


// TODO - to docelowo oddzielny pakiet, żeby na pewno nie widziało simulationRunnera'a
// albo nawet lepiej - jeśli się da - oddzielna binarka

type VehicleActor struct {
	Origin      types.DestinationPoint
	Destination types.DestinationPoint

	// represents the center of the vehicle
	X         types.XCoord
	Y         types.YCoord
	Alpha     float64
	Speed_mps float64
	Width     types.Meter
	Length    types.Meter

	HasFinished bool
	AccelerateState AccelerateState

	roadGraph *types.Graph
	route     []RouteElem
}

type RouteElem struct {
	node *types.Node
	metersToDest types.Meter
	metersDone types.Meter
}

func InitVehicleActor(origin, destination types.DestinationPoint, roadGraph *types.Graph) *VehicleActor {
	res := &VehicleActor{
		Origin: origin,
		Destination: destination,
		X: origin.X,
		Y: origin.Y,
		Speed_mps: 0,
		Width: 1.981,
		Length: 4.636,
		AccelerateState: ACCELERATING,
		roadGraph: roadGraph,
	}

	res.planRoute()

	return res
	// TODO - wlaczanie sie do ruchu
}

func (v *VehicleActor) move(distSpare types.Meter) {

	dest := &v.route[0]

	xDiff := dest.node.X - v.X
	yDiff := dest.node.Y - v.Y
	d := math.Sqrt(xDiff * xDiff + yDiff * yDiff)

	if d <= distSpare {
		distSpare -= d

		dest.metersDone += d
		v.X = dest.node.X
		v.Y = dest.node.Y

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

	dest.metersDone = dest.metersDone + distSpare
	v.X += moveX
	v.Y += moveY
	v.Alpha = math.Atan(-moveY / moveX)

	if math.IsNaN(v.Alpha) {
		panic("Alpha is NaN")
	}
}

func (v *VehicleActor) Ping(ts types.Timestamp) {
	if v.HasFinished {
		return
	}

	switch v.AccelerateState {
	case SUSTAIN:
	case DECELERATING:
		delta_v := VEHICLE_BRAKING_DECELERATION * STEP_INTERVAL_MS / 1000
		v.Speed_mps -= delta_v
		if v.Speed_mps < 0 {
			v.Speed_mps = 0
		}
	case ACCELERATING:
		var delta_w float64 = VEHICLE_POWER * STEP_INTERVAL_MS / 1000
		delta_v := math.Sqrt(2 * delta_w / VEHICLE_MASS)
		v.Speed_mps += delta_v

		if v.Speed_mps >= VEHICLE_MAX_SPEED {
			v.AccelerateState = SUSTAIN
		}
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
		case v.Destination.NodeSecond:
			nodeTo = &v.roadGraph.AllNodes[i]
		}
	}
	if nextNode == nil || nodeFrom == nil || nodeTo == nil {
		panic("Not found")
	}

	path := findShortestPathFromTo(nodeFrom, nodeTo, v.roadGraph)
	route := make([]RouteElem, len(path))
	dist := 0.0
	var prevNode *types.Node = nil
	for i := len(path)-1; i >= 0; i-- {
		route[i].node = path[i]

		if prevNode == nil {
			route[i].metersToDest = dist
		} else {
			// find edge by id
			edgeLength := -1.0
			for _, e := range path[i].Edges {
				if e.To.Id == prevNode.Id {
					edgeLength = e.Distance
				}
			}
			if edgeLength < 0 {
				panic("Edge not found")
			}
			dist += edgeLength
			route[i].metersToDest = dist
		}
		prevNode = path[i]
	}

	v.route = route
}

