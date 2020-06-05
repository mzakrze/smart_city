package util

import (
	"io/ioutil"
	"gopkg.in/yaml.v2"
)

type Configuration struct {
	VehiclesPerMinute             int
	IntersectionPolicy            string
	SimulationDuration            int
	SimulationWarmUp			  int
	RandomSeed                    int64
	VehiclePower                  float64
	VehicleBrakingForce           float64
	VehicleWeight                 float64
	VehicleMaxAngularSpeed        float64
	VehicleMaxSpeedOnConflictZone float64
	DsrcMsgLossProbability		  float64
	DsrcMsgAvgDelay				  int
	MapLanes				      int
}

type config struct {
	Simulation struct {
		Vrp 		int `yaml:"vehicles_per_minute"`
		Ip 			string `yaml:"intersection_policy"`
		Duration	int `yaml:"duration"`
		WarmUp 		int `yaml:"warmup"`
	} `yaml:"simulation"`
	Map struct {
		Type 	string `yaml:"type"`
		Lanes 	int `yaml:"lanes"`
	} `yaml:"map"`
	Vehicle struct {
		Weight 					int `yaml:"weight"`
		Power 					int `yaml:"power"`
		BrakingForce 			int `yaml:"braking_force"`
		MaxAngularSpeed 		float64 `yaml:"max_angular_speed"`
		MaxSpeedOnConflictZone 	float64 `yaml:"max_speed_on_conflict_zone"`
	} `yaml:"vehicle"`
	Dsrc struct {
		LossP 					float64 `yaml:"loss_p"`
		AvgLatency 				int `yaml:"avg_latency"`
	} `yaml:"dsrc"`
	RandomSeed					int64 `yaml:"random_seed"`
}

func ReadConfiguration(filePath string) (Configuration, error) {
	data, err := ioutil.ReadFile(filePath)

	if err != nil {
		panic(err)
	}

	var cfg config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		panic(err)
	}

	if cfg.Simulation.Ip == "trafficLights" {
		if cfg.Dsrc.AvgLatency != 0 {
			panic("Illegal dsrc.avg_latency for traffic lights (must be 0)")
		}
		if cfg.Dsrc.LossP != 0 {
			panic("Illegal dsrc.loss_p for traffic lights (must be 0)")
		}
	}

	return Configuration{
		VehiclesPerMinute: cfg.Simulation.Vrp,
		IntersectionPolicy: cfg.Simulation.Ip,
		SimulationDuration: cfg.Simulation.Duration,
		SimulationWarmUp: cfg.Simulation.WarmUp,
		RandomSeed: cfg.RandomSeed,
		VehiclePower: float64(cfg.Vehicle.Power),
		VehicleBrakingForce: float64(cfg.Vehicle.BrakingForce),
		VehicleWeight: float64(cfg.Vehicle.Weight),
		VehicleMaxAngularSpeed: float64(cfg.Vehicle.MaxAngularSpeed),
		VehicleMaxSpeedOnConflictZone: float64(cfg.Vehicle.MaxSpeedOnConflictZone),
		DsrcMsgLossProbability: cfg.Dsrc.LossP,
		DsrcMsgAvgDelay: cfg.Dsrc.AvgLatency,
		MapLanes: cfg.Map.Lanes,
	}, nil
}
