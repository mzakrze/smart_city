package algorithm

import (
	"mzakrze/smart_city/types"
	"testing"
)

func TestFindPath1(t *testing.T) {

	// given
	a := types.Node{ Id: 1, Edges: []types.Edge{} }
	b := types.Node{ Id: 2, Edges: []types.Edge{}  }
	c := types.Node{ Id: 3, Edges: []types.Edge{}  }
	d := types.Node{ Id: 4, Edges: []types.Edge{}  }
	e := types.Node{ Id: 5, Edges: []types.Edge{}  }
	f := types.Node{ Id: 6, Edges: []types.Edge{}  }

	a.Edges = append(a.Edges, types.Edge{To: &b, Distance: 1})
	b.Edges = append(b.Edges, types.Edge{To: &c, Distance: 1})
	c.Edges = append(b.Edges, types.Edge{To: &d, Distance: 1})
	d.Edges = append(b.Edges, types.Edge{To: &e, Distance: 1})
	e.Edges = append(b.Edges, types.Edge{To: &f, Distance: 1})

	allNodes := []types.Node{a,b,c,d,e,f}

	graph := types.Graph{
		AllNodes: allNodes,
	}

	// when
	path := findShortestPathFromTo(&a, &d, &graph)

	// then
	if path[0].Id != 1 {
		t.Fatal("Err")
	}
	if path[1].Id != 2 {
		t.Fatal("Err")
	}
	if path[2].Id != 3 {
		t.Fatal("Err")
	}
	if path[3].Id != 4 {
		t.Fatal("Err")
	}
}

func TestFindPath2(t *testing.T) {

	// given
	a := types.Node{ Id: 1, Edges: []types.Edge{} }
	b := types.Node{ Id: 2, Edges: []types.Edge{}  }
	c := types.Node{ Id: 3, Edges: []types.Edge{}  }
	d := types.Node{ Id: 4, Edges: []types.Edge{}  }
	e := types.Node{ Id: 5, Edges: []types.Edge{}  }
	f := types.Node{ Id: 6, Edges: []types.Edge{}  }

	a.Edges = append(a.Edges, types.Edge{To: &b, Distance: 1})
	a.Edges = append(a.Edges, types.Edge{To: &c, Distance: 1})

	b.Edges = append(b.Edges, types.Edge{To: &c, Distance: 1})
	b.Edges = append(b.Edges, types.Edge{To: &d, Distance: 999})

	c.Edges = append(b.Edges, types.Edge{To: &d, Distance: 1})
	d.Edges = append(b.Edges, types.Edge{To: &e, Distance: 1})
	e.Edges = append(b.Edges, types.Edge{To: &f, Distance: 1})

	allNodes := []types.Node{a,b,c,d,e,f}

	graph := types.Graph{
		AllNodes: allNodes,
	}

	// when
	path := findShortestPathFromTo(&a, &d, &graph)

	// then
	if path[0].Id != 1 {
		t.Fatal("Err")
	}
	if path[1].Id != 3 {
		t.Fatal("Err")
	}
	if path[2].Id != 4 {
		t.Fatal("Err")
	}
}