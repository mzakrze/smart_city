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

	fcfs := CreateIntersectionPolicyFcfs(nil, util.Configuration{})

	for counter := 0; counter < 100; counter++ {

		//fcfs.appendToReservationTable(nil)

		colors := []color.Color{color.Black, color.White}
		rect := image.Rect(0, 0, fcfs.gridNoX, fcfs.gridNoY)
		newFrame := image.NewPaletted(rect, colors)
		for x := range fcfs.reservationTable {
			for y := range fcfs.reservationTable[x] {
				if fcfs.reservationTable[0][x][y] == taken {
					for xx := -5; xx < 5; xx++ {
						for yy := -5; yy < 5; yy++ {
							if x + xx < 0 || y + yy < 0 {
								continue
							}
							if x + xx >= fcfs.gridNoX || y + yy >= fcfs.gridNoY {
								continue
							}
							newFrame.Set(x + xx, fcfs.gridNoY - (y + yy) , color.Black)
						}
					}
				} else {
					newFrame.Set(x, fcfs.gridNoY - y, color.White)
				}
			}
		}

		f, _ := os.Create(fmt.Sprintf("img/image%05d.png", counter))
		err := png.Encode(f, newFrame)
		if err != nil {
			panic(err)
		}
	}
}