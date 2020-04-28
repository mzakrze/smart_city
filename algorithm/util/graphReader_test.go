package util

import (
	"fmt"
	"math"
	"mzakrze/smart_city/types"
	"testing"
)


func TestAssembleGraph(t *testing.T) {

	nodesRaw := readNodesRaw("../../map_visualization/public/nodes.ndjson")
	edgesRaw := readEdgesRaw("../../map_visualization/public/edges.ndjson")
	mapBox := types.MapBBox{}

	graph := assembleGraph(nodesRaw, edgesRaw, mapBox)

	degIn := make(map[types.NodeId]int)
	degOut := make(map[types.NodeId]int)

	for _, n := range graph.AllNodes {
		degIn[n.Id] = 0
		degOut[n.Id] = 0
	}

	for _, n := range graph.AllNodes {
		for _, e := range n.Edges {
			if e.To == nil {
				t.Fatal("Found empty edge")
			}
			if math.IsNaN(e.Distance) {
				t.Fatal("Distance between nodes is NaN")
			}
			if e.To.Id == n.Id {
				t.Fatal(fmt.Sprintf("Self loop, nodeId: %d", n.Id))
			}

			degOut[n.Id] = degOut[n.Id] + 1
			degIn[e.To.Id] = degIn[e.To.Id] + 1
		}
	}

	for _, n := range graph.AllNodes {
		if degOut[n.Id] + degIn[n.Id] == 0 {
			t.Fatal(fmt.Sprintf("Found isolated node, id: %d, degOut: %d, degIn: %d", n.Id, degOut[n.Id], degIn[n.Id]))
		}
	}



}
