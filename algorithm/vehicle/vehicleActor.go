package vehicle

import (
	"algorithm/constants"
	"algorithm/types"
	"algorithm/util"
	"fmt"
	"math"
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
	State       vehicleState

	entryPoint                             *util.Node
	exitPoint                              *util.Node
	roadGraph                              *util.Graph
	sensor                                 *SensorLayer
	networkCard                            *CommunicationLayer
	route                                  []*util.Edge
	reservation                            *reservation
	platooningReservation                  *platooningReservation
	approachConflictZoneNoReservationPlan  map[types.Millisecond]types.MetersPerSecond
	approachConflictZoneHasReservationPlan map[types.Millisecond]types.MetersPerSecond
	leaverIntersectionPlan                 map[types.Millisecond]types.MetersPerSecond
	AlphaInitiated                         bool
	hipotheticalPlan                       map[types.Millisecond]types.MetersPerSecond
	lastMsgSentTs                          types.Millisecond
}


func NewVehicleActor(id types.VehicleId, ts types.Millisecond, entrypoint, exitpoint *util.Node, initSpeed types.MetersPerSecond, roadGraph *util.Graph, sensor *SensorLayer, nc *CommunicationLayer) *VehicleActor {
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
		networkCard:    nc,
		lastMsgSentTs: 	0,
	}

	v.planRoute()

	v.planApproachConflictZoneNoReservation(ts)

	return &v
}

func (v *VehicleActor) Ping(ts types.Millisecond) {

	v.checkForMessages(ts)

	if v.State == beforeIntersectionNotAllowed {
		if v.isFirstToConflictZone() {
			if ts - v.lastMsgSentTs > 50 {
				v.sendRequestReservation(ts)
				v.lastMsgSentTs = ts
			}
		} else {
			v.sendRequestPermissionPlatooning(ts)
		}
	}

	v.controlVelocity(ts)

	v.move(v.Speed * constants.SimulationStepIntervalSeconds)

	switch v.State {
	case beforeIntersectionNotAllowed:
		if v.isCenterInConflictZone() { panic("Oops") }
	case beforeIntersectionHasReservation:
		if v.isCenterInConflictZone() {
			ts0 := v.reservation.arriveConflictZoneTs
			if math.Abs(float64(ts - ts0)) > 100 {
				panic("Oops")
			}
			v.State = atIntersection
		}
	case atIntersection:
		if v.isCenterInConflictZone() == false {
			ts0 := v.reservation.leaveConflictZoneTs
			if math.Abs(float64(ts - ts0)) > 100 {
				panic("Oops")
			}
			v.State = afterIntersection
		}
	}
}

func Initiate(conf util.Configuration) {
	vehiclePower = 745.699872 * conf.VehiclePower
	vehicleBrakingForce = conf.VehicleBrakingForce
	vehicleweight = conf.VehicleWeight
	vehicleMaxAngularSpeed = conf.VehicleMaxAngularSpeed
	vehicleMaxSpeedOnConflictZone = conf.VehicleMaxSpeedOnConflictZone
}

var vehiclePower = 120.0 * 745.699872 		// horse power
var vehicleBrakingForce = 3000.0      		// N
var vehicleweight = 1200.0            		// kilogram
var vehicleMaxAngularSpeed = 0.4			// radians/s
var vehicleMaxSpeedOnConflictZone = 15.0	// m/s

type vehicleState = int
const (
	beforeIntersectionNotAllowed vehicleState = 1 + iota
	beforeIntersectionHasReservation
	atIntersection
	afterIntersection
)


type reservation struct {
	reservationId		types.ReservationId
	arriveConflictZoneTs types.Millisecond
	arriveConflictZoneSpeed types.MetersPerSecond
	leaveConflictZoneTs types.Millisecond
	speedPerTime map[types.Millisecond]types.MetersPerSecond
}

type platooningReservation struct {
	reservationId            types.ReservationId
	arriveConflictZoneTs     types.Millisecond
	arriveConflictZoneSpeed  types.MetersPerSecond
	approachConflictZonePlan map[types.Millisecond]types.MetersPerSecond
}


