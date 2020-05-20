package vehicle

import (
	"algorithm2.0/constants"
	"algorithm2.0/types"
	"algorithm2.0/util"
	"math"
)

type IntersectionPolicyFcfs struct {
	reservationTable [][]bool
	replies          []*DsrcR2VMessage
	gridNoX          int
	gridNoY          int
}
const gridSize types.Meter = 0.1

func CreateIntersectionPolicyFcfs(graph *util.Graph) *IntersectionPolicyFcfs {
	gridNoX := int((graph.ConflictZone.MaxX - graph.ConflictZone.MinX) / gridSize)
	gridNoY := int((graph.ConflictZone.MaxY - graph.ConflictZone.MinY) / gridSize)

	grid := make([][]bool, gridNoX)
	for x := range grid {
		grid[x] = make([]bool, gridNoY)
	}

	return &IntersectionPolicyFcfs{
		reservationTable: grid,
		gridNoX: gridNoX,
		gridNoY: gridNoY,
		replies: []*DsrcR2VMessage{},
	}
}

func (ip * IntersectionPolicyFcfs) ProcessMsg(m DsrcV2RMessage) {

}

func (ip *IntersectionPolicyFcfs) GetReplies(ts types.Millisecond) []*DsrcR2VMessage {
	res := ip.replies
	ip.replies = []*DsrcR2VMessage{}
	return res
}

func (ip *IntersectionPolicyFcfs) appendToReservationTable(v *VehicleActor) {
	r := 0.5 * math.Sqrt(constants.VehicleLength*constants.VehicleLength+constants.VehicleWidth*constants.VehicleWidth)

	l := math.Atan(constants.VehicleWidth / constants.VehicleLength)

	leftFront := -v.Alpha + l
	leftRear := -v.Alpha + l - math.Pi
	rightFront := -v.Alpha - l
	rightRear := -v.Alpha - l - math.Pi

	// w metrach
	leftFrontX := v.X + r*math.Cos(math.Pi-leftFront)
	leftFrontY := v.Y - r*math.Cos(math.Pi/2-leftFront)

	rightFrontX := v.X + r*math.Cos(math.Pi-rightFront)
	rightFrontY := v.Y - r*math.Cos(math.Pi/2-rightFront)

	rightRearX := v.X + r*math.Cos(math.Pi-rightRear)
	rightRearY := v.Y - r*math.Cos(math.Pi/2-rightRear)

	leftRearX := v.X + r*math.Cos(math.Pi-leftRear)
	leftRearY := v.Y - r*math.Cos(math.Pi/2-leftRear)

	drawRectange := func(x1 int, y1 int, x2 int, y2 int, x3 int, y3 int, x4 int, y4 int) {

		minx := min(x1, x2, x3, x4)
		maxx := max(x1, x2, x3, x4)

		miny := min(y1, y2, y3, y4)
		maxy := max(y1, y2, y3, y4)

		if minx == 0 || miny == 0 {
			return
		}

		for x := minx; x < maxx; x++ {
			for y := miny; y < maxy; y++ {
				ip.reservationTable[x][y] = true
			}
		}
	}

	drawPixel := func (x, y float64) (int, int) {
		pixelX := int(x / gridSize)
		pixelY := int(y / gridSize)
		if pixelX < 0 || pixelY < 0 {
			return 0,0
		}
		if pixelX >= ip.gridNoX || pixelY >= ip.gridNoY {
			return 0,0
		}
		ip.reservationTable[pixelX][pixelY] = true
		return pixelX, pixelY
	}

	x1, y1 := drawPixel(leftFrontX, leftFrontY)
	x2, y2 := drawPixel(rightFrontX, rightFrontY)
	x3, y3 := drawPixel(rightRearX, rightRearY)
	x4, y4 := drawPixel(leftRearX, leftRearY)

	drawRectange(x1, y1, x2, y2, x3, y3, x4, y4)

}

func min(ns ...int) int {
	min0 := math.MaxInt32
	for n := range ns {
		if n < min0 {
			min0 = n
		}
	}
	return min0
}

func max(ns ...int) int {
	max0 := math.MinInt32
	for n := range ns {
		if n > max0 {
			max0 = n
		}
	}
	return max0
}