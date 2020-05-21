package vehicle

import (
	"algorithm2.0/constants"
	"algorithm2.0/types"
	"algorithm2.0/util"
	"fmt"
	"math"
)

type GridState int
const (
	free GridState = iota // has to be 0 (default value)
	taken
)

type IntersectionPolicyFcfs struct {
	reservationTable  [][][]GridState
	replies           []*DsrcR2VMessage
	gridNoX           int
	gridNoY           int
	reservations      []*ipFcfsReservation
	nextReservationId types.ReservationId
	graph             *util.Graph
}
const gridSize types.Meter = 0.1 // FIXME

type ipFcfsReservation struct {
	reservationId 	types.ReservationId
	startTs			types.Millisecond
	reservedGrids	[][]grid
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
	gridNoX = int((graph.ConflictZone.MaxX - graph.ConflictZone.MinX) / gridSize)
	gridNoY = int((graph.ConflictZone.MaxY - graph.ConflictZone.MinY) / gridSize)
	conflictZoneMinX = graph.ConflictZone.MinX
	conflictZoneMinY = graph.ConflictZone.MinY

	// times 2 just for optimization (to not deallocate memory multiple times)
	reservationTable :=  make([][][]GridState, int(configuration.SimulationDuration.Seconds() * 2 * 1000 / float64(constants.SimulationStepInterval)))
	//reservationTable :=  make([][][]GridState, 1 * 1)
	for i := 0; i < len(reservationTable); i++ {
		reservationTable[i] = make([][]GridState, gridNoX)
		for x := range reservationTable[i] {
			reservationTable[i][x] = make([]GridState, gridNoY)
		}
	}

	return &IntersectionPolicyFcfs{
		replies: []*DsrcR2VMessage{},
		reservationTable: reservationTable,
		nextReservationId: 1,
		graph: graph,
	}
}

func (ip * IntersectionPolicyFcfs) ProcessMsg(m *DsrcV2RMessage) {
	if m.MsgType == AimProtocolMsgReservationCancelation {
		ip.cancelReservation(m.ReservationToCancelId)
		return
	}

	//fmt.Print("IM processing msg from ", m.Sender, ", result =")

	before := ip.calcOccupied()

	success := ip.makeReservationIfFitsInReservationTable(m)
	if success == false {
		//fmt.Println("Rejected")
		return
	}

	//fmt.Println("Accepted")
	after := ip.calcOccupied()
	fmt.Println("request from ", m.Sender, ", Before:", before, ", after:", after, ", diff: ", (after - before))

}


func (ip *IntersectionPolicyFcfs) GetReplies(ts types.Millisecond) []*DsrcR2VMessage {
	res := ip.replies
	ip.replies = []*DsrcR2VMessage{}
	return res
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

func (ip *IntersectionPolicyFcfs) makeReservationIfFitsInReservationTable(msg *DsrcV2RMessage) bool {
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
		vMax = maxSpeedOnConflictZone
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
		const maxSpeedDiff = float64(maxAcc * float64(constants.SimulationStepInterval) / 1000.0)
		if speed == vMax {
			return
		}
		if speed + maxSpeedDiff >= vMax {
			speed = vMax
		} else if speed + maxSpeedDiff < vMax {
			speed += maxSpeedDiff
		} else {
			panic("Oops")
		}
	}

	reservationFromTs := msg.ApproachConflictZoneMinTs

	// zero state
	if msg.IsTurning {
		vehicleRoute = vehicleRoute[1:]
	}
	x = vehicleRoute[0].X
	y = vehicleRoute[0].Y
	exitX = vehicleRoute[len(vehicleRoute) - 1].X
	exitY = vehicleRoute[len(vehicleRoute) - 1].Y

	//alpha = getAlpha(msg.VehicleX, msg.VehicleY, entry.X, entry.Y)
	alpha = 0 // FIXME
	speed = msg.ApproachConflictZoneSpeed
	currentRouteIndex = 0
	ts = reservationFromTs
	reservationTsToSpeed = make(map[types.Millisecond]types.MetersPerSecond)

	fitsInReservationTable := true
	guard := 0
	maxTs := types.Millisecond(0)
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
			asdf := ip.reservationTable[ts/10]
			state := asdf[grids[i].x][grids[i].y]
			if state == taken {
				fitsInReservationTable = false
				break outer_for_label
			}
		}

		accelerateVehicle()
		d := speed * float64(constants.SimulationStepInterval) / 1000.0
		moveVehicle(d)

		maxTs = ts
		_ = maxTs
	}

	if fitsInReservationTable == false {
		return false
	}

	// vehicle fits in the reservation table - so let's actually reserve it
	ts = reservationFromTs

	for j := range reservedGrids {
		for i := range reservedGrids[j] {
			x := reservedGrids[j][i].x
			y := reservedGrids[j][i].y

			ip.reservationTable[ts/10][x][y] = taken
		}
		ts += constants.SimulationStepInterval
	}

	reservation := &ipFcfsReservation{
		reservationId: ip.nextReservationId,
		startTs: reservationFromTs,
		reservedGrids: reservedGrids,
	}
	ip.nextReservationId += 1

	reply := &DsrcR2VMessage{
		msgType: AimProtocolMsgAllow,
		receiver: msg.Sender,
		reservationFromTs: reservationFromTs,
		reservationToTs: ts,
		reservationDesiredSpeed: msg.ApproachConflictZoneSpeed,
		reservationTsToSpeed: reservationTsToSpeed,
	}

	ip.reservations = append(ip.reservations, reservation)
	ip.replies = append(ip.replies, reply)

	return true
}

func (ip *IntersectionPolicyFcfs) cancelReservation(reservationId types.ReservationId) {
	// FIXME - plz implement me

	// TODO
	// 1 - funkcja (ip *IntersectionPolicyFcfs) cancelReservation(reservationId types.ReservationId)
	// 2 - przyspieszanie i hamowanie za pomocą mocy i masy
	// 3 - usunac niepotrzebny kodzik
	// 4 - simulationrunner - ostroznie z generowaniem nowych (zmieniac predkosc)
	// 5 - afterintersection - sprawdzać czy nie zderzy się
	// 6 - więcej testów

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

func getOccupingGrids(x types.XCoord, y types.YCoord, alpha types.Angle) []grid {
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

	drawRectange := func(x1 int, y1 int, x2 int, y2 int, x3 int, y3 int, x4 int, y4 int) {

		minx := min(x1, x2, x3, x4) - 10
		maxx := max(x1, x2, x3, x4) + 10

		miny := min(y1, y2, y3, y4) - 10
		maxy := max(y1, y2, y3, y4) + 10

		if minx == 0 || miny == 0 {
			return
		}

		for x := minx; x <= maxx; x++ {
			for y := miny; y <= maxy; y++ {
				// FIXME - to troche inaczej trzeba liczyć
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
		panic("Oops")
	}
}