func (v *VehicleActor) controlVelocity(ts types.Millisecond) {
	var desiredSpeed float64
	var exists bool
	switch v.State {
	case beforeIntersectionNotAllowed:
		desiredSpeed, exists = v.approachConflictZoneNoReservationPlan[ts]
		if exists == false {
			v.planApproachConflictZoneNoReservation(ts)
			desiredSpeed, exists = v.approachConflictZoneNoReservationPlan[ts]
		}
	case beforeIntersectionHasReservation:
		desiredSpeed, exists = v.approachConflictZoneHasReservationPlan[ts]
	case atIntersection:
		desiredSpeed, exists = v.reservation.speedPerTime[ts]
		if exists == false {
			if ts < v.reservation.arriveConflictZoneTs {
				if ts - v.reservation.arriveConflictZoneTs < 100 {
					desiredSpeed = v.Speed
				} else {
					panic("Oops")
				}
			} else {
				if v.reservation.arriveConflictZoneTs - ts < 100 {
					desiredSpeed = v.Speed
				} else {
					panic("Oops")
				}
			}
		}
	case afterIntersection:
		desiredSpeed, exists = v.leaverIntersectionPlan[ts]
		if exists == false {
			v.planLeaverIntersection(ts)
			desiredSpeed, exists = v.leaverIntersectionPlan[ts]
		}
	}

	if exists {
		v.Speed = desiredSpeed
	} else {
		// vehicle is late, lets continue with the same speed
	}
}


func (v *VehicleActor) checkForMessages(ts types.Millisecond) {
	messages := v.networkCard.VehicleReceive(ts, v.Id)
	for _, m := range messages {
		switch m.msgType {
		case AimProtocolMsgAllow:
			if v.isFirstToConflictZone() == false {
				//panic("Oops")
				fmt.Println("Error: vId=", v.Id, " got reservation but is not first")
			}

			reservation := &reservation{
				reservationId: m.reservationId,
				arriveConflictZoneTs: m.reservationFromTs,
				arriveConflictZoneSpeed: m.reservationDesiredSpeed,
				leaveConflictZoneTs: m.reservationToTs,
				speedPerTime: m.reservationTsToSpeed,
			}

			d := v.calculateDistanceCenterToConflictZone()
			for t0 := ts; t0 < reservation.arriveConflictZoneTs; t0 += constants.SimulationStepInterval {
				v0, e := v.hipotheticalPlan[t0]
				if e == false {
					//panic("Oops")
				}
				d -= v0 * constants.SimulationStepIntervalSeconds
			}

			if math.Abs(d) > 0.25 {
				reply := DsrcV2RMessage{
					MsgType:               AimProtocolMsgReservationCancelation,
					ReservationToCancelId: m.reservationId,
				}
				//fmt.Println("Cancelling reservation")
				v.networkCard.SendDsrcV2R(reply)
			} else {
				v.State = beforeIntersectionHasReservation
				v.sendReservationInfoBroadcast(ts, reservation, v.hipotheticalPlan)
				v.reservation = reservation
				v.approachConflictZoneHasReservationPlan = v.hipotheticalPlan
				for t := ts; true; t += constants.SimulationStepInterval {
					_, e := v.approachConflictZoneHasReservationPlan[t]
					if e == false {
						break
					}
					//fmt.Println(v.approachConflictZoneHasReservationPlan[t])
				}
				//fmt.Println("=----------------------------")
			}
		}
	}

	for _, m := range v.networkCard.VehicleReceiveV2V(ts, v.Id) {
		switch m.msgType {
		case AimProtocolMsgReservationInfo:
			_, vId := v.sensor.ScanVehiclesAhead(v)
			if vId != m.sender {
				continue
			}

			v.platooningReservation = &platooningReservation{
				reservationId: m.reservationId,
				arriveConflictZoneTs: m.arriveConflictZoneTs,
				arriveConflictZoneSpeed: m.arriveConflictZoneSpeed,
				approachConflictZonePlan: m.approachConflictZonePlan,
			}
		}
	}
}


