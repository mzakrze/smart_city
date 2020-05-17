package vehicle

import (
	"algorithm2.0/constants"
	"algorithm2.0/types"
	"algorithm2.0/util"
	"fmt"
	"math"
)

type vehicleState = int
const (
	beforeIntersectionNoPermission = 1 + iota
	beforeIntersectionHasPermission
	atIntersection
	afterIntersection
)

type VehicleActor struct {
	Id          types.VehicleId
	X           types.XCoord
	Y           types.YCoord
	EdgeAt      *util.Edge
	Alpha       types.Angle
	Speed       types.MetersPerSecond
	Acc         types.MetersPerSecond2
	HasFinished bool

	entryPoint  *util.Node
	exitPoint   *util.Node
	roadGraph *util.Graph
	route     []*util.Edge
	//breakingPidController *PIDController TODO
	sensor      *SensorLayer
	state       vehicleState
	networkCard *CommunicationLayer
}

func NewVehicleActor(id types.VehicleId, entrypoint, exitpoint *util.Node, initSpeed types.MetersPerSecond, roadGraph *util.Graph, sensor *SensorLayer, comm *CommunicationLayer) *VehicleActor {
	v := VehicleActor{
		Id: id,
		X: entrypoint.X,
		Y: entrypoint.Y,
		entryPoint: entrypoint,
		exitPoint: exitpoint,
		EdgeAt: entrypoint.EdgesFrom[0],
		Speed: 10.0,
		Alpha: 0.0,
		Acc: 0.0,
		HasFinished: false,
		roadGraph: roadGraph,
		sensor: sensor,
		state: beforeIntersectionNoPermission,
		networkCard: comm,
	}

	v.planRoute()

	return &v
}

func (v *VehicleActor) Ping(ts types.Millisecond) {
	v.handleMessages()

	switch v.state {
	case beforeIntersectionNoPermission:
		v.sendRequestPermission(ts)
	case beforeIntersectionHasPermission:
		panic("not implemented")
	case atIntersection:
		panic("not implemented")
	case afterIntersection:
		panic("not implemented")
	}

	//if math.IsNaN(acc) {
	//	fmt.Println("oops")
	//}
	//if math.IsInf(acc, 1) {
	//	fmt.Println("oops")
	//}
	//if math.IsInf(acc, -1) {
	//	fmt.Println("oops")
	//}

	v.Acc = v.calcAcceleration(ts)
	v.Speed += v.Acc * float64(constants.SimulationStepInterval) / 1000.0

	//if v.Speed == 0 {
	//	fmt.Println("0 speed")
	//}

	if math.IsInf(v.Speed, -1) {
		fmt.Println("oops")
	}

	dist := v.Speed * float64(constants.SimulationStepInterval) / 1000.0
	v.move(dist)

	switch v.state {
	case beforeIntersectionHasPermission:
		if v.isInConflictZone() {
			v.state = atIntersection
		}
	case atIntersection:
		if v.isInConflictZone() == false {
			v.state = afterIntersection
		}
	}
}

func (v *VehicleActor) calcAcceleration(ts types.Millisecond) float64 {
	switch v.state {
	case beforeIntersectionNoPermission:
		d1 := v.calculateDistanceToConflictZone()
		d2 := v.sensor.ScanVehiclesAhead(v)
		fmt.Println(d2)
		d := math.Min(d1, d2)
		//d -= constants.VehicleLength / 2
		// TODO - docelowo - użyj pid controllera żeby dostosować, tymczasoso - wyhamuj do 0
		//if d < 0.0 {
		//	panic("xd")
		//}
		if d == 0 {
			return 0.0
		}
		//if 0.0 < d && d < 1.0 {
		//	return 0.0
		//}
		if d <= 1.0 {
			// dopasowujemy tak, żeby w tym kroku wyhamował
			t := float64(constants.SimulationStepInterval) / 1000.0 // [s]
			acc := -1 * v.Speed / t
			//acc := -math.Sqrt(2.0 * d / (t * t))

			if math.IsInf(acc, -1) {
				fmt.Println("oops")
			}
			return acc
		} else {
			return 0.0
		}
	case beforeIntersectionHasPermission:
		// TODO - docelowo - użyj pid controllera, tymczasoswo
		fallthrough
	case atIntersection:
		panic("not implemented")
		// TODO - docelowo - użyj pid controllera, tymczasoswo
		if v.Speed < 10 {
			diff := 10.0 - v.Speed
			acc := diff * 1000.0
			return acc
		} else if v.Speed == 10 {
			// nothing
		} else {
			panic("to much speed")
		}
	default:
		panic("Illegal state")
	}

	panic("illegal state")
}

