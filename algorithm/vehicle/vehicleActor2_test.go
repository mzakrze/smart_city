package vehicle

import (
	"algorithm/constants"
	"algorithm/types"
	"math"
	"math/rand"
	"testing"
)

func getPlanMockKeepVelocity(v types.MetersPerSecond) map[types.Millisecond]types.MetersPerSecond {
	plan := make(map[types.Millisecond]types.MetersPerSecond)
	for ts := types.Millisecond(0); ts < 10000; ts += constants.SimulationStepInterval {
		plan[ts] = v
	}
	return plan
}


func getPlanMockAccelerate(v0, v2 types.MetersPerSecond) map[types.Millisecond]types.MetersPerSecond {
	plan := make(map[types.Millisecond]types.MetersPerSecond)

	v := v0
	for ts := types.Millisecond(0); v < v2; ts += constants.SimulationStepInterval {
		v += maxAcc * constants.SimulationStepIntervalSeconds
		plan[ts] = v
	}

	return plan
}

//func TestApproachConflictZone(t *testing.T) {
//
//	// given
//	s := 100.0
//	vMax := 10.0
//	plan := getPlanMockKeepVelocity(10)
//
//	// when
//	v, s1 := calculateDistSpeedAfter(plan, 500)
//	s -= s1
//	t, v2 := calculateApproachConflictZoneTimeSpeed(v, s, vMax)
//
//
//
//
//	// then
//
//}



func TestCalculateDistSpeedAfter(t *testing.T) {
	const maxErr = 0.00001
	testCases := []struct{
		plan map[types.Millisecond]types.MetersPerSecond
		v types.MetersPerSecond
		s types.Meter
	}{
		{getPlanMockKeepVelocity(10), 10, 5},
		{getPlanMockKeepVelocity(5), 5, 2.5},
		{getPlanMockKeepVelocity(20), 20, 10},
		{getPlanMockAccelerate(0, 10), 1, 0.255},
		{getPlanMockAccelerate(10, 20), 11, 5.255},
	}

	for _, table := range testCases {
		v, s := calculateDistSpeedAfter(table.plan, 0, 500)
		if math.Abs(v - table.v) > maxErr {
			t.Errorf("Velocity miscalculated: actual = %v, and Expected = %v.", v, table.v)
		}
		_ = s
		if math.Abs(s - table.s) > maxErr {
			t.Errorf("Distance miscalculated: actual = %v, and Expected = %v.", s, table.s)
		}
	}
}

func TestCalculateApproachConflictZoneTimeSpeedUnit(t *testing.T) {

	testCases := []struct{
		v0 float64
		s float64
		v2 float64
		t types.Millisecond
		v float64
	}{
		{0, 100, 10e4, 10000, 20},
		{0, 1, 10e4, 1000, 2},
		{8.059999999999999, 37.0086, 15 , 3267, 14.594451000294598},
	}

	for _, test := range testCases {
		tt, v, ok := calculateApproachConflictZoneTimeSpeed(test.v0, test.s, test.v2)
		if ok == false {
			panic("Oops")
		}
		if test.t != tt {
			t.Errorf("Time miscalculated, expected: %v, actual: %v", test.t, tt)
		}

		if test.v != v {
			t.Errorf("Speed miscalculated, expected: %v, actual: %v", test.v, v)
		}

		_, success := calculatePlanEnterConflictZone(test.v, v, test.s, float64(test.t) / 1000.0)

		if success == false {
			t.Errorf("Vehcile cannot fulfill reservation it asked for")
		}

	}

}