func (vehicle *VehicleActor) sendRequestReservation(ts types.Millisecond) {
	if vehicle.AlphaInitiated == false {
		return
	}
	const continueWithoutReservationTime = 500 // [ms]

	hipotheticalPlan := make(map[types.Millisecond]types.MetersPerSecond)

	// prędkość oraz dystans do conflict zone w oczekiwanym momencie otrzymania pozwolenia
	_, e := vehicle.approachConflictZoneNoReservationPlan[ts + continueWithoutReservationTime]
	if e == false {
		vehicle.planApproachConflictZoneNoReservation(ts)
	}
	v0, s1 := calculateDistSpeedAfter(vehicle.approachConflictZoneNoReservationPlan, ts, continueWithoutReservationTime)
	s := vehicle.calculateDistanceCenterToConflictZone() - s1

	for t := ts; t < ts + continueWithoutReservationTime; t += constants.SimulationStepInterval {
		hipotheticalPlan[t] = vehicle.approachConflictZoneNoReservationPlan[t]
	}

	vMax := vehicleMaxSpeedOnConflictZone
	if vehicle.isTurning() {
		vMax = vehicle.calculateMaxSpeedOnCurve()
	}


	t, v, plan, ok := calculateApproachConflictZoneTimeSpeed(v0, s, vMax)
	t = t - t % 10
	if ok == false {
		// TODO
		//fmt.Println("Vehicle cannot approach conflict zone. v0, s, vMax = ", v0, s, vMax)
		return
		//panic("Oops")
	}

	for th := range plan {
		hipotheticalPlan[ts + continueWithoutReservationTime + th] = plan[th]
	}

	{
		d := vehicle.calculateDistanceCenterToConflictZone()
		for t0 := ts; t0 < ts + continueWithoutReservationTime + t; t0 += constants.SimulationStepInterval {
			v0, e := hipotheticalPlan[t0]
			if e == false {
				panic("Oops")
			}
			d -= v0 * constants.SimulationStepIntervalSeconds
		}
	}

	enter, exit := vehicle.conflictZoneNodeEnterExit()
	msg := DsrcV2RMessage{
		MsgType: AimProtocolMsgRequest,
		TsSent: ts,
		Sender: vehicle.Id,
		VehicleX: vehicle.X,
		VehicleY: vehicle.Y,
		VehicleSpeed: vehicle.Speed,
		ApproachConflictZoneMinTs: ts + t + continueWithoutReservationTime,
		ApproachConflictZoneSpeed: v,
		ConflictZoneNodeEnter: enter,
		ConflictZoneNodeExit: exit,
		MaxSpeedOnCurve: vehicle.calculateMaxSpeedOnCurve(),
		IsTurning: vehicle.isTurning(),
		EntryPointId: vehicle.entryPoint.Id,
		ExitPointId: vehicle.exitPoint.Id,
		Route: vehicle.getRouteCoordinates(),
	}
	vehicle.networkCard.SendDsrcV2R(msg)
	vehicle.hipotheticalPlan = hipotheticalPlan
}

func (v *VehicleActor) isFirstToConflictZone() bool {
	d1, _ := v.sensor.ScanVehiclesAhead(v)
	d2 := v.calculateDistanceCenterToConflictZone()
	return d1 > d2
}

func (v *VehicleActor) sendRequestPermissionPlatooning(ts types.Millisecond) {
	// FIXME - implement this
}

