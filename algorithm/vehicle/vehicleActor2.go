package vehicle

import (
	"algorithm/constants"
	"algorithm/types"
	"algorithm/util"
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
const maxAngularSpeed = 0.5
const maxSpeedOnConflictZone = 15.0

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

	v.planApproachConflictZoneNoReservation(ts)

	return &v
}

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


func (v *VehicleActor) Ping(ts types.Millisecond) {

	v.checkForMessages(ts)

	if v.State == beforeIntersectionNotAllowed {
		if v.isFirstToConflictZone() {
			v.sendRequestReservation(ts)
		} else {
			//if v.platooningReservation != nil {
			//	v.sendRequestReservation(ts) // TODO - inny algorytm wyliczania, gdy jest platoon
			//}
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
				//panic("Oops")
				fmt.Println("Oops (leave), is late: ",  float64(ts - ts0), ", isturning:", v.isTurning())
			}
			v.State = afterIntersection
		}
	}
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
		if v.Speed > desiredSpeed {
			if v.Speed - maxDecel * constants.SimulationStepIntervalSeconds > desiredSpeed {
				//panic("Oops")
			}
		}
		if v.Speed < desiredSpeed {
			if v.Speed + maxAcc * constants.SimulationStepIntervalSeconds < desiredSpeed {
				//panic("Oops")
			}
		}
		v.Speed = desiredSpeed
	} else {
		// vehicle is late, lets continue with the same speed
	}
}


