package algorithm

// // https://github.com/felixge/pidctrl/blob/master/pidctrl.go

import (
	"math"
	"mzakrze/smart_city/types"
)

const approachConflictZonePid_P = 0.6
const approachConflictZonePid_I = 0.0
const approachConflictZonePid_D = 1.2


func NewApproachConflictZonePidController() *PIDController {
	return newPIDController(
		approachConflictZonePid_P,
		approachConflictZonePid_I,
		approachConflictZonePid_D);
}

func newPIDController(p, i, d float64) *PIDController {
	return &PIDController{p: p, i: i, d: d, outMin: math.Inf(-1), outMax: math.Inf(0), prevValue: []float64{}, lastUpdate: -1}
}

const prevValueWindowSize = 10
type PIDController struct {
	p          float64   // proportional gain
	i          float64   // integral gain
	d          float64   // derrivate gain
	setpoint   float64   // current setpoint
	//prevValue  float64   // last process value
	prevValue  []float64
	prevValueIndex int
	lastUpdate types.Milisecond // time of last update
	outMin     float64   // Output Min
	outMax     float64   // Output Max
}

func (c *PIDController) Set(setpoint float64, ts types.Milisecond) {
	c.setpoint = setpoint
	c.lastUpdate = ts
}

func (c *PIDController) SetOutputLimits(min, max float64) {
	if min > max {
		panic("Illegal argument")
	}
	c.outMin = min
	c.outMax = max

}

func (c *PIDController) Feedback(value float64, ts types.Milisecond) float64 {
	if c.lastUpdate < 0 {
		panic("PID controller hasnt been set before use")
	}
	var (
		dt  = float64(ts - c.lastUpdate) / 1000
		err = c.setpoint - value
		d   float64
	)
	c.lastUpdate = ts

	if c.prevValueIndex == prevValueWindowSize - 1 {
		c.prevValue = append(c.prevValue[ 1: prevValueWindowSize], value)
	} else {
		c.prevValue = append(c.prevValue, value)
		c.prevValueIndex += 1
	}

	integral_sum := 0.0
	for _, v := range c.prevValue[ : c.prevValueIndex] {
		integral_sum += v
	}

	if c.prevValueIndex > 2 {
		d = -((value - c.prevValue[c.prevValueIndex - 2]) / dt)
	} else {
		d = 0
	}

	output := (c.p * err) + c.i * integral_sum + (c.d * d)

	if output > c.outMax {
		output = c.outMax
	} else if output < c.outMin {
		output = c.outMin
	}

	return -output
}