func (v *VehicleActor) getRouteCoordinates() []types.Location {
	if v.isTurning() {
		res := make([]types.Location, 0)
		res = append(res, types.Location{X: v.entryPoint.X, Y: v.entryPoint.Y})

		for i := 1 ; i < len(v.route); i++ {
			if v.route[i].IsArc == false {
				break
			}
			res = append(res, types.Location{X: v.route[i].To.X, Y: v.route[i].To.Y})
		}
		return res
	} else {
		res := make([]types.Location, 2)
		res[0] = types.Location{X: v.route[0].To.X, Y: v.route[0].To.Y}
		res[1] = types.Location{X: v.route[1].To.X, Y: v.route[1].To.Y}
		return res
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

func (v *VehicleActor) calculateMaxSpeedOnCurve() types.MetersPerSecond {
	if v.isTurning() {
		r := v.getRouteCoordinates()
		x1 := r[1].X
		x2 := r[len(r) - 1].X
		radius := math.Abs(x1 - x2)
		maxSpeedOnCurve := 2.0 * vehicleMaxAngularSpeed * radius
		if radius == 0.0 {
			fmt.Println("Oops")
		}
		return maxSpeedOnCurve
	} else {
		return math.MaxFloat64
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

func (v *VehicleActor) sendReservationInfoBroadcast(ts types.Millisecond, reservation *reservation, plan map[types.Millisecond]types.MetersPerSecond) {
	// FIXME - implement this
}

func (vehicle *VehicleActor) planApproachConflictZoneNoReservation(ts types.Millisecond) {
	const planNextSteps = 100
	result := make(map[types.Millisecond]types.MetersPerSecond)
	v := vehicle.Speed
	d1 := vehicle.calculateDistanceCenterToConflictZone() - constants.VehicleLength / 2
	d2, _ := vehicle.sensor.ScanVehiclesAhead(vehicle)
	d := math.Min(d1, d2)

	brakingDist := func() float64 {
		w := 0.5 * vehicleweight * v * v
		s := w / vehicleBrakingForce
		return s
	}

	for step := 0; step < planNextSteps; step += 1 {
		result[types.Millisecond(step) * constants.SimulationStepInterval + ts] = v
		d -= v * constants.SimulationStepIntervalSeconds
		if brakingDist() >= d {
			v -= velocityDiffStepBraking(v)
			v = math.Max(0, v)
			continue
		}
		v += 1.0 * constants.SimulationStepIntervalSeconds
	}

	vehicle.approachConflictZoneNoReservationPlan = result
}

func (vehicle *VehicleActor) planLeaverIntersection(ts types.Millisecond) {
	v := vehicle.Speed
	result := make(map[types.Millisecond]types.MetersPerSecond)
	for t := ts; t < ts + 30*100; t += constants.SimulationStepInterval {
		result[t] = v
		v += velocityDiffStepAccelerating(v)
	}
	vehicle.leaverIntersectionPlan = result
}


func calculateDistSpeedAfter(plan map[types.Millisecond]types.MetersPerSecond, ts types.Millisecond, timeDuration types.Millisecond) (types.MetersPerSecond, types.Meter) {
	s := 0.0
	v := 0.0
	for t := ts; t < ts + timeDuration; t += constants.SimulationStepInterval {
		v0, e := plan[t]
		if e == false {
			panic("Oops")
		}
		s += v0 * constants.SimulationStepIntervalSeconds
		v = v0
	}
	return v, s
}



func calculateApproachConflictZoneTimeSpeed(v0, s, v2 float64) (types.Millisecond, float64, map[types.Millisecond]types.MetersPerSecond, bool) {
	const maxError = 0.2

	var tryTargetVRange func (v_a, v_b float64) (float64, map[types.Millisecond]types.MetersPerSecond, types.Millisecond)
	depth := 0
	tryTargetVRange = func  (v_a, v_b float64) (float64, map[types.Millisecond]types.MetersPerSecond, types.Millisecond) {
		if depth > 200 {
			return 0, nil, 0
		}
		depth += 1

		targetV := (v_a + v_b) / 2
		t :=  types.Millisecond(0)
		currV := v0
		sTotal := 0.0
		plan := make(map[types.Millisecond]types.MetersPerSecond)

		for currV < targetV {
			plan[t] = currV
			sTotal += currV * constants.SimulationStepIntervalSeconds
			t += constants.SimulationStepInterval
			currV += velocityDiffStepAccelerating(currV)

			if math.Abs(sTotal - s) < maxError && currV < v2 {
				return currV, plan, t
			}

			if sTotal < s && s < sTotal + currV * constants.SimulationStepIntervalSeconds &&
				sTotal + currV * constants.SimulationStepIntervalSeconds - s > maxError {

					return tryTargetVRange(v_a, targetV)
			}
		}

		for currV > v2 {
			plan[t] = currV
			t += constants.SimulationStepInterval
			currV -= velocityDiffStepBraking(currV)
			sTotal += currV * constants.SimulationStepIntervalSeconds

			if sTotal < s && s < sTotal + currV * constants.SimulationStepIntervalSeconds &&
				sTotal + currV * constants.SimulationStepIntervalSeconds - s > maxError {

				return tryTargetVRange(v_a, targetV)
			}
		}

		if math.IsNaN(math.Abs(sTotal - s)) {
			panic("Oops")
		}

		if math.Abs(sTotal - s) < maxError && math.Abs(currV - v2) < maxError  {
			return currV, plan, t
		}

		if sTotal > s {
			return tryTargetVRange(v_a, targetV)
		} else {
			return tryTargetVRange(targetV, v_b)
		}
	}

	finalV, plan, time := tryTargetVRange(v0, 100)
	if plan == nil {
		return 0, 0, nil, false
	}

	return time, finalV, plan, true
}


func velocityDiffStepBraking(v float64) float64 {
	s := v * constants.SimulationStepIntervalSeconds
	w_diff := vehicleBrakingForce * s
	if v * v - 2.0 * w_diff /vehicleweight < 0 {
		return v
	}
	v2 := math.Sqrt(v * v - 2.0 * w_diff /vehicleweight)
	if v2 >= v && v != 0 {
		panic("Oops")
	}
	return v - v2
}

func velocityDiffStepAccelerating(v float64) float64 {
	w_diff := vehiclePower * constants.SimulationStepIntervalSeconds
	v2 := math.Sqrt(v * v + 2.0 * w_diff /vehicleweight)
	if v >= v2 {
		panic("Oops")
	}
	return v2 - v
}


































































