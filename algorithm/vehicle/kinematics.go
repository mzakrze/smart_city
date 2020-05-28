package vehicle

import (
	"algorithm/types"
	"math"
)


func roundTimeResult(res types.Millisecond) types.Millisecond {
	return res - res % 10
}



/**
Pojazd jedzie z prędkością v, zaczyna hamować, a po czasie t zatrzymuje się do zera.
Jaka była wartość hamowania (w m/s^2)?
 */
func brakingDecel(v types.MetersPerSecond, t types.Millisecond) types.MetersPerSecond2 {
	return -float64(v) / (float64(t) / 1000.0)
}


/**
Pojazd jedzie z predkoscia v, zaczyna przyspieszać, maksymalnie do prędkości maxV, z przyspieszeniem a.
W ile czasu pokona dystans s?
 */
func arrivalTimeAccelerating(v0, vMax types.MetersPerSecond, a types.MetersPerSecond2, s types.Meter) types.Millisecond {
	v1 := vAfterAcceleratingOverDistance(v0, a, s)

	if v1 > vMax {
		// przez czas t1 przyspieszamy, przez czas t2 jedziemy jednostajnie

		t1 := (vMax - v0) / a
		t2 := (s * a + 0.5 * v0 * v0 - 0.5 * vMax * vMax) / (vMax * a)
		if t2 < 0.0 { panic("Oops") }

		sCheck := v0 * t1 + 0.5 * a * t1 * t1 + vMax * t2
		if math.Abs(sCheck - s) > 0.01 { panic("Oops") }

		return roundTimeResult(types.Millisecond((t1 + t2) * 1000.0))
	} else {
		// na całym odcinku przyspieszamy
		return roundTimeResult(types.Millisecond((math.Sqrt(v0 * v0 + 2 * s * a) - v0) / a * 1000.0))
	}
}

/**
Pojazd jedzie z predkoscia v, zaczyna przyspieszać, maksymalnie do prędkości maxV, z przyspieszeniem a
W ile czasu (t1 + t2 +t3) pokona dystans s, zakładając, że na końcu odcinka musi mieć predkosc v2 (może hamować z wartością d)?
Ile czasu t1 przyspiesza, ile czasu t2 jedzie jednostajnie, ile czasu t3 hamuje?
 */
func arrivalTimeAcceleratingEnterWithSpeed(v0, vMax types.MetersPerSecond, a, d types.MetersPerSecond2, s types.Meter, v2 types.MetersPerSecond) (types.Millisecond, types.Millisecond, types.Millisecond) {

	if v0 > v2 {
		// musimy zwolnic


	}

	v1 := vAfterAcceleratingOverDistance(v0, a, s)
	if v1 < v2 {
		// caly czas przyspieszamy
		t1 := roundTimeResult(types.Millisecond((math.Sqrt(v0 * v0 + 2 * s * a) - v0) / a * 1000.0))
		return t1, 0, 0
	} else {
		// przez czas t1 przyspieszamy, przez czas t2 (możliwe że t2 = 0) jedziemy jednostajnie, przez czas t3 hamujemy

		t1 := float64((vMax - v0) / a)
		t3 := float64((vMax - v2) / d)

		t2 := float64(s / vMax + (v0 * v0 - vMax * vMax) / (2.0 * vMax * a) + (v2 * v2 - vMax * vMax) / (2.0 * vMax * d))
		if t2 < 0.0 {
			panic("Oops")
		}

		return roundTimeResult(types.Millisecond(t1 * 1000.0)), roundTimeResult(types.Millisecond(t2 * 1000.0)), roundTimeResult(types.Millisecond(t3 * 1000.0))
	}
}


/**
Pojazd jedzie z prędkością v, zaczyna przyspieszać z przyspieszeniem a, na dystansie s.
Do jakiej prędkości się rozpędzi?
 */
func vAfterAcceleratingOverDistance(v0 types.MetersPerSecond, a types.MetersPerSecond2, s types.Meter) types.MetersPerSecond {
	return math.Sqrt(v0 * v0 + 2 * s * a)
}


/**
Pojazd jedzie z predkoscia v, przyspiesza z wartością a (może być ujemna) przez czas t.
Ile drogi s pokona przez ten czas, jaką będzie mieć końcową prędkość?
 */
func distSpeedAfterGivenTime(ts types.Millisecond, v types.MetersPerSecond, a types.MetersPerSecond2, vMax types.MetersPerSecond) (types.Meter, types.MetersPerSecond) {
	t := float64(ts) / 1000 // ms -> s

	delta_v := float64(a) / float64(t)

	if v + delta_v <= vMax {
		s := v * float64(t) + 0.5 * float64(a) * float64(t) * float64(t)
		return s, v + delta_v
	} else {
		// przyspieszamy przez czas t1, potem przez czas t2 jedziemy jednostajnie
		t1 := float64((vMax - v) / a)
		t2 := t - t1
		s := v * float64(t1) + 0.5 * float64(a) * float64(t1) * float64(t1) + vMax * t2
		return s, vMax
	}
}