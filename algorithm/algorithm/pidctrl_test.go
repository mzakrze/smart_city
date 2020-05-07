package algorithm

import (
	"fmt"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"math"
	"math/rand"
	"mzakrze/smart_city/types"
	"testing"
	"time"
)

var tests = []struct {
	p       float64
	i       float64
	d       float64
	min     float64
	max     float64
	updates []*testUpdate
}{
	// p-only controller
	{
		p: 0.5,
		updates: []*testUpdate{
			{setpoint: 10, input: 5, output: 2.5},
			{input: 10, output: 0},
			{input: 15, output: -2.5},
			{input: 100, output: -45},
			{setpoint: 1, input: 0, output: 0.5},
		},
	},
	// i-only controller
	{
		i: 0.5,
		updates: []*testUpdate{
			{setpoint: 10, input: 5, output: 2.5},
			{input: 5, output: 5},
			{input: 5, output: 7.5},
			{input: 15, output: 5},
			{input: 20, output: 0},
		},
	},
	// d-only controller
	{
		d: 0.5,
		updates: []*testUpdate{
			{setpoint: 10, input: 5, output: -2.5},
			{input: 5, output: 0},
			{input: 10, output: -2.5},
		},
	},
	// pid controller
	{
		p: 0.5,
		i: 0.5,
		d: 0.5,
		updates: []*testUpdate{
			{setpoint: 10, input: 5, output: 2.5},
			{input: 10, output: 0},
			{input: 15, output: -5},
			{input: 100, output: -132.5},
			{setpoint: 1, input: 0, output: 6},
		},
	},
	// Thermostat example
	{
		p:   0.6,
		i:   1.2,
		d:   0.075,
		max: 1, // on or off
		updates: []*testUpdate{
			{setpoint: 72, input: 50, output: 1},
			{input: 51, output: 1},
			{input: 55, output: 1},
			{input: 60, output: 1},
			{input: 75, output: 0},
			{input: 76, output: 0},
			{input: 74, output: 0},
			{input: 72, output: 0.15},
			{input: 71, output: 1},
		},
	},
	// pd controller, demonstrating we need to prevent i windup before summing up output
	{
		p:   40.0,
		i:   0,
		d:   12.0,
		max: 255, min: 0,
		updates: []*testUpdate{
			{setpoint: 90.0, input: 22.00, output: 255.00},
			{input: 25.29, output: 255.00},
			{input: 28.56, output: 255.00},
			{input: 31.80, output: 255.00},
			{input: 35.02, output: 255.00},
			{input: 38.21, output: 255.00},
			{input: 41.38, output: 255.00},
			{input: 44.53, output: 255.00},
			{input: 47.66, output: 255.00},
			{input: 50.76, output: 255.00},
			{input: 53.84, output: 255.00},
			{input: 56.90, output: 255.00},
			{input: 59.93, output: 255.00},
			{input: 62.95, output: 255.00},
			{input: 65.94, output: 255.00},
			{input: 68.91, output: 255.00},
			{input: 71.85, output: 255.00},
			{input: 74.78, output: 255.00},
			{input: 77.69, output: 255.00},
			{input: 80.57, output: 255.00},
			{input: 83.43, output: 228.48},
			{input: 85.93, output: 132.80},
			{input: 87.18, output: 97.80},
			{input: 87.96, output: 72.24},
			{input: 88.41, output: 58.20},
			{input: 88.68, output: 49.56},
			{input: 88.83, output: 45.00},
			{input: 88.92, output: 42.12},
			{input: 88.98, output: 40.08},
			{input: 89.00, output: 39.76},
			{input: 89.03, output: 38.44},
			{input: 89.03, output: 38.80},
			{input: 89.05, output: 37.76},
			{input: 89.04, output: 38.52},
			{input: 89.05, output: 37.88},
			{input: 89.05, output: 38.00},
			{input: 89.05, output: 38.00},
			{input: 89.05, output: 38.00},
			{input: 89.05, output: 38.00},
		},
	},
}

type testUpdate struct {
	setpoint float64
	input    float64
	output   float64
}

func round(v float64, decimals int) float64 {
	var pow float64 = 1
	for i := 0; i < decimals; i++ {
		pow *= 10
	}
	return float64(int((v*pow)+0.5)) / pow
}

func (u *testUpdate) check(c *PIDController, ts types.Milisecond) error {
	if u.setpoint != 0 {
		c.Set(u.setpoint, 0)
	}
	output := c.Feedback(u.input, ts)
	if round(output, 63) != round(u.output, 63) {
		return fmt.Errorf("Bad output: %f != %f (%#v)", output, u.output, u)
	}
	return nil
}

func TestUpdate_p(t *testing.T) {
	ts := 0
	for i, test := range tests {
		t.Logf("-- test #%d", i+1)
		c := newPIDController(test.p, test.i, test.d)
		if test.min != 0 || test.max != 0 {
			c.SetOutputLimits(test.min, test.max)
		}
		ts = 0
		for _, u := range test.updates {
			ts += 1
			if err := u.check(c, ts); err != nil {
				t.Error(err)
			}
		}
	}
}

