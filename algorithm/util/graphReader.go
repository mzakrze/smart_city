package util

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"mzakrze/smart_city/types"
	"os"
)

// TODO - pewnie trzeba będzie gdzies wyniesc, ale narazie nie wiem gdzie
const GRAPH_PATH = "../map_visualization/public/"
const NODES_FILE = GRAPH_PATH + "nodes.ndjson"
const EDGES_FILE = GRAPH_PATH + "edges.ndjson"
const GRAPH_BBOX_FILE = GRAPH_PATH + "graph.json"

type node struct {
	Id int32
	X float64
	Y float64
	Lat float64
	Lon float64
}

type edge struct {
	From types.NodeId
	To types.NodeId
	Transitive bool

}


func ReadGraph() *types.Graph {

	nodesRaw := readNodesRaw(NODES_FILE)
	edgesRaw := readEdgesRaw(EDGES_FILE)
	graphBBox := readGraphBBoxRaw(GRAPH_BBOX_FILE)

	return assembleGraph(nodesRaw, edgesRaw, graphBBox)
}

func assembleGraph(nodesRaw []node, edgesRaw []edge, mapBox types.MapBBox)  *types.Graph {

	nodesMap := make(map[types.NodeId]*types.Node)
	nodes := make([]types.Node, len(nodesRaw))

	for i, n := range nodesRaw {
		node := types.Node{
			Id: n.Id,
			X: n.X,
			Y: n.Y,
			Edges: []types.Edge{},
		}

		nodes[i] = node
	}

	for i := range nodes {
		nodesMap[nodes[i].Id] = &nodes[i]
	}

	for _, e := range edgesRaw {
		if e.Transitive {
			continue
		}
		nodeFrom, ok := nodesMap[e.From]; if !ok { panic("Error when parsing graph") }
		if nodeFrom.Id != e.From {
			panic(fmt.Sprintf("Something not ok, wanted: %d, got: %d", e.From, nodeFrom.Id))
		}
		nodeTo, ok := nodesMap[e.To]; if !ok { panic("Error when parsing graph") }
		distance := math.Sqrt((nodeFrom.X - nodeTo.X) * (nodeFrom.X - nodeTo.X) + (nodeFrom.Y - nodeTo.Y) * (nodeFrom.Y - nodeTo.Y))
		edge := types.Edge{To: nodeTo, Distance: distance}
		nodeFrom.Edges = append(nodeFrom.Edges, edge)
	}

	return &types.Graph{
		MapBBox: mapBox,
		AllNodes: nodes,
		IntersectionManagers: make([]types.IntersectionManager, 0), // FIXME - python musi generować IM + i dodac w Go czytanie
	}
}

func readNodesRaw(filepath string) []node {
	var nodes = make([]node, 0)

	readFile, err := os.Open(filepath)

	if err != nil {
		log.Fatalf("failed to open file: %s", err)
	}

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	for i := 0; fileScanner.Scan() ; i++ {
		nodes = append(nodes, node{})
		err := json.Unmarshal([]byte(fileScanner.Text()), &nodes[i])
		if err != nil { panic(err) }
		if nodes[i].Id == 0 { panic("Error reading nodes - empty ID, line: " + fileScanner.Text()) }
	}

	return nodes
}

func readEdgesRaw(filepath string) []edge {
	var edges = make([]edge, 0)

	readFile, err := os.Open(filepath)

	if err != nil {
		log.Fatalf("failed to open file: %s", err)
	}

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	for i := 0; fileScanner.Scan() ; i++ {
		edges = append(edges, edge{})
		err := json.Unmarshal([]byte(fileScanner.Text()), &edges[i])
		if err != nil { panic(err) }
		if edges[i].From == 0 { panic("Error reading edges - empty From, line: " + fileScanner.Text()) }
	}

	return edges
}

func readGraphBBoxRaw(filepath string) types.MapBBox {
	result := types.MapBBox{}

	readFile, err := os.Open(filepath)

	if err != nil {
		log.Fatalf("failed to open file: %s", err)
	}

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	// tu i tak spodziewamy sie jednej linijki
	for i := 0; fileScanner.Scan() ; i++ { // FIXME - warning tutaj
		err := json.Unmarshal([]byte(fileScanner.Text()), &result)
		if err != nil { panic(err) }
		if result.Width == 0 { panic("Error reading bbox - empty Width, line: " + fileScanner.Text()) }
		return result
	}

	panic("Could not read Graph Bbox")
}