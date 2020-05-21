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
const maxAngularSpeed = 5.0

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
	platooningReservation *reservation
	reservationBroadcasted bool
	approachConflictZonePlan map[types.Millisecond]types.MetersPerSecond
	AlphaInitiated bool
}

type reservation struct {
	reservationId		types.ReservationId
	arriveConflictZoneTs types.Millisecond
	arriveConflictZoneSpeed types.MetersPerSecond
	leaveConflictZoneTs types.Millisecond
	speedPerTime map[types.Millisecond]types.MetersPerSecond
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
	v.handleMessages(ts)

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

	v.updateState(ts)
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
		d1 := v.calculateDistanceCenterToConflictZone() - constants.VehicleLength / 2
		d2, _ := v.sensor.ScanVehiclesAhead(v)
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

		const maxDecelDiff = maxDecel * float64(constants.SimulationStepInterval) / 1000.0
		const maxAccDiff = maxAcc * float64(constants.SimulationStepInterval) / 1000.0

		desiredSpeed, exists := v.approachConflictZonePlan[ts]
		if exists == false {
			fmt.Println("Key: ", ts, "doesnt exist")
			fmt.Println(v.approachConflictZonePlan)
			//panic("xd")
			desiredSpeed = v.Speed
		}
		if v.Speed + maxAccDiff < desiredSpeed {
			panic("xd")
		}
		if v.Speed - maxDecelDiff > desiredSpeed {
			panic("xd")
		}
		v.Speed = desiredSpeed
		return 0.0 // TODO - hack
	case atIntersection:
		if v.reservation == nil { panic("Oops") }
		desiredSpeed, exists := v.reservation.speedPerTime[ts]
		if exists == false {
			// something is wrong, but we can recover
			// vehicle is late - hence has no entry in table for that timestamp
			desiredSpeed = v.Speed
		}
		if v.Speed == desiredSpeed {
			return 0.0
		}

		if desiredSpeed > v.Speed {
			maxSpeedIncrease := types.MetersPerSecond(maxAcc * float64(constants.SimulationStepInterval)) / 1000.0
			if v.Speed + maxSpeedIncrease < desiredSpeed {
				//panic("Vehicle couldn't catch up with reservation")
			}
			v.Speed = desiredSpeed // TODO hack
			return 0.0
		} else {
			maxSpeedDecrease := types.MetersPerSecond(maxDecel * float64(constants.SimulationStepInterval)) / 1000.0
			if v.Speed - maxSpeedDecrease > desiredSpeed {
				panic("Vehicle couldn't catch up with reservation")
			}
			v.Speed = desiredSpeed // TODO hack
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

func (v *VehicleActor) isCenterInConflictZone() bool {
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
	d, _ := v.sensor.ScanVehiclesAhead(v)
	isFirst := d >= MaxDistanceMeasurement
	if isFirst == false {
		if v.platooningReservation != nil {
			v.sendRequestPermissionPlatooning(ts)
		}
		return
	}

	toConflictZone := v.calculateDistanceCenterToConflictZone()
	if toConflictZone < 0 {
		panic("oops")
	}


	var tdiff types.Millisecond
	var maxSpeedOnCurve types.MetersPerSecond = math.MaxFloat64
	var approachConflictZoneSpeedMax types.MetersPerSecond

	t := (math.Sqrt(v.Speed * v.Speed + 2 * maxAcc * toConflictZone) - v.Speed) / 2.0
	deltaV := maxAcc * t
	approachConflictZoneSpeedMax = v.Speed + deltaV

	if v.isTurning() {
		radius := math.Abs(v.entryPoint.X - v.exitPoint.X)
		maxSpeedOnCurve = 2.0 * math.Pi * math.Sqrt(maxAngularSpeed * radius)

		t1, t2, t3 := arrivalTimeAcceleratingEnterWithSpeed(v.Speed, maxSpeed, maxAcc, maxDecel, toConflictZone, maxSpeedOnCurve)
		tdiff = t1 + t2 + t3
	} else {
		tdiff = arrivalTimeAccelerating(v.Speed, maxSpeed, maxAcc, toConflictZone)
	}

	approachConflictZoneSpeedMax = math.Min(maxSpeedOnCurve, math.Min(approachConflictZoneSpeedMax, 10.0))
	approachConflictZoneMinTs := ts+tdiff



	if math.IsNaN(approachConflictZoneSpeedMax) {
		fmt.Println("oops")
	}

	conflictZoneNodeEnter, conflictZoneNodeExit := v.conflictZoneNodeEnterExit()

	m := DsrcV2RMessage{
		MsgType:                      AimProtocolMsgRequest,
		Sender:                       v.Id,
		TsSent:                       ts,
		ApproachConflictZoneMinTs:    approachConflictZoneMinTs,
		ApproachConflictZoneSpeedMax: approachConflictZoneSpeedMax,
		MaxSpeedOnCurve:			  maxSpeedOnCurve,
		IsTurning:					  v.isTurning(),
		EntryPointId:                 v.entryPoint.Id,
		ExitPointId:                  v.exitPoint.Id,
		ConflictZoneNodeEnter:		  conflictZoneNodeEnter,
		ConflictZoneNodeExit:		  conflictZoneNodeExit,
		VehicleX:                     v.X,
		VehicleY:                     v.Y,
		VehicleSpeed:                 v.Speed,
	}

	v.networkCard.SendDsrcV2R(m)
}

func (v *VehicleActor) handleMessages(ts types.Millisecond) {
	messages := v.networkCard.VehicleReceive(v.Id)
	for _, m := range messages {
		switch m.msgType {

		case AimProtocolMsgAllow:
			if v.State != beforeIntersectionNotAllowed {
				continue
			}
			d := v.calculateDistanceCenterToConflictZone()

			canArriveAt := ts + arrivalTimeAccelerating(v.Speed, maxSpeed, maxAcc, d)

			if canArriveAt > m.reservationFromTs {
				reply := DsrcV2RMessage{
					MsgType:               AimProtocolMsgReservationCancelation,
					ReservationToCancelId: m.reservationId,
				}
				// TODO - troche hack bo w innym momencie powinny byc wysyłane requesty
				v.networkCard.SendDsrcV2R(reply)
				continue
			}

			reservation := reservation{
				reservationId: m.reservationId,
				arriveConflictZoneTs: m.reservationFromTs,
				arriveConflictZoneSpeed: m.reservationDesiredSpeed,
				leaveConflictZoneTs: m.reservationToTs,
				speedPerTime: m.reservationTsToSpeed,
			}
			v.reservation = &reservation
			v.planApproachToConflictZone(ts)
			v.State = beforeIntersectionHasReservation
		}
	}

	for _, m := range v.networkCard.VehicleReceiveV2V(v.Id) {
		switch m.msgType {
		case AimProtocolMsgReservationInfo:
			_, vId := v.sensor.ScanVehiclesAhead(v)
			if vId != m.sender {
				continue
			}

			v.platooningReservation = &reservation{
				reservationId: m.reservationId,
				arriveConflictZoneTs: m.reservationFromTs,
				arriveConflictZoneSpeed: m.reservationTsToSpeed[m.reservationFromTs],
				speedPerTime: m.reservationTsToSpeed,
			}

		}
	}
}

func (v *VehicleActor) calculateDistanceCenterToConflictZone() float64 {
	c := &v.roadGraph.ConflictZone

	var r float64
	if c.MinX < v.X && v.X < c.MaxX {
		r = math.Min(math.Abs(v.Y - c.MaxY), math.Abs(v.Y - c.MinY))
	} else if c.MinY < v.Y && v.Y < c.MaxY {
		r = math.Min(math.Abs(v.X - c.MaxX), math.Abs(v.X - c.MinX))
	} else {
		panic("Illegal location of vehicle")
	}

	return r
}

func (v *VehicleActor) sendMessages(ts types.Millisecond) {
	switch v.State {
	case beforeIntersectionNotAllowed:
		if v.AlphaInitiated == false {
			panic("Sending request before initiating alpha")
		}
		v.sendRequestPermission(ts)
	case beforeIntersectionHasReservation:
		if v.reservationBroadcasted == false {
			v.sendReservationInfoBroadcast()
			v.reservationBroadcasted = true
		}
	case atIntersection:
		// nothing
	case afterIntersection:
		// nothing
	}
}

func (v *VehicleActor) updateState(ts types.Millisecond) {
	switch v.State {
	case beforeIntersectionNotAllowed:
		if v.isCenterInConflictZone() {
			panic("Entered conflict zone without reservation")
		}
	case beforeIntersectionHasReservation:
		if v.isCenterInConflictZone() {
			v.State = atIntersection

			if math.Abs(float64(v.reservation.arriveConflictZoneTs - ts)) > 0 {
				//fmt.Println("Vehicle should arrive at:", v.reservation.arriveConflictZoneTs, "but arrived:", ts, "difference =", (ts - v.reservation.arriveConflictZoneTs))
				fmt.Println("Vehicle", v.Id, "should arrive at: ", v.reservation.arriveConflictZoneTs, ", but arrived:", ts)
				//panic("Vehicle missed its reservation")
			}

			if math.Abs(v.reservation.arriveConflictZoneSpeed - v.Speed) > 0 {
				fmt.Println("Vehicle", v.Id, "should arrive with v=", v.reservation.arriveConflictZoneSpeed, ", but arrived with v=", v.Speed, ", diff = ", (v.reservation.arriveConflictZoneSpeed - v.Speed))
			}
		}
	case atIntersection:
		if v.isCenterInConflictZone() == false {
			v.State = afterIntersection

			if math.Abs(float64(v.reservation.leaveConflictZoneTs - ts)) > 0 {
				fmt.Println("Vehicle", v.Id, "should leave at:", v.reservation.leaveConflictZoneTs, "but left:", ts, "difference =", (ts - v.reservation.leaveConflictZoneTs), "Is turning?", v.isTurning())
				//panic("Vehicle missed its reservation")
			}

			if v.platooningReservation != nil {
				fmt.Println("Vehicle With Platooning reservation completed successfully")
			}
		}
	}
}

func (v *VehicleActor) isTurning() bool {
	diff := v.entryPoint.WayId - v.exitPoint.WayId

	if diff == -1 || diff == 3 {
		return true
	} else if diff == 1 || diff == -3 {
		return true
	} else if diff == -2 || diff == 2 {
		return false
	} else {
		panic("Illegal way")
	}
}

func (v *VehicleActor) sendReservationInfoBroadcast() {
	msg := DsrcV2VMessage{
		msgType: AimProtocolMsgReservationInfo,
		sender: v.Id,
		reservationId: v.reservation.reservationId,
		reservationFromTs: v.reservation.arriveConflictZoneTs,
		reservationTsToSpeed: v.reservation.speedPerTime,
	}

	v.networkCard.SendDsrcV2V(msg)
}

func (v *VehicleActor) conflictZoneNodeEnterExit() (*util.Node, *util.Node) {
	var conflictZoneNodeEnter, conflictZoneNodeExit *util.Node
	conflictZoneNodeEnter = v.route[0].To
	if v.isTurning() {
		for i := 1; i < len(v.route); i += 1 {
			if v.route[i].IsArc == false {
				break
			}
			conflictZoneNodeExit = v.route[i].To
		}
		if conflictZoneNodeExit == nil {
			panic("Oops")
		}
	} else {
		if len(v.route) != 3 {
			panic("Oops")
		}
		conflictZoneNodeExit = v.route[1].To
	}

	return conflictZoneNodeEnter, conflictZoneNodeExit
}

func (v *VehicleActor) sendRequestPermissionPlatooning(ts types.Millisecond) {
	r := v.platooningReservation
	dist, _ := v.sensor.ScanVehiclesAhead(v)
	//dist += constants.VehicleLength / 2

	if v.Speed == 0.0 {
		initSpeed := r.speedPerTime[r.arriveConflictZoneTs]
		tsDiff := types.Millisecond(1000.0 * dist / initSpeed)
		tsDiff = tsDiff - tsDiff % 10

		conflictZoneNodeEnter, conflictZoneNodeExit := v.conflictZoneNodeEnterExit()

		var maxSpeedOnCurve types.MetersPerSecond = math.MaxFloat64
		if v.isTurning() {
			// FIXME - skopiowany kod
			radius := math.Abs(v.entryPoint.X - v.exitPoint.X)
			maxSpeedOnCurve = 2.0 * math.Pi * math.Sqrt(maxAngularSpeed * radius)
		}

		req := DsrcV2RMessage{
			MsgType: AimProtocolMsgRequest,
			Sender: v.Id,
			PlatooningReservationId: r.reservationId,
			TsSent: ts,
			ApproachConflictZoneSpeedMax: initSpeed,
			ApproachConflictZoneMinTs: r.arriveConflictZoneTs + tsDiff,
			ConflictZoneNodeEnter: conflictZoneNodeEnter,
			ConflictZoneNodeExit: conflictZoneNodeExit,
			VehicleX: v.X,
			VehicleY: v.Y,
			VehicleSpeed: v.Speed,
			MaxSpeedOnCurve: maxSpeedOnCurve,
			IsTurning: v.isTurning(),
			EntryPointId: v.entryPoint.Id,
			ExitPointId: v.exitPoint.Id,
		}

		if v.Id == 22 {
			fmt.Println("xd")
		}

		v.networkCard.SendDsrcV2R(req)
	} else {

		// TODO - plz implement me

	}

}

func (v *VehicleActor) planApproachToConflictZone(currTs types.Millisecond) {
	// FIXME - po wyliczeniu broadcastować
	const comfortableDecel = 1.0

	v0 := v.Speed
	v2 := v.reservation.arriveConflictZoneSpeed
	t := v.reservation.arriveConflictZoneTs - currTs
	s := v.calculateDistanceCenterToConflictZone()

	var tryRange func(a1, a2 types.MetersPerSecond2) (types.MetersPerSecond2, types.Millisecond, types.Millisecond, types.Millisecond, bool)
	tryRange = func(a1, a2 types.MetersPerSecond2) (types.MetersPerSecond2, types.Millisecond, types.Millisecond, types.Millisecond, bool) {
		if a2 - a1 < 0.0000000000000001 {
			panic("Oops")
			// give up
			//return 0, 0, 0, 0, false
		}

		a := (a1 + a2) / 2

		t1, t2, t3 := arrivalTimeAcceleratingEnterWithSpeed(v0, maxSpeed, a, comfortableDecel, s, v2)
		tReal := t1 + t2 + t3

		if tReal == t {
			return a, t1, t2, t3, true
		}

		if tReal > t {
			// gotta go faster
			return tryRange(a, a2)
		} else {
			// gotta go slower
			return tryRange(a1, a)
		}
	}

	acc, accTime, constTime, decelTime, ok := tryRange(0.0, maxAcc)
	if ok == false {
		panic("Xd")
	}

	speed := v.Speed
	stopAcc := currTs + accTime + constants.SimulationStepInterval
	startDecel := currTs + accTime + constTime
	end := startDecel + decelTime

	if v.reservation.arriveConflictZoneTs != end {
		fmt.Println("aiming at: ", v.reservation.arriveConflictZoneTs, ", in reality is:", end)
	}

	v.approachConflictZonePlan = make(map[types.Millisecond]types.MetersPerSecond)
	for ts := currTs; ts <= v.reservation.arriveConflictZoneTs + constants.SimulationStepInterval; ts += constants.SimulationStepInterval {
		v.approachConflictZonePlan[ts] = speed
		if ts < stopAcc {
			speed += acc * float64(constants.SimulationStepInterval) / 1000.0
		} else if ts < startDecel {
			// speed is the same
		} else {
			speed -= comfortableDecel * float64(constants.SimulationStepInterval) / 1000.0
		}
	}
}