func TestApproachConflictZonePidController(t *testing.T) {
	const setpoint = 0.0

	controller := NewApproachConflictZonePidController()
	controller.SetOutputLimits(-2.0, 3.5)
	controller.Set(setpoint, 0)
	vehicle := vehicleMock{speed: 10, position: 50}

	controller2 := NewApproachConflictZonePidController()
	controller2.SetOutputLimits(-2.0, 3.5)
	controller2.Set(setpoint, 0)
	vehicle2 := vehicleMock{speed: 10, position: 27}

	x, y, score := evaluatePidController(vehicle, controller)
	x2, y2, score2 := evaluatePidController(vehicle2, controller2)

	fmt.Printf("Best pid (on distance 50) scored: %f (position = %f)\n", score, vehicle.position)
	fmt.Printf("Best pid (on distance 27) scored: %f (position = %f)\n", score2, vehicle2.position)

	plotPidResult(controller.p, controller.i, controller.d, x, y, setpoint, "res1")
	plotPidResult(controller2.p, controller2.i, controller2.d, x2, y2, setpoint, "res2")
}

type vehicleMock struct {
	speed float64
	position float64
}

func (v *vehicleMock) accelerate(acc float64, dt types.Milisecond) {
	v.speed += acc * float64(dt) / 1000.0
}

func (v *vehicleMock) move(dt types.Milisecond) {
	v.position -= v.speed * float64(dt) / 1000
}

func evaluatePidController(v vehicleMock, controller *PIDController) ([]float64, []float64, float64) {
	result_x := []float64{}
	result_y := []float64{}
	result_score := 0.0

	const stepDuration types.Milisecond = 100
	vehicleStartingPosition := v.position

	// scores:
	const veryBadScore = -10000000
	const hasntReachedDestScore = veryBadScore
	const gotToConflictZoneScore = hasntReachedDestScore * 10
	const startedToRevert = hasntReachedDestScore * 100

	ts := 0
	for n := range make([]struct{}, 3000) {
		ts += stepDuration

		// ask the pid for acceleration
		acc := controller.Feedback(v.position, ts)

		// apply deceleration / deceleration
		v.accelerate(acc, stepDuration)

		// move the vehicle
		v.move(stepDuration)

		// score
		if v.position < 0 {
			result_score = gotToConflictZoneScore // very bad score
			break
		}

		result_score -= acc * acc
		if v.position < 5 {
			result_score -= v.position
		}

		result_x = append(result_x, float64(ts))
		result_y = append(result_y, v.position)

		_ = n
	}

	if v.position > 1 {
		result_score = hasntReachedDestScore
	}
	if v.position > vehicleStartingPosition {
		result_score = startedToRevert
	}

	return result_x, result_y, result_score
}

// not really a test, but its easier to run it this way
func TestFindBestApproachConflictZonePidCoeficients(t *testing.T) {
	const max_p = 10
	const max_i = 10
	const max_d = 10
	const step = 0.1

	best_p := 0.0
	best_i := 0.0
	best_d := 0.0
	bestPidScore1 := math.MinInt32 * 1.0
	bestPidScore2 := math.MinInt32 * 1.0

	stepsTotal := float64( max_p / step)
	stepsDone := 0.0

	go func() {
		prevPercentDone :=  -1
		for x := range time.Tick(time.Second) {
			p := int(100 * stepsDone / stepsTotal)

			if p != prevPercentDone {
				fmt.Printf("Done: %d%%\n", p)
			}

			_ = x
		}
	}()

	for p := 0.0; p < max_p; p += step {
		for i := 0.0; i < max_i; i += step {
			for d := 0.0; d < max_d; d += step {

				// arrive smoothly
				controller := newPIDController(p, i, d)
				controller.SetOutputLimits(-2.0, 3.5)
				controller.Set(0, 0)
				v := vehicleMock{speed: 10, position: 50}
				_, _ , score1 := evaluatePidController(v, controller)

				// and also - break hard on short distances
				controller2 := newPIDController(p, i, d)
				controller2.SetOutputLimits(-2.0, 3.5)
				controller2.Set(0, 0)
				v2 := vehicleMock{speed: 10, position: 27}
				_, _ , score2 := evaluatePidController(v2, controller2)

				if score1 + score2 > bestPidScore1 + bestPidScore2 {
					best_p = p
					best_i = i
					best_d = d
					bestPidScore1 = score1
					bestPidScore2 = score2
				}

			}
		}

		stepsDone += 1
	}

	fmt.Printf("Best pid = %f, %f, %f, with score: %f, %f\n", best_p, best_i, best_d, bestPidScore1, bestPidScore2)

}


func plotPidResult(p_coef, i_coef, d_coef float64, x, y []float64, setpoint float64, filename string) {

	//randomPoints := func (n int) plotter.XYs {
	//	pts := make(plotter.XYs, n)
	//	for i := range pts {
	//	if i == 0 {
	//	pts[i].X = rand.Float64()
	//} else {
	//	pts[i].X = pts[i-1].X + rand.Float64()
	//}
	//	pts[i].Y = pts[i].X + 10*rand.Float64()
	//}
	//	return pts
	//}

	getPlotterXy := func (x, y []float64) plotter.XYs {
		pts := make(plotter.XYs, len(x))
		for i := range pts {
			pts[i].X = x[i]
			pts[i].Y = y[i]
		}
		return pts
	}


	rand.Seed(int64(0))

	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	p.Title.Text = fmt.Sprintf("Pid results %f, %f, %f", p_coef, i_coef, d_coef)
	p.X.Label.Text = "time"
	p.Y.Label.Text = "meters"

	err = plotutil.AddLinePoints(p,
		"position", getPlotterXy(x, y))
	if err != nil {
		panic(err)
	}

	// Save the plot to a PNG file.
	if err := p.Save(4*vg.Inch, 4*vg.Inch, filename + ".png"); err != nil {
		panic(err)
	}


}