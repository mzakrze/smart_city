package vehicle

import (
	"algorithm/constants"
	"algorithm/types"
	"algorithm/util"
	"math"
	"runtime"
)

type GridState int
const (
	free GridState = iota + 1
	taken
)

type IntersectionPolicyFcfs struct {
	reservationTable  [][][]GridState
	replies           []DsrcR2VMessage
	reservations      map[types.ReservationId]ipFcfsReservation
	nextReservationId types.ReservationId
	graph             *util.Graph
	tableIndex        types.Millisecond
}
const gridSize types.Meter = 0.1
const tableSize = 60 * 100 // na 60 sekund

type ipFcfsReservation struct {
	reservationId types.ReservationId
	startTs       types.Millisecond
	endTs         types.Millisecond
	reservedGrids [][]grid
}

var gridNoX = -1
var gridNoY = -1
var conflictZoneMinX = -1.0
var conflictZoneMinY = -1.0

type grid struct {
	x int
	y int
}

func CreateIntersectionPolicyFcfs(graph *util.Graph, configuration util.Configuration) *IntersectionPolicyFcfs {
	// TODO - nieużywany argument
	gridNoX = int((graph.ConflictZone.MaxX - graph.ConflictZone.MinX) / gridSize)
	gridNoY = int((graph.ConflictZone.MaxY - graph.ConflictZone.MinY) / gridSize)
	conflictZoneMinX = graph.ConflictZone.MinX
	conflictZoneMinY = graph.ConflictZone.MinY

	reservationTable :=  make([][][]GridState, tableSize)
	for i := 0; i < len(reservationTable); i++ {
		reservationTable[i] = make([][]GridState, gridNoX)
		for x := range reservationTable[i] {
			reservationTable[i][x] = make([]GridState, gridNoY)
			for y := range reservationTable[i][x] {
				reservationTable[i][x][y] = free
			}
		}
	}

	r := &IntersectionPolicyFcfs{
		replies: []DsrcR2VMessage{},
		reservationTable: reservationTable,
		reservations: make(map[types.ReservationId]ipFcfsReservation),
		nextReservationId: 1,
		graph: graph,
		tableIndex: 0,
	}

	if r.calcOccupied() > 0 {
		panic("Oops")
	}

	return r
}

func (ip * IntersectionPolicyFcfs) ProcessMsg(m DsrcV2RMessage) {
	if m.MsgType == AimProtocolMsgReservationCancelation {
		ip.cancelReservation(m.ReservationToCancelId)
		return
	}

	success := ip.makeReservationIfFitsInReservationTable(m)
	if success == false {
		return
	}
}

func (ip *IntersectionPolicyFcfs) GetReplies(ts types.Millisecond) []DsrcR2VMessage {
	res := ip.replies
	ip.replies = []DsrcR2VMessage{}
	if ts % 1000 == 0 {
		ip.cleanupOldReservations(ts)
	}
	ip.relocateTableIfNecessary(ts)

	return res
}

func (ip *IntersectionPolicyFcfs) cleanupOldReservations(ts types.Millisecond) {
	toDelete := []types.ReservationId{}
	for rId := range ip.reservations {
		if ip.reservations[rId].endTs < ts {
			toDelete = append(toDelete, rId)
		}
	}

	for _, rId := range toDelete {
		delete(ip.reservations, rId)
	}

	runtime.GC()

	//bToMb := func (b uint64) uint64 {
	//	return b / 1024 / 1024
	//}

	//var m runtime.MemStats
	//runtime.ReadMemStats(&m)
	//// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	//fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	//fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	//fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	//fmt.Printf("\tNumGC = %v\n", m.NumGC)
	//
	//if bToMb(m.Sys) > 8000 {
	//	panic("To much memory allocated")
	//}
}

func xToGridX(x types.XCoord) int {
	return int(float64(x) / gridSize)
}

func yToGridY(y types.YCoord) int {
	return int(float64(y) / gridSize)
}

func min(ns ...int) int {
	min0 := math.MaxInt32
	for _,n := range ns {
		if n < min0 {
			min0 = n
		}
	}
	return min0
}

func max(ns ...int) int {
	max0 := math.MinInt32
	for _, n := range ns {
		if n > max0 {
			max0 = n
		}
	}
	return max0
}

func (ip *IntersectionPolicyFcfs) getPointerByIdEntryExit (entryId, exitId types.NodeId) (*util.Node, *util.Node) {
	if entryId == exitId {
		panic("Oops")
	}
	var entry *util.Node = nil
	var exit *util.Node = nil
	for i, e := range ip.graph.Entrypoints {
		if e.Id == entryId {
			entry = ip.graph.Entrypoints[i]
		}
	}
	for i, e := range ip.graph.Exitpoints {
		if e.Id == exitId {
			exit = ip.graph.Exitpoints[i]
		}
	}
	if entry == nil { panic("Oops") }
	if exit == nil { panic("Oops") }

	return entry, exit
}

func (ip *IntersectionPolicyFcfs) makeReservationIfFitsInReservationTable(msg DsrcV2RMessage) bool {
	var moveVehicle func(meter types.Meter)
	var accelerateVehicle func()
	var exitX, exitY float64
	var ts types.Millisecond
	vehicleRoute := msg.Route
	reservedGrids := [][]grid{}
	var vMax types.MetersPerSecond
	if msg.IsTurning {
		vMax = msg.MaxSpeedOnCurve
	} else {
		vMax = vehicleMaxSpeedOnConflictZone
	}

	var reservationTsToSpeed map[types.Millisecond]types.MetersPerSecond
	var x types.XCoord
	var y types.YCoord
	var alpha types.Angle
	var speed types.MetersPerSecond
	var currentRouteIndex int

	exited := func() bool {
		return x == exitX && y == exitY
	}

	moveVehicle = func(distSpare types.Meter) {
		if distSpare == 0.0 {
			return
		}

		if currentRouteIndex == len(vehicleRoute) {
			x = exitX
			y = exitY
			return
		}

		dest := vehicleRoute[currentRouteIndex]

		xDiff := dest.X - x
		yDiff := dest.Y - y
		d := math.Sqrt(xDiff * xDiff + yDiff * yDiff)

		if d <= distSpare {
			distSpare -= d

			x = dest.X
			y = dest.Y

			currentRouteIndex += 1

			moveVehicle(distSpare)
		} else {
			moveX := distSpare * xDiff / d
			moveY := distSpare * yDiff / d

			x += moveX
			y += moveY

			switch {
			case moveX == 0 && moveY > 0: // up
				alpha = -math.Pi / 2
			case moveX == 0 && moveY < 0: // down
				alpha = math.Pi / 2
			case moveY == 0 && moveX > 0: // right
				alpha = 0
			case moveY == 0 && moveX < 0: // left
				alpha = math.Pi
			default:
				alpha = math.Atan(-moveY / moveX)
			}
		}
	}

	accelerateVehicle = func () {
		if speed == vMax {
			return
		}

		diff := velocityDiffStepAccelerating(speed)
		if speed + diff >= vMax {
			speed = vMax
		} else if speed + diff < vMax {
			speed += diff
		} else {
			panic("Oops")
		}
	}

	reservationFromTs := msg.ApproachConflictZoneMinTs

	if msg.IsTurning {
		vehicleRoute = vehicleRoute[1:]
	}
	x = vehicleRoute[0].X
	y = vehicleRoute[0].Y
	exitX = vehicleRoute[len(vehicleRoute) - 1].X
	exitY = vehicleRoute[len(vehicleRoute) - 1].Y

	alpha = getAlpha(x, y, exitX, exitY)
	speed = msg.ApproachConflictZoneSpeed
	currentRouteIndex = 0
	ts = reservationFromTs
	reservationTsToSpeed = make(map[types.Millisecond]types.MetersPerSecond)

	fitsInReservationTable := true
	guard := 0
	outer_for_label:
	for ; exited() == false; ts += constants.SimulationStepInterval {
		if guard > 5000 {
			panic("Oops")
		}
		guard += 1

		grids := getOccupingGrids(x, y, alpha)

		reservedGrids = append(reservedGrids, grids)
		reservationTsToSpeed[ts] = speed

		for i := range grids {
			x, y := grids[i].x, grids[i].y
			frame := ip.reservationTable[ip.tableIndex/10 + ts/10]
			if frame[x][y] == free {
				continue
			}
			if frame[x][y] == taken {
				fitsInReservationTable = false
				break outer_for_label
			}
			panic("Oops")
		}

		accelerateVehicle()
		d := speed * float64(constants.SimulationStepInterval) / 1000.0
		moveVehicle(d)

	}

	if fitsInReservationTable == false {
		return false
	}

	// vehicle fits in the reservation table - so let's actually reserve it
	ts = reservationFromTs
	for i := range reservedGrids {
		for j := range reservedGrids[i] {
			x := reservedGrids[i][j].x
			y := reservedGrids[i][j].y

			if ip.reservationTable[ip.tableIndex/10 + ts/10][x][y] == taken {
				panic("OOps")
			}
			ip.reservationTable[ip.tableIndex/10 + ts/10][x][y] = taken
		}
		ts += constants.SimulationStepInterval
	}

	reservation := ipFcfsReservation{
		reservationId: ip.nextReservationId,
		startTs: reservationFromTs,
		endTs: ts,
		reservedGrids: reservedGrids,
	}

	reply := DsrcR2VMessage{
		reservationId: ip.nextReservationId,
		msgType: AimProtocolMsgAllow,
		receiver: msg.Sender,
		reservationFromTs: reservationFromTs,
		reservationToTs: ts,
		reservationDesiredSpeed: msg.ApproachConflictZoneSpeed,
		reservationTsToSpeed: reservationTsToSpeed,
	}

	ip.nextReservationId += 1
	ip.reservations[reservation.reservationId] = reservation
	ip.replies = append(ip.replies, reply)


	return true
}

