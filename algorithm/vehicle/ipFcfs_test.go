package vehicle

import (
	"algorithm/util"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"
)

func TestReservationTable(t *testing.T) {

	graph := &util.Graph{
		ConflictZone: util.ConflictZone{
			MinX: 0,
			MaxX: 40,
			MinY: 0,
			MaxY: 40,
		},
	}

	fcfs := CreateIntersectionPolicyFcfs(graph, util.Configuration{})

	for counter := 0; counter < 100; counter++ {

		colors := []color.Color{color.Black, color.White}
		rect := image.Rect(0, 0, gridNoX, gridNoY)
		newFrame := image.NewPaletted(rect, colors)
		for x := range fcfs.reservationTable[counter] {
			for y := range fcfs.reservationTable[counter][x] {
				if fcfs.reservationTable[counter][x][y] == taken {
					if x  < 0 || y < 0 {
						continue
					}
					if x >= gridNoX || y >= gridNoY {
						continue
					}
					newFrame.Set(x, gridNoY - y , color.Black)
				} else {
					newFrame.Set(x, gridNoY - y, color.White)
				}
			}
		}

		f, _ := os.Create(fmt.Sprintf("img/image%05d.png", counter))
		png.Encode(f, newFrame)
	}
}
