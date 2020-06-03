package vehicle

import (
	"algorithm/constants"
	"algorithm/types"
	"algorithm/util"
	"testing"
)

func TestVehicleRoadCommunictionNoDelay(t *testing.T) {
	// given
	p := AllVehiclesProxySingleton()
	conf := util.Configuration{
		DsrcMsgLossProbability: 0.0,
		DsrcMsgAvgDelay: 0,
	}
	c := CommunicationLayerSingleton(p, conf)

	// when
	c.SendDsrcV2R(getV2R(0))
	c.SendDsrcV2R(getV2R(10))
	c.SendDsrcV2R(getV2R(10))
	c.SendDsrcV2R(getV2R(20))
	c.SendDsrcV2R(getV2R(30))
	c.SendDsrcV2R(getV2R(40))

	// then
	assertEquals(t, 1, len(c.IntersectionManagerReceive(0)))
	assertEquals(t, 2, len(c.IntersectionManagerReceive(10)))
	assertEquals(t, 1, len(c.IntersectionManagerReceive(20)))
	assertEquals(t, 1, len(c.IntersectionManagerReceive(30)))
	assertEquals(t, 1, len(c.IntersectionManagerReceive(40)))
}

func TestVehicleRoadCommunictionSomeDelay(t *testing.T) {
	// given
	const delay = 20
	const tests = 10e3
	p := AllVehiclesProxySingleton()
	conf := util.Configuration{
		DsrcMsgLossProbability: 0.0,
		DsrcMsgAvgDelay: delay,
}
	c := CommunicationLayerSingleton(p, conf)

	for i := types.Millisecond(10); i < tests; i += constants.SimulationStepInterval {
		c.SendDsrcV2R(getV2R(i))
		assertEquals(t, 1, len(c.IntersectionManagerReceive(i + 2 * delay)))
	}

	counter := 0
	for i := types.Millisecond(10); i < tests; i += constants.SimulationStepInterval {
		c.SendDsrcV2R(getV2R(i))
		if 1 == len(c.IntersectionManagerReceive(i)) {
			counter += 1
		}
	}
	if counter > 1.5 * tests / delay {
		t.Error("Delay is not visible enough")
	}
}

func TestVehicleRoadCommunictionSomeLoss(t *testing.T) {
	// given
	const lossP = 15.0
	const sent = 10e5
	p := AllVehiclesProxySingleton()
	conf := util.Configuration{
		DsrcMsgLossProbability: lossP,
		DsrcMsgAvgDelay: 0,
	}
	c := CommunicationLayerSingleton(p, conf)

	for i := 0; i < sent; i++ {
		c.SendDsrcV2R(getV2R(types.Millisecond(i)))
	}

	actual := len(c.IntersectionManagerReceive(10e6))

	if actual < int(0.75 * float64(sent)) {
		t.Error("To much messages lost")
	}
	if actual > int(0.95 * float64(sent)) {
		t.Error("To few messages lost")
	}

}

func getV2R(ts types.Millisecond) DsrcV2RMessage {
	return DsrcV2RMessage{
		TsSent: ts,
	}
}

func assertEquals(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Fail()
	}
}