func (v *VehicleActor) checkForMessages(ts types.Millisecond) {
	messages := v.networkCard.VehicleReceive(v.Id)
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
					//panic("xd")
				}
				d -= v0 * constants.SimulationStepIntervalSeconds
			}

			if math.Abs(d) > 0.1 {
				reply := DsrcV2RMessage{
					MsgType:               AimProtocolMsgReservationCancelation,
					ReservationToCancelId: m.reservationId,
				}
				fmt.Println("Cancelling reservation")
				v.networkCard.SendDsrcV2R(&reply)
			} else {
				v.State = beforeIntersectionHasReservation
				v.sendReservationInfoBroadcast(ts, reservation, v.hipotheticalPlan)
				v.reservation = reservation
				v.approachConflictZoneHasReservationPlan = v.hipotheticalPlan
			}
		}
	}

	for _, m := range v.networkCard.VehicleReceiveV2V(v.Id) {
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

func (v *VehicleActor) planApproachToConflictZoneWithReservation(ts types.Millisecond, msg *DsrcR2VMessage) (map[types.Millisecond]types.MetersPerSecond, *reservation, bool) {
	// aktualna predkosc
	v0 := v.Speed
	// prędkość z którą musi wjechac na skrzyzowanie
	v2 := msg.reservationDesiredSpeed
	// czas w którym musi wjechać na skrzyżowanie
	t := float64(msg.reservationFromTs - ts) / 1000.0
	// odległość do skrzyżowania
	s := v.calculateDistanceCenterToConflictZone()

	plan, success := calculatePlanEnterConflictZone(v0, v2, s, t)

	if success == false {
		return nil, nil, false
	}

	resultPlan := make(map[types.Millisecond]types.MetersPerSecond)
	for t0, speed := range plan {
		resultPlan[ts + types.Millisecond(t0 * 1000)] = speed
	}

	reservation := reservation{
		reservationId: msg.reservationId,
		arriveConflictZoneTs: msg.reservationFromTs,
		arriveConflictZoneSpeed: msg.reservationDesiredSpeed,
		leaveConflictZoneTs: msg.reservationToTs,
		speedPerTime: msg.reservationTsToSpeed,
	}

	return resultPlan, &reservation, true
}

func (vehicle *VehicleActor) sendRequestReservation(ts types.Millisecond) {
	if ts % 100 != 0 {
		return
	}
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

	vMax := maxSpeedOnConflictZone
	if vehicle.isTurning() {
		vMax = vehicle.calculateMaxSpeedOnCurve()
	}

	t, v, plan, ok := calculateApproachConflictZoneTimeSpeed(v0, s, vMax)
	t = t - t % 10
	if ok == false {
		panic("Oops")
	}

	for th := range plan {
		_, e := hipotheticalPlan[ts + continueWithoutReservationTime + th]
		if e {
			fmt.Println("Overriding")
		}
		v = plan[th]
		if v == 0 {
			fmt.Println("Missing")
		}
		hipotheticalPlan[ts + continueWithoutReservationTime + th] = plan[th]
	}

	{
		d := vehicle.calculateDistanceCenterToConflictZone()
		for t0 := ts; t0 < ts + continueWithoutReservationTime + t; t0 += constants.SimulationStepInterval {
			v0, e := hipotheticalPlan[t0]
			if e == false {
				panic("xd")
			}
			d -= v0 * constants.SimulationStepIntervalSeconds
		}
	}


	enter, exit := vehicle.conflictZoneNodeEnterExit()
	msg := &DsrcV2RMessage{
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
	d2 := v.calculateDistanceCenterToConflictZone() //- constants.VehicleLength / 2
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
		maxSpeedOnCurve := 2.0 * maxAngularSpeed * radius
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

}

func (vehicle *VehicleActor) planApproachConflictZoneNoReservation(ts types.Millisecond) {
	const planNextSteps = 100 // TODO - tu jest jakiś bug, jak ustawic na 50 nie bedzie dzialac -> zbyt czeste przeladowanie planu powoduje ze nie wychamowuje
	const comfortableDecel = 1.0
	result := make(map[types.Millisecond]types.MetersPerSecond)
	v := vehicle.Speed
	d1 := vehicle.calculateDistanceCenterToConflictZone() - constants.VehicleLength / 2
	d2, _ := vehicle.sensor.ScanVehiclesAhead(vehicle)
	d := math.Min(d1, d2)

	brakingDist := func() float64 {
		t := v / comfortableDecel
		return 0.5 * comfortableDecel * t * t
	}
	emergencyBrakingDist := func() float64 {
		t := v / maxDecel
		return 0.5 * maxDecel * t * t
	}

	if d < 10 {
		//fmt.Println("xd")
	}

	for step := 0; step < planNextSteps; step += 1 {
		result[types.Millisecond(step) * constants.SimulationStepInterval + ts] = v
		//fmt.Println(v)
		d -= v * constants.SimulationStepIntervalSeconds
		if d < 0 {
			//fmt.Println(emergencyBrakingDist())
		}
		if emergencyBrakingDist() >= d {
			v -= maxDecel * constants.SimulationStepIntervalSeconds
			v = math.Max(0, v)
			continue
		}
		if brakingDist() >= d {
			v -= comfortableDecel * constants.SimulationStepIntervalSeconds
			v = math.Max(0, v)
			continue
		}
		v += 1.0 * constants.SimulationStepIntervalSeconds
	}
	//fmt.Println("--------------")

	vehicle.approachConflictZoneNoReservationPlan = result
}

func (vehicle *VehicleActor) planLeaverIntersection(ts types.Millisecond) {
	v := vehicle.Speed
	result := make(map[types.Millisecond]types.MetersPerSecond)
	for t := ts; t < ts + 30*100; t += constants.SimulationStepInterval {
		result[t] = v
		v += maxAcc * constants.SimulationStepIntervalSeconds
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


func calculatePlanEnterConflictZone(v0, v2, s, t float64) (map[types.Millisecond]float64, bool) {
	result := make(map[types.Millisecond]float64)

	v1, t1,t2,t3, success := calcTimeVelocityAccKeepDecel(v0, v2, s, t)
	if success {
		ts1 := types.Millisecond(t1 * 1000)
		ts2 := ts1 + types.Millisecond(t2 * 1000)
		ts3 := ts2 + types.Millisecond(t3 * 1000)

		v := v0
		for t := types.Millisecond(0); t < ts3; t += constants.SimulationStepInterval {
			if t < ts1 {
				v += maxAcc * constants.SimulationStepIntervalSeconds
			} else if t < ts2 {
				// const speed
				v = v1
			} else {
				v -= maxDecel* constants.SimulationStepIntervalSeconds
				if v < 0.0 {
					v = 0
				}
			}

			result[t] = v
			result[t+10] = v
			result[t+20] = v
			result[t+30] = v
		}

		return result, true
	}

	v1, t1,t2,t3, success = calcTimeVelocityDecelKeepAcc(v0, v2, s, t)
	if success { // TODO - copy pasted code
		ts1 := types.Millisecond(t1 * 1000)
		ts2 := ts1 + types.Millisecond(t2 * 1000)
		ts3 := ts2 + types.Millisecond(t3 * 1000)

		v := v0
		for t := types.Millisecond(0); t < ts3; t += constants.SimulationStepInterval {

			if t < ts1 {
				v -= maxDecel * constants.SimulationStepIntervalSeconds
				if v < 0.0 {
					v = 0.0
				}
			} else if t < ts2 {
				// const speed
				v = v1
			} else {
				v += maxAcc * constants.SimulationStepIntervalSeconds
			}

			result[t] = v
			result[t+10] = v
			result[t+20] = v
			result[t+30] = v
		}

		return result, true
	}

	v1, t1,t2,t3, success = calcTimeVelocityKeepAccDecel(v0, v2, s, t)
	if success { // TODO - copy pasted code
		ts1 := types.Millisecond(t1 * 1000)
		ts2 := ts1 + types.Millisecond(t2 * 1000)
		ts3 := ts2 + types.Millisecond(t3 * 1000)

		v := v0
		for t := types.Millisecond(0); t < ts3; t += constants.SimulationStepInterval {

			if t < ts1 {
				v = v1
			} else if t < ts2 {
				v += maxAcc * constants.SimulationStepIntervalSeconds
			} else {
				v -= maxDecel * constants.SimulationStepIntervalSeconds
			}

			result[t] = v
			result[t+10] = v
			result[t+20] = v
			result[t+30] = v
		}

		return result, true
	}

	return nil, false
}


func calculateApproachConflictZoneTimeSpeed(v0, s, v2 float64) (types.Millisecond, float64, map[types.Millisecond]types.MetersPerSecond, bool) {
	const maxError = 0.001
	{
		if v2 > v0 {
			t1 := (v2 - v0) / maxAcc
			s1 := v0 * t1 + 0.5 * maxAcc * t1 * t1
			if s1 >= s {
				// przyspieszamy na całym odcinku
				trueV2 := math.Sqrt(v0 * v0 + 2 * s * maxAcc)
				ts1 := types.Millisecond((trueV2 - v0) / maxAcc * 1000)

				v := v0
				plan := make(map[types.Millisecond]types.MetersPerSecond)
				for t := types.Millisecond(0); t < ts1; t += constants.SimulationStepInterval {
					v += maxAcc * constants.SimulationStepIntervalSeconds
					plan[t] = v
				}

				return ts1, trueV2, plan, true
			}
		}

		if v2 < v0 {
			t1 := (v0 - v2) / maxDecel
			s1 := v0 * t1 + 0.5 * maxAcc * t1 * t1
			if s1 > s {
				// fail - nie zdazymy nawet wychamowac do takiej predkosci
				return 0,0, nil, false
			}
		}
	}
	// przez t1 przyspieszamy, przez t2 hamujemy
	var tryV1Range func (v_a, v_b float64) (float64, float64)
	depth := 0
	tryV1Range = func  (v_a, v_b float64) (float64, float64) {
		if depth > 200 {
			panic("xd")
		}
		depth += 1
		v1 := (v_b + v_a) / 2

		t1 := (v1 - v0) / maxAcc
		t2 := (v1 - v2) / maxDecel

		s1 := v0 * t1 + 0.5 * maxAcc * t1 * t1 + v2 * t2 + 0.5 * maxDecel * t2 * t2
		if math.Abs(s - s1) < maxError {

			if t1 < 0 || t2 < 0 {
				fmt.Println("OOps")
			}

			return t1, t2
		}

		if s1 < s {
			return tryV1Range(v1, v_b)
		} else {
			return tryV1Range(v_a, v1)
		}
	}

	t1, t2 := tryV1Range(v0, 10e4)
	v := v0
	ts1 := types.Millisecond((t1) * 1000)
	ts2 := types.Millisecond((t2) * 1000)
	plan := make(map[types.Millisecond]types.MetersPerSecond)
	for t := types.Millisecond(0); t < ts1 + ts2; t += constants.SimulationStepInterval {
		if t < ts1 {
			v += maxAcc * constants.SimulationStepIntervalSeconds
		} else {
			v -= maxDecel* constants.SimulationStepIntervalSeconds
		}
		plan[t] = v
	}

	return ts1 + ts2, v2, plan, true
}



func calcTimeVelocityAccKeepDecel(v0, v2, s, t float64) (float64, float64, float64, float64, bool) {
	{
		if v0 < v2 {
			vdiff := maxAcc * t
			if v0 + vdiff < v2 {
				return 0, 0,0,0, false
			}
		} else {
			vdiff := maxDecel * t
			if v0 - vdiff > v2 {
				return 0, 0,0,0, false
			}
		}
	}

	const maxError = 0.5
	var tryV1Range func(float64, float64) (float64, float64, float64, float64, bool)
	depth := 0
	tryV1Range = func(vMin, vMax float64) (float64, float64, float64, float64, bool) {
		if depth > 1000 {
			return 0, 0 ,0, 0, false
		}
		depth += 1
		v1 := (vMax + vMin) / 2

		t1 := (v1 - v0) / maxAcc
		t3 := (v1 - v2) / maxDecel
		if t1 + t3 > t { // v1 to large
			return tryV1Range(vMin, v1)
		}

		t2 := t - t1 - t3
		s1 := v0 * t1 + 0.5 * maxAcc * t1 * t1 + v1 * t2 + v2 * t3 + 0.5 * maxDecel  * t3 * t3

		if math.Abs(s - s1) < maxError {
			return v1, t1, t2, t3, true
		}

		if v1 == vMax || v1 == vMin {
			return 0 ,0 ,0, 0, false
		}

		if s1 > s {
			return tryV1Range(vMin, v1)
		} else {
			return tryV1Range(v1, vMax)
		}
	}

	vMax := 100.0
	vMin := math.Max(v0, v2)

	v1, t1, t2, t3, success := tryV1Range(vMin, vMax)

	if t1 < 0.0 || t2 < 0.0 || t3 < 0.0 || v1 < 0.0 {
		return 0, 0,0,0,false
	}
	return v1, t1, t2, t3, success
}

func calcTimeVelocityKeepAccDecel(v0, v2, s, t float64) (float64, float64, float64, float64, bool) {
	const maxError = 0.5
	var tryV1Range func(float64, float64) (float64, float64, float64, float64, bool)
	depth := 0
	tryV1Range = func(vMin, vMax float64) (float64, float64, float64, float64, bool) {
		if depth > 1000 {
			return 0, 0 ,0, 0, false
		}
		depth += 1
		v1 := (vMax + vMin) / 2

		t3 := (v1 - v2) / maxDecel
		t2 := (v1 - v0) / maxAcc
		if t3 + t2 > t {
			return tryV1Range(vMin, v1)
		}
		t1 := t - t2 - t3

		s1 := t1 * v0 + t2 * v0 + 0.5 * maxAcc * t2 * t2 + v2 * t3 + 0.5 * maxDecel * t3 * t3

		if math.Abs(s - s1) < maxError {
			return v1, t1, t2, t3, true
		}

		if v1 == vMax || v1 == vMin {
			return 0,0 ,0, 0, false
		}

		if s1 > s {
			return tryV1Range(vMin, v1)
		} else {
			return tryV1Range(v1, vMax)
		}
	}

	v1, t1, t2, t3, success :=  tryV1Range(v2, 10e4)

	if t1 < 0.0 || t2 < 0.0 || t3 < 0.0 || v1 < 0.0 {
		return 0, 0,0,0,false
	}
	return v1, t1, t2, t3, success
}


func calcTimeVelocityDecelKeepAcc(v0, v2, s, t float64) (float64, float64, float64, float64, bool) {
	const maxError = 0.5
	var tryV1Range func(float64, float64) (float64, float64, float64, float64, bool)
	depth := 0
	tryV1Range = func(vMin, vMax float64) (float64, float64, float64, float64, bool) {
		if depth > 1000 {
			return 0, 0 ,0, 0, false
		}
		depth += 1
		v1 := (vMax + vMin) / 2

		t1 := (v0 - v1) / maxDecel
		t3 := (v2 - v1) / maxAcc
		if t1 + t3 > t { // v1 to small
			return tryV1Range(v1, vMax)
		}

		t2 := t - t1 - t3
		s1 := v1 * t + 0.5 * maxDecel * t1 * t1 + 0.5 * maxAcc * t3 * t3

		if math.Abs(s - s1) < maxError {
			return v1, t1, t2, t3, true
		}

		if v1 == vMax || v1 == vMin {
			return 0,0 ,0, 0, false
		}

		if s1 > s {
			return tryV1Range(vMin, v1)
		} else {
			return tryV1Range(v1, vMax)
		}
	}

	v1, t1, t2, t3, success :=  tryV1Range(0, v0)

	if t1 < 0.0 || t2 < 0.0 || t3 < 0.0 || v1 < 0.0 {
		return 0, 0,0,0,false
	}
	return v1, t1, t2, t3, success
}










































































