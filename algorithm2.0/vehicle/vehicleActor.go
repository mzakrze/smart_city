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
	beforeIntersectionNotAllowed = 1 + iota
	beforeIntersectionHasReservation
	atIntersection
	afterIntersection
)

const maxAcc = 2.0
const maxDecel = 3.5
const maxSpeed = 10.0

// TODO - docelowo wszystko prywatne - tylko jakiś Getter do stanu
type VehicleActor struct {
	Id          types.VehicleId
	X           types.XCoord
	Y           types.YCoord
	EdgeAt      *util.Edge
	Alpha       types.Angle
	Speed       types.MetersPerSecond
	Acc         types.MetersPerSecond2
	HasFinished bool
	State          vehicleState

	entryPoint     *util.Node
	exitPoint      *util.Node
	roadGraph      *util.Graph
	route          []*util.Edge
	sensor         *SensorLayer
	networkCard    *CommunicationLayer
	reservation    *reservation
	AlphaInitiated bool
}

type reservation struct {
	arriveConflictZoneTs types.Millisecond
	arriveConflictZoneSpeed types.MetersPerSecond
	leaveConflictZoneTs types.Millisecond
	locationPerTime map[types.Millisecond]types.Location
}

func NewVehicleActor(id types.VehicleId, ts types.Millisecond, entrypoint, exitpoint *util.Node, initSpeed types.MetersPerSecond, roadGraph *util.Graph, sensor *SensorLayer, comm *CommunicationLayer) *VehicleActor {
	v := VehicleActor{
		Id:             id,
		X:              entrypoint.X,
		Y:              entrypoint.Y,
		entryPoint:     entrypoint,
		exitPoint:      exitpoint,
		EdgeAt:         entrypoint.EdgesFrom[0],
		Speed:          initSpeed,
		Alpha:          0.0,
		AlphaInitiated: false,
		Acc:            0.0,
		HasFinished:    false,
		roadGraph:      roadGraph,
		sensor:         sensor,
		State:          beforeIntersectionNotAllowed,
		networkCard:    comm,
	}

	v.planRoute()

	return &v
}

func (v *VehicleActor) Ping(ts types.Millisecond) {
	v.handleMessages()

	v.Acc = v.calcAcceleration(ts)
	v.Speed += v.Acc * float64(constants.SimulationStepInterval) / 1000.0
	if v.Speed > maxSpeed {
		//panic("xd")
	}
	if v.Speed < 0.0 {
		panic("xd")
	}

	dist := v.Speed * float64(constants.SimulationStepInterval) / 1000.0
	v.move(dist)

	v.sendMessages(ts)

	v.updateState()
}

func (v *VehicleActor) calcAcceleration(ts types.Millisecond) float64 {
	moderate := func (decel float64) float64 {
		vdiff := decel * float64(constants.SimulationStepInterval) / 1000.0
		if v.Speed + vdiff < 0.0 {
			//decelModerated := brakingDecel(v.Speed, constants.SimulationStepInterval)
			//if math.Abs(decel) < math.Abs(decelModerated) {
			//	panic("oops")
			//}
			//return decelModerated
			// FIXME:
			v.Speed = 0.0 // hack (powyższe nie dziala (blad przy operacjach na liczbach zmiennoprzecinkowych, dlatego narazie hack)
			return 0.0
		}
		return decel
	}
	switch v.State {
	case beforeIntersectionNotAllowed:
		d1 := v.calculateDistanceToConflictZone()
		d2 := v.sensor.ScanVehiclesAhead(v)
		d := math.Min(d1, d2)

		if d < 0.0 {
			panic("Oops - vehicle crashed")
		}

		const maxComfortableDecel = 1.0
		distToStartBraking := v.Speed * v.Speed / (2.0 * maxComfortableDecel)
		distToStartEmergencyBreaking := v.Speed * v.Speed / (2.0 * maxDecel) + 1.0
		if d <= distToStartEmergencyBreaking {
			return moderate(-maxDecel)
		}
		if d <= distToStartBraking {
			return moderate(-maxComfortableDecel)
		}
		if d > 1.0 && v.Speed <= maxSpeed {
			return 1.0
		}
		return 0.0
	case beforeIntersectionHasReservation:
		if v.reservation == nil {
			panic("oops")
		}
		desiredSpeed := v.reservation.arriveConflictZoneSpeed
		timeLeft := v.reservation.arriveConflictZoneTs - ts
		d := v.calculateDistanceToConflictZone()
		if v.Speed == desiredSpeed {
			if d0 := float64(v.Speed) * float64(timeLeft); d != d0  {
				panic(fmt.Sprintf("Vehicle missed it reservation: shoudl be %f, is: %f", d, d0))
			}
		}
		deltaV := 2.0 / float64(constants.SimulationStepInterval)
		if v.Speed + deltaV <= desiredSpeed  {
			return 2.0
		} else if v.Speed < desiredSpeed {
			diff := desiredSpeed - v.Speed
			acc := diff / float64(constants.SimulationStepInterval)
			if acc > 2.0 {
				panic("oops")
			}
			return acc
		} else {
			panic("oops")
		}
	case atIntersection:
		if v.Speed < 10 {
			diff := 10.0 - v.Speed // [m/s]
			acc := diff * 1000.0 / 10
			return acc
		} else if v.Speed == 10 {
			return 0.0
		} else {
			//panic("to much speed")
			return 0.0
		}
	case afterIntersection:
		maxAcc := 2.0
		deltaV := maxAcc * float64(constants.SimulationStepInterval)
		if v.Speed > maxSpeed {
			//panic("oops")
			return 0.0
		} else if v.Speed == 0 {
			return 0.0
		} else {
			if v.Speed + deltaV <= 10.0 {
				return maxAcc
			} else {
				diff := 10.0 - v.Speed
				acc := diff / float64(constants.SimulationStepInterval)
				if acc > 2.0 { panic("oops") }
				return acc
			}
		}
	default:
		panic("Illegal State")
	}
}