func TestIntegrationCalculateMakeAndFulfillReservation(t *testing.T) {
	// jeśli pojazd poprosi o rezerwację, i ta zostanie zaakceptowana w czasie do 500 ms -> musi być w stanie ją zrealizować

	plan := getPlanMockKeepVelocity(10)

	testCases := 10000
	skipTestCounter := 0
	for i := 0; i < testCases; i++ {
		s := rand.Float64() * 50.0

		v0, s1 := calculateDistSpeedAfter(plan, 0, 500)
		if s - s1 < 0 {
			skipTestCounter += 1
			continue
		}
		vMax := maxSpeedOnConflictZone
		ts, v2, ok := calculateApproachConflictZoneTimeSpeed(v0, s-s1, vMax)

		if ok == false {
			skipTestCounter += 1
			continue
		}

		_, success := calculatePlanEnterConflictZone(v0, v2, s, float64(500 + ts) / 1000.0)

		if success == false {
			t.Errorf("Vehcile cannot fulfill reservation it asked for")
		}

		v1, s1 := calculateDistSpeedAfter(plan, 0, 200)

		_, success = calculatePlanEnterConflictZone(v1, v2, s - s1, float64(300 + ts) / 1000.0)

		if success == false {
			t.Errorf("Vehcile cannot fulfill reservation it asked for")
		}

	}

	t.Logf("Tested %d, out of total %d", (testCases - skipTestCounter), testCases)
}


func TestCalculatePlanEnterConflictZone(t *testing.T) {
	testCases := 1000
	skipTestCounter := 0
	for i := 0; i < testCases; i += 1 {
		v0 := rand.Float64() * 10
		v2 := rand.Float64() * 10
		s := rand.Float64() * 50
		t0 := rand.Float64() * 25

		plan, success := calculatePlanEnterConflictZone(v0, v2, s, t0)

		if success == false {
			skipTestCounter += 1
			continue
		}

		for _, v := range plan {
			if v < 0.0 {
				t.Errorf("Illegal - negative speed")
			}
		}

		prevV := -1.0
		trueS := 0.0
		for ts := types.Millisecond(0); ts < types.Millisecond(t0 * 1000) ; ts += constants.SimulationStepInterval {

			v := plan[ts]

			if ts != 0 && prevV < 0 {
				if v > prevV {
					if v - prevV > maxAcc * constants.SimulationStepIntervalSeconds {
						t.Errorf("Unrealistic plan: to much acceleration - from %v to %v in single simulation step", prevV, v)
					}
				}
				if v < prevV {
					if prevV - v > maxDecel * constants.SimulationStepIntervalSeconds {
						t.Errorf("Unrealistic plan: to much deceleration - from %v to %v in single simulation step", prevV, v)
					}
				}
			}

			trueS += v * constants.SimulationStepIntervalSeconds

			prevV = v
		}
		trueV := prevV

		const maxErr = 0.1
		if math.Abs(v2 - trueV) > maxErr {
			t.Errorf("Incorrect speed on conflict zone, expcted %v, actual %v", v2, trueV)
		}

		if math.Abs(s - trueS) > maxErr {
			t.Errorf("Incorrect distance, expcted %v, actual %v", s, trueS)
		}
	}

	t.Logf("Tested %d, out of total %d", (testCases - skipTestCounter), testCases)
}

func TestCalculatePlanEnterConflictZoneUnit(t *testing.T) {

	testCases := []struct{
		v0 float64
		v2 float64
		s float64
		t0 float64
		success bool
	}{
		//{2.0318687664732287, 3.60871416856906, 28.53366380355113, 21.56228593619716, true},
		//{3.746416013780387, 8.911528829443832, 24.44058645766909, 22.752579401258487, true}, // TODO
		{8.0, 12.153781304598171, 40.84, 6.076, true},
	}

	for _, table := range testCases {
		plan, success := calculatePlanEnterConflictZone(table.v0, table.v2, table.s, table.t0)
		if success != table.success {
			t.Fail()
			continue
		}
		if success == false {
			continue
		}

		prevV := -1.0
		trueS := 0.0
		for ts := types.Millisecond(0); ts < types.Millisecond(table.t0 * 1000) ; ts += constants.SimulationStepInterval {
			v := plan[ts]
			trueS += v * constants.SimulationStepIntervalSeconds
			prevV = v
		}
		trueV := prevV

		const maxErr = 0.1
		if math.Abs(table.v2 - trueV) > maxErr {
			t.Errorf("Incorrect speed on conflict zone, expcted %v, actual %v", table.v2, trueV)
		}

		if math.Abs(table.s - trueS) > maxErr {
			t.Errorf("Incorrect distance, expcted %v, actual %v", table.s, trueS)
		}
	}
}

