package util

import (
	"algorithm2.0/types"
	"encoding/json"
	"math"

	"io/ioutil"
	"net/http"
)

type Graph struct {
	MapWidth     types.Meter
	MapHeight    types.Meter
	ConflictZone ConflictZone
	ImX			 types.XCoord
	ImY			 types.YCoord
	AllNodes        []Node
	Entrypoints	 []*Node
	Exitpoints	 []*Node
}

type ConflictZone struct {
	MinX types.XCoord
	MaxX types.XCoord
	MinY types.YCoord
	MaxY types.YCoord
}

type Node struct {
	Id types.NodeId
	X types.XCoord
	Y types.YCoord
	EdgesFrom []*Edge
	IsEntryPoint bool
	IsExitPoint bool
	EntryPointId int
	ExitPointId	int
	WayId int
}

type Edge struct {
	Id types.EdgeId
	From *Node
	To *Node
	IsArc bool
	Length types.Meter
}

type graphRaw struct {
	Nodes []nodeRaw `json:"nodes"`
	Edges []edgeRaw `json:"edges"`
	MapWidth int `json:"mapWidth"`
	MapHeight int `json:"mapHeight"`
	ConflictZone conflictZoneRaw `json:"conflictZone"`
}

type nodeRaw struct {
	Id int
	X float64
	Y float64
	IsEntryPoint bool
	IsExitPoint bool
	EntryPointId int
	ExitPointId int
	WayId int
}

type edgeRaw struct {
	From int
	To int
	Arc bool
}

type conflictZoneRaw struct {
	MinX float64
	MaxX float64
	MinY float64
	MaxY float64
}

func ReadGraph(simName string) (*Graph, error) {
	client := &http.Client{}
	resp, err := client.Get("http://localhost:9200/simulation-info/_doc/" + simName)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	respObject := struct {
		Source struct {
			GraphRaw string `json:"graph_raw"`
		} `json:"_source"`
	}{}
	err = json.Unmarshal([]byte(body), &respObject)
	if err != nil {
		return nil, err
	}

	graphObj := graphRaw{}
	err = json.Unmarshal([]byte(respObject.Source.GraphRaw), &graphObj)
	if err != nil {
		return nil, err
	}

	graph := Graph{}
	graph.ConflictZone = ConflictZone{
		MinX: graphObj.ConflictZone.MinX,
		MaxX: graphObj.ConflictZone.MaxX,
		MinY: graphObj.ConflictZone.MinY,
		MaxY: graphObj.ConflictZone.MaxY,
	}
	graph.Entrypoints = []*Node{}
	graph.Exitpoints = []*Node{}

	graph.ImX = 0 // TODO
	graph.ImY = 0

	graph.MapHeight = types.Meter(graphObj.MapHeight)
	graph.MapWidth = types.Meter(graphObj.MapWidth)

	idToPointer := make(map[types.VehicleId]*Node)

	graph.AllNodes = make([]Node, len(graphObj.Nodes))
	for i, n := range graphObj.Nodes {
		nn := Node{
			Id: types.NodeId(n.Id),
			X: types.XCoord(n.X),
			Y: types.YCoord(n.Y),
			IsEntryPoint: n.IsEntryPoint,
			IsExitPoint: n.IsExitPoint,
			EntryPointId: n.EntryPointId,
			ExitPointId: n.ExitPointId,
			WayId: n.WayId,
			EdgesFrom: []*Edge{},
		}

		graph.AllNodes[i] = nn
		idToPointer[nn.Id] = &nn
		if nn.IsEntryPoint {
			graph.Entrypoints = append(graph.Entrypoints, &nn)
		}
		if nn.IsExitPoint{
			graph.Exitpoints = append(graph.Exitpoints, &nn)
		}

	}

	eId := types.EdgeId(0)
	for _, e := range graphObj.Edges {
		f := idToPointer[types.NodeId(e.From)]
		t := idToPointer[types.NodeId(e.To)]
		x := f.X - t.X
		y := f.Y - t.Y
		f.EdgesFrom = append(f.EdgesFrom, &Edge{
			Id: types.EdgeId(eId),
			From: f,
			To: t,
			IsArc: e.Arc,
			Length: math.Sqrt(x * x + y * y),
		})
		eId += 1
	}

	return &graph, nil
}
