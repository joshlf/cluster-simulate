package encoding

import (
	"fmt"
	"testing"

	. "github.com/synful/cluster-simulate/graph"
)

func makeTestGraph() *Graph {
	/*
		  B
		 / \
		A---C
		    |
		E---D
		 \ /
		  F
	*/
	def := GraphDef{
		Nodes: []NodeDef{
			NodeDef{"A", "C1"},
			NodeDef{"B", "C1"},
			NodeDef{"C", "C1"},
			NodeDef{"D", "C2"},
			NodeDef{"E", "C2"},
			NodeDef{"F", "C3"},
		},
		Links: []LinkDef{
			LinkDef{"A", "B", 1},
			LinkDef{"B", "C", 2},
			LinkDef{"C", "A", 3},
			LinkDef{"C", "D", 4},
			LinkDef{"D", "E", 5},
			LinkDef{"E", "F", 6},
			LinkDef{"F", "D", 7},
		},
	}

	return NewGraph(def, MaxCost)
}

func makeTestGraphNoClusters() *Graph {
	/*
		  B
		 / \
		A---C
		    |
		E---D
		 \ /
		  F
	*/
	def := GraphDef{
		Nodes: []NodeDef{
			NodeDef{"A", "C1"},
			NodeDef{"B", "C2"},
			NodeDef{"C", "C3"},
			NodeDef{"D", "C4"},
			NodeDef{"E", "C5"},
			NodeDef{"F", "C6"},
		},
		Links: []LinkDef{
			LinkDef{"A", "B", 1},
			LinkDef{"B", "C", 2},
			LinkDef{"C", "A", 3},
			LinkDef{"C", "D", 4},
			LinkDef{"D", "E", 5},
			LinkDef{"E", "F", 6},
			LinkDef{"F", "D", 7},
		},
	}

	return NewGraph(def, MaxCost)
}

func TestEncoding(t *testing.T) {
	testEncoding(makeTestGraph(), t)
	testEncoding(makeTestGraphNoClusters(), t)
}

func testEncoding(g *Graph, t *testing.T) {
	encoding, err := Marshal(g.GraphDef())
	if err != nil {
		t.Errorf("Error marshalling: %v", err)
	}
	def, err := Unmarshal(encoding)
	h := NewGraph(def, MaxCost)
	if err != nil {
		t.Errorf("Error unmarshalling: %v", err)
	}
	if !g.Equal(h) {
		fmt.Println(g)
		fmt.Println("\n")
		fmt.Println(h)
		t.Errorf("Expected graphs to be equal")
	}
}