func (v *VehicleActor) move(distSpare types.Meter) {
	if distSpare == 0.0 {
		return
	}

	if len(v.route) == 0 {
		v.HasFinished = true
		v.X -= math.Cos(math.Pi - v.Alpha) * distSpare
		v.Y -= math.Cos(math.Pi / 2 - v.Alpha) * distSpare
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
	switch {
	case moveX == 0 && moveY > 0: // up
		v.Alpha = -math.Pi / 2
	case moveX == 0 && moveY < 0: // down
		v.Alpha = math.Pi / 2
	case moveY == 0 && moveX > 0: // right
		v.Alpha = 0
	case moveY == 0 && moveX < 0: // left
		v.Alpha = math.Pi
	default:
		v.Alpha = math.Atan(-moveY / moveX)
	}
	v.AlphaInitiated = true
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
	isFirst := d >= MaxDistanceMeasurement
	if isFirst == false {
		return
	}

	toConflictZone := v.calculateDistanceToConflictZone()
	if toConflictZone < 0 {
		panic("oops")
	}
	tdiff := types.Millisecond(1000 * toConflictZone / 10)

	approachConflictZoneTs := ts+tdiff

	conflictZoneLength := 0.0
	for _, r := range v.route {
		if r.IsArc {
			conflictZoneLength += r.Length
		}
	}
	conflictZoneLength += constants.VehicleLength
	duration := conflictZoneLength / 10.0 * 1000.0
	leaveConflictZoneTs := ts + tdiff + types.Millisecond(duration)

	t := (math.Sqrt(v.Speed * v.Speed + 2 * maxAcc * toConflictZone) - v.Speed) / 2.0
	deltaV := maxAcc * t
	approachConflictZoneSpeedMax := v.Speed + deltaV
	approachConflictZoneSpeedMax = math.Min(approachConflictZoneSpeedMax, 10.0)

	if math.IsNaN(approachConflictZoneSpeedMax) {
		fmt.Println("oops")
	}

	t = (math.Sqrt(v.Speed * v.Speed + 2 * maxDecel * toConflictZone) - v.Speed) / 2.0
	deltaV = maxDecel * t
	approachConflictZoneSpeedMin := v.Speed - deltaV

	m := DsrcV2RMessage{
		MsgType:               AimProtocolMsgRequest,
		Sender:                v.Id,
		TsSent:                ts,
		ApproachConflictZoneTs: approachConflictZoneTs,
		LeaveConflictZoneTs: 	leaveConflictZoneTs,
		ApproachConflictZoneSpeedMax: approachConflictZoneSpeedMax,
		ApproachConflictZoneSpeedMin: approachConflictZoneSpeedMin,
		X:                     v.X,
		Y:                     v.Y,
		Speed:                 v.Speed,
		Acc:                   v.Acc,
	}

	v.networkCard.SendDsrcV2R(m)
}

func (v *VehicleActor) handleMessages() {
	messages := v.networkCard.VehicleReceive(v.Id)
	for _, m := range messages {
		if m.msgType == AimProtocolMsgAllow {
			if v.State == beforeIntersectionNotAllowed {
				v.State = beforeIntersectionHasReservation
			}
			reservation := reservation{
				arriveConflictZoneTs: m.reservationFromTs,
				arriveConflictZoneSpeed: m.reservationDesiredSpeed,
				leaveConflictZoneTs: m.reservationToTs,
				// TODO - locationPerTime
				//:locationPerTime map[types.Millisecond]types.Location
			}
			v.reservation = &reservation
		}
	}
}

func (v *VehicleActor) calculateDistanceToConflictZone() float64 {
	c := &v.roadGraph.ConflictZone

	var r float64
	if c.MinX < v.X && v.X < c.MaxX {
		r = math.Min(math.Abs(v.Y - c.MaxY), math.Abs(v.Y - c.MinY))
	} else if c.MinY < v.Y && v.Y < c.MaxY {
		r = math.Min(math.Abs(v.X - c.MaxX), math.Abs(v.X - c.MinX))
	} else {
		panic("Illegal location of vehicle")
	}

	// what interests us is the whole not it's center point, hence minus half of length
	return r - constants.VehicleLength / 2
}

func (v *VehicleActor) sendMessages(ts types.Millisecond) {
	switch v.State {
	case beforeIntersectionNotAllowed:
		if v.AlphaInitiated == false {
			panic("Sending request before initiating alpha")
		}
		v.sendRequestPermission(ts)
	case beforeIntersectionHasReservation:
		// nothing
	case atIntersection:
		// nothing
	case afterIntersection:
		// nothing
	}
}

func (v *VehicleActor) updateState() {
	switch v.State {
	case beforeIntersectionNotAllowed:
		if v.isInConflictZone() {
			panic("Entered conflict zone without reservation")
		}
	case beforeIntersectionHasReservation:
		if v.isInConflictZone() {
			v.State = atIntersection
		}
	case atIntersection:
		if v.isInConflictZone() == false {
			v.State = afterIntersection
		}
	}
}
