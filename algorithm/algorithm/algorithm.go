package algorithm

import (
	"fmt"
	"math"
	"mzakrze/smart_city/types"
)

type IntersectionState = int
const (
	BEFORE = 1 + iota
	DURING
	AFTER
)


// FIXME - to jest skopiowane z simulationrunner, docelowo trzeba jakoś przepychać
const STEP_INTERVAL_MS = 100


// TODO - to docelowo oddzielny pakiet, żeby na pewno nie widziało simulationRunnera'a
// albo nawet lepiej - jeśli się da - oddzielna binarka

type VehicleActor struct {
	Id          types.VehicleId
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

	roadGraph                   *types.Graph
	Route                       []RouteElem
	accessToIntersectionGranted bool
	intersectionState           IntersectionState
	breakingPidController       *PIDController
	aheadSensor                 *VehicleAheadSensor
}

type RouteElem struct {
	Node *types.Node
	MetersToDest types.Meter
	MetersDone types.Meter
}

func InitVehicleActor(id types.VehicleId, origin, destination types.DestinationPoint, ts types.Milisecond, roadGraph *types.Graph, sensor *VehicleAheadSensor) *VehicleActor {
	res := &VehicleActor{
		Id: id,
		Origin: origin,
		Destination: destination,
		X: origin.X,
		Y: origin.Y,
		Speed_mps: 10,
		Width: 1.981,
		Length: 4.636,
		roadGraph: roadGraph,

		accessToIntersectionGranted: false,
		intersectionState: BEFORE,
		aheadSensor: sensor,
	}

	const desiredDistanceToConflictZone = 0.0

	res.breakingPidController = NewApproachConflictZonePidController()
	res.breakingPidController.SetOutputLimits(-2.0, 3.5)
	res.breakingPidController.Set(desiredDistanceToConflictZone, ts)

	res.planRoute()

	return res
}

func (v *VehicleActor) move(distSpare types.Meter) {

	dest := &v.Route[0]

	xDiff := dest.Node.X - v.X
	yDiff := dest.Node.Y - v.Y
	d := math.Sqrt(xDiff * xDiff + yDiff * yDiff)

	if d <= distSpare {
		distSpare -= d

		dest.MetersDone += d
		v.X = dest.Node.X
		v.Y = dest.Node.Y

		if len(v.Route) == 1 {
			v.HasFinished = true
			return
		}

		v.Route = v.Route[1:]
		if distSpare > 0 {
			v.move(distSpare)
		}

		return
	}

	moveX := distSpare * xDiff / d
	moveY := distSpare * yDiff / d

	dest.MetersDone = dest.MetersDone + distSpare
	v.X += moveX
	v.Y += moveY
	v.Alpha = math.Atan(-moveY / moveX)
	if v.Alpha == -0 {
		v.Alpha = -math.Pi
	}
	if math.IsNaN(v.Alpha) {
		panic("Alpha is NaN")
	}
}


func (v *VehicleActor) calculateDistanceToConflictZone() float64 {
	// FIXME - wykrywać pojazd przed
	// (to będzie działać - jak samochod przed odjedzie - pid ogarnie i przyspieszy)
	im := &v.roadGraph.IntersectionManager
	x := math.Min(math.Abs(v.X - im.BboxLeft), math.Abs(v.X - im.BboxRight))
	y := math.Min(math.Abs(v.Y - im.BboxDown), math.Abs(v.Y - im.BboxUp))

	// what interests us is the whole not it's center point, hence minus half of length
	r := math.Sqrt(x * x + y * y) - v.Length / 2
	//fmt.Printf("dist: %f", r)
	return r
}

func (v *VehicleActor) isInConflictZone() bool {
	im := &v.roadGraph.IntersectionManager
	return im.BboxDown <= v.Y && v.Y <= im.BboxUp && im.BboxLeft <= v.X && v.X <= im.BboxRight
}

func (v *VehicleActor) Ping(ts types.Milisecond) {
	if v.HasFinished {
		return
	}

	wasInConflictZone := v.isInConflictZone()

	if v.intersectionState == BEFORE && v.accessToIntersectionGranted == false {
		d1 := v.calculateDistanceToConflictZone()
		d2 := v.aheadSensor.ScanVehiclesAhead(v) - 5
		d := math.Min(d1, d2)

		fmt.Printf("Distance to intersection: %f, distance other car: %f\n", d1, d2)

		// can be negative
		acceleration := v.breakingPidController.Feedback(d, ts)
		//fmt.Printf(", acc: %f\n", acceleration)
		v.Speed_mps += acceleration * STEP_INTERVAL_MS / 1000

		dist := v.Speed_mps * STEP_INTERVAL_MS / 1000

		//if dist - 0.5 > d { // TODO 0.2 tymczasowo - jest błąd, ale niewielki (odstęp pożądany między pojazdami to 2 metry, więc i tak sie nie zderzą)
		//	panic("Vehicle crashed into an obstacle :(")
		//}

		v.move(dist)

	} else if v.intersectionState == BEFORE && v.accessToIntersectionGranted == true {
		if v.Speed_mps < 10 {
			// acceleration = 2 m/s^2
			v.Speed_mps += 2 * STEP_INTERVAL_MS / 1000
		}
	}



	if v.isInConflictZone() {
		//if v.accessToIntersectionGranted == false {
		//	panic("Vehicle entered into intersection without permission")
		//}
		v.intersectionState = DURING
	} else {
		if wasInConflictZone {
			v.intersectionState = AFTER
		}
	}


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
		route[i].Node = path[i]

		if prevNode == nil {
			route[i].MetersToDest = dist
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
			route[i].MetersToDest = dist
		}
		prevNode = path[i]
	}

	v.Route = route
}

