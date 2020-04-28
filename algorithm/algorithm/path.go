package algorithm

import (
	"math"
	"mzakrze/smart_city/types"
	"sort"
)

type node struct {
	id types.NodeId
	visited bool
	totalDist float64
	spTparent *node
	neighbours []*neighbour
}

type neighbour struct {
	n *node
	dist float64
}

func findShortestPathFromTo(source, dest *types.Node, graph *types.Graph) []*types.Node {

	_, allNodes := createShortestPathTree(source, dest, graph)
	nodesMap := make(map[types.NodeId]*types.Node)

	var s *node
	for i, n := range allNodes {
		if n.id == dest.Id {
			s = allNodes[i]
		}
		if n.id == source.Id {
			if allNodes[i].spTparent != nil {
				panic("source has a parent")
			}
		}
		nodesMap[n.id] = &graph.AllNodes[i]
	}
	if s == nil { panic("err") }

	result := make([]*types.Node, 0)
	for s != nil {
		result = append([]*types.Node{nodesMap[s.id]}, result...)
		s = s.spTparent
	}

	return result
}

func createShortestPathTree(source, dest *types.Node, graph *types.Graph) ( *node, []*node) {

	allNodes := make([]*node, len(graph.AllNodes))
	distances := make(map[types.NodeId]float64)
	nodesMap := make(map[types.NodeId]*node)
	var currentNode *node
	for i, n := range graph.AllNodes {
		distances[n.Id] = math.Inf(1)
		allNodes[i] = &node{
			id: n.Id,
			totalDist: math.Inf(1),
			visited: false,
			spTparent: nil,
			neighbours: []*neighbour{},
		}
		nodesMap[n.Id] = allNodes[i]
	}

	for i, n := range graph.AllNodes {
		if n.Id == source.Id {
			currentNode = allNodes[i]
		}
		for _, e := range n.Edges {
			// tutaj polegamy na tym, ze indeksy sa takie same
			allNodes[i].neighbours = append(allNodes[i].neighbours, &neighbour{
				n: nodesMap[e.To.Id],
				dist: e.Distance,
			})
		}
	}

	getNextFromPriorityQueue := func() *node {
		var res *node
		minDist := math.Inf(1)
		for i, n := range allNodes {
			if n.visited == false {
				if n.totalDist < minDist {
					res = allNodes[i]
					minDist = res.totalDist
				}
			}
		}
		return res
	}


	filterNotVisited := func(ns []*neighbour) []*neighbour {
		res := make([]*neighbour, 0)
		for i := range ns {
			res = append(res, ns[i])
		}
		return res
	}

	sortAscByDistance := func(ns []*neighbour) []*neighbour {
		sort.Slice(ns, func(i, j int) bool {
			return ns[i].dist < ns[j].dist
		})
		return ns
	}

	currentNode.totalDist = 0
	distances[currentNode.id] = 0
	for currentNode != nil {
		neighbours := sortAscByDistance(filterNotVisited(currentNode.neighbours))

		for i, n := range neighbours {
			if currentNode.totalDist + n.dist < distances[n.n.id] {
				distances[n.n.id] = currentNode.totalDist + n.dist
				n.n.totalDist = distances[n.n.id]
				neighbours[i].n.spTparent= currentNode
			}
		}

		currentNode.visited = true
		currentNode = getNextFromPriorityQueue()
	}

	return currentNode, allNodes
}