func (v *VehicleActor) move(distSpare types.Meter) {

	if distSpare == 0.0 {
		return
	}

	if len(v.route) == 0 {
		switch v.Alpha {
		case math.Pi / 2: // riding down
			v.Y -= distSpare
		case -math.Pi / 2: // riding up
			v.Y += distSpare
		case 0: // riding right
			v.X += distSpare
		case -math.Pi: // riding left
			v.X -= distSpare
		default:
			panic("vehicle cant know where to ride")
		}

		return
	}

	dest := v.route[0]
	v.EdgeAt = dest

	xDiff := dest.To.X - v.X
	yDiff := dest.To.Y - v.Y
	d := math.Sqrt(xDiff * xDiff + yDiff * yDiff)

	if d == 0.0 && len(v.route) == 1 {
		v.route = []*util.Edge{}
		v.move(distSpare)
	}

	if d <= distSpare {
		distSpare -= d

		v.X = dest.To.X
		v.Y = dest.To.Y

		if len(v.route) == 1 {
			v.HasFinished = true
			return
		}

		if len(v.route) == 0 {
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
	v.Alpha = math.Atan(-moveY / moveX)
	if v.Alpha == -0 {
		v.Alpha = -math.Pi
	}
	if math.IsNaN(v.Alpha) {
		panic("Alpha is NaN")
	}
}

func (v *VehicleActor) isInConflictZone() bool {
	c := &v.roadGraph.ConflictZone
	return c.MinY <= v.Y && v.Y <= c.MaxY && c.MinX <= v.X && v.X <= c.MaxX
}

func (v *VehicleActor) planRoute() {
	var visit func (node *util.Node) bool
	path := []*util.Edge{}

	visit = func (node *util.Node) bool {
		if node.Id == v.exitPoint.Id {
			return true
		}
		for i := range node.EdgesFrom {
			path = append(path, node.EdgesFrom[i])
			res := visit(node.EdgesFrom[i].To)
			if res {
				return true
			}
			path = path[:len(path) - 1]
		}

		return false

	}

	res := visit(v.entryPoint)

	if res == false || len(path) == 0 {
		panic("Vehicle route not found")
	}

	v.route = path
}

func (v *VehicleActor) sendRequestPermission(ts types.Millisecond) {
	d := v.sensor.ScanVehiclesAhead(v)
	isFirst := d >= MaxDistanceMeasurment

	m := DsrcV2RMessage{
		MsgType:               AimProtocolMsgRequest,
		Sender:                v.Id,
		TsSent:                ts,
		IsFirstOnIntersection: isFirst,
		VehicleInFrontId:      0, // TODO
		X:                     v.X,
		Y:                     v.Y,
		Speed:                 v.Speed,
		Acc:                   v.Acc,
	}

	v.networkCard.SendDsrcV2R(m)
}

func (v *VehicleActor) handleMessages() {
	v.networkCard.VehicleReceive(v.Id)
}

func (v *VehicleActor) calculateDistanceToConflictZone() float64 {
	c := &v.roadGraph.ConflictZone

	var r float64
	if c.MinX < v.X && v.X < c.MaxX {
		r = math.Min(math.Abs(v.Y - c.MaxY), math.Abs(v.Y - c.MinY))
	} else if c.MinY < v.Y && v.Y < c.MaxY {
		r = math.Min(math.Abs(v.X - c.MaxX), math.Abs(v.X - c.MinX))
	} else {
		panic("xD")
	}

	// what interests us is the whole not it's center point, hence minus half of length
	return r - constants.VehicleLength / 2
}