package util

import "algorithm2.0/types"

type Graph struct {
	MapWidth     types.Meter
	MapHeight    types.Meter
	ConflictZone ConflictZone
	ImX			 types.XCoord
	ImY			 types.YCoord
	Nodes        map[types.NodeId]Node
}

type ConflictZone struct {
	MinX types.XCoord
	MaxX types.XCoord
	MinY types.YCoord
	MaxY types.YCoord
}

type Node struct {
	Id types.NodeId
	IsEntryPoint bool
	IsExitPoint	bool
	WayId types.WayId
	EdgesFrom []Edge
}

type Edge struct {
	Id types.EdgeId
	From *Node
	To *Node
}


func ReadGraph(simName string) (*Graph, error) {
	return &Graph{}, nil
}