func (ip *IntersectionPolicyFcfs) cancelReservation(reservationId types.ReservationId) {
	reservation, e := ip.reservations[reservationId]
	if e == false {
		panic("Oops")
	}

	counter := 0
	for t := range reservation.reservedGrids {

		for i := range reservation.reservedGrids[t] {
			x := reservation.reservedGrids[t][i].x
			y := reservation.reservedGrids[t][i].y

			index := ip.tableIndex/10 + reservation.startTs/10 + types.Millisecond(t)
			v := ip.reservationTable[index][x][y]
			if v != taken {
				panic("Oops")
			}

			ip.reservationTable[index][x][y] = free
			counter += 1
		}
	}

	delete(ip.reservations, reservationId)
}

func (ip *IntersectionPolicyFcfs) calcOccupied() int {
	calc := 0
	for t := range ip.reservationTable {
		for x := range ip.reservationTable[t] {
			for y := range ip.reservationTable[t][x] {
				if ip.reservationTable[t][x][y] == taken {
					calc += 1
				}
			}
		}
	}
	return calc
}

func (ip *IntersectionPolicyFcfs) relocateTableIfNecessary(ts types.Millisecond) {
	const size = 1000
	if ts % size != 0 || ts == 0 {
		return
	}
	ip.tableIndex -= size

	for i := size/10; i < len(ip.reservationTable); i++ {
		for x := range ip.reservationTable[i] {
			for y := range ip.reservationTable[i][x] {

				ip.reservationTable[i - size/10][x][y] = ip.reservationTable[i][x][y]

			}
		}
	}

	framesCleaned := 0
	for i := len(ip.reservationTable) - size/10; i < len(ip.reservationTable); i++ {
		framesCleaned += 1
		for x := range ip.reservationTable[i] {
			for y := range ip.reservationTable[i][x] {
				ip.reservationTable[i][x][y] = free
			}
		}
	}
}


func getOccupingGrids(x types.XCoord, y types.YCoord, alpha types.Angle) []grid {
	debug := false

	if debug {
		res := []grid{}
		for x := 0; x <gridNoX; x++ {
			for y := 0; y < gridNoY; y++ {
				res = append(res, grid{x:x, y:y})
			}
		}
		return res
	}

	res := []grid{}

	r := 0.5 * math.Sqrt(constants.VehicleLength*constants.VehicleLength+constants.VehicleWidth*constants.VehicleWidth)

	l := math.Atan(constants.VehicleWidth / constants.VehicleLength)

	leftFront := -alpha + l
	leftRear := -alpha + l - math.Pi
	rightFront := -alpha - l
	rightRear := -alpha - l - math.Pi

	// w metrach
	leftFrontX := x + r*math.Cos(math.Pi-leftFront)
	leftFrontY := y - r*math.Cos(math.Pi/2-leftFront)

	rightFrontX := x + r*math.Cos(math.Pi-rightFront)
	rightFrontY := y - r*math.Cos(math.Pi/2-rightFront)

	rightRearX := x + r*math.Cos(math.Pi-rightRear)
	rightRearY := y - r*math.Cos(math.Pi/2-rightRear)

	leftRearX := x + r*math.Cos(math.Pi-leftRear)
	leftRearY := y - r*math.Cos(math.Pi/2-leftRear)

	const marginMeters = 0.0
	margin := int(math.Floor(marginMeters / gridSize)) // [m] -> pixels
	drawRectange := func(x1, y1, x2, y2, x3, y3, x4, y4 int) {

		minx := min(x1, x2, x3, x4) - margin
		maxx := max(x1, x2, x3, x4) + margin

		miny := min(y1, y2, y3, y4) - margin
		maxy := max(y1, y2, y3, y4) + margin

		if minx == 0 || miny == 0 {
			return
		}

		for x := minx; x <= maxx; x++ {
			for y := miny; y <= maxy; y++ {
				if x < 0 || y < 0 || x >= gridNoX || y >= gridNoY {
					continue
				}
				res = append(res, grid{x: x, y: y})
			}
		}
	}

	x1, y1 := xToGridX(leftFrontX - conflictZoneMinX), yToGridY(leftFrontY - conflictZoneMinY)
	x2, y2 := xToGridX(rightFrontX - conflictZoneMinX), yToGridY(rightFrontY - conflictZoneMinY)
	x3, y3 := xToGridX(rightRearX - conflictZoneMinX), yToGridY(rightRearY - conflictZoneMinY)
	x4, y4 := xToGridX(leftRearX - conflictZoneMinX), yToGridY(leftRearY - conflictZoneMinY)

	drawRectange(x1, y1, x2, y2, x3, y3, x4, y4)

	return res
}

func getAlpha(xFrom types.XCoord, yFrom types.YCoord, xTo types.XCoord, yTo types.YCoord) types.Angle {
	xDiff := xFrom - xTo
	yDiff := yFrom - yTo

	switch {
	case xDiff == 0 && yDiff > 0:
		return -math.Pi / 2
	case xDiff == 0 && yDiff < 0:
		return math.Pi / 2
	case yDiff == 0 && xDiff > 0:
		return 0.0
	case yDiff == 0 && xDiff < 0:
		return math.Pi
	default:
		return math.Atan(-yDiff / xDiff)
	}
}
