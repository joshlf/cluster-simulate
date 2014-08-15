package graph

import (
	"fmt"
	"testing"
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

func TestNumEdges(t *testing.T) {
	c := Cluster{}
	if c.cachedNumEdges != nil {
		t.Errorf("Expected nil cachedNumEdges; got %v", c.cachedNumEdges)
	}
	c.NumEdges()
	if c.cachedNumEdges == nil {
		t.Errorf("Expected non-nil cachedNumEdges; got nil")
	}
}

func TestMakeGraph(t *testing.T) {
	g := makeTestGraph()
	fmt.Println(g.stringPrefix("  ", ""))
	fmt.Println()
}

func TestGraphEqual(t *testing.T) {
	g, h := makeTestGraph(), makeTestGraph()
	if !g.Equal(h) {
		t.Errorf("Expected graphs to be equal; were not: \ng:\n%vh:\n%v", g, h)
	}

	// Test to make sure empty clusters
	// don't affect equality semantics.
	h.clusters.Add("foo", newCluster("foo"))
	if !g.Equal(h) {
		t.Errorf("Expected graphs to be equal; were not: \ng:\n%vh:\n%v", g, h)
	}

	if !h.Equal(g) {
		t.Errorf("Expected graphs to be equal; were not: \ng:\n%vh:\n%v", h, g)
	}
}

func TestNeighborClusters(t *testing.T) {
	g := makeTestGraph()
	neighbors := g.clusters[ClusterID("C2")].NeighborClusters()
	if len(neighbors) != 2 || (neighbors[0] != "C1" && neighbors[1] != "C1") || (neighbors[0] != "C3" && neighbors[1] != "C3") {
		t.Errorf("Expected neighbors C1 and C3; got %v", neighbors)
	}
}

func TestMerge(t *testing.T) {
	g, h := makeTestGraph(), makeTestGraph()
	c2 := g.clusters["C2"]

	// Cache values so we can check
	// if they've been flushed after merging.
	c2.NumBorderEdges()
	c2.NumBorderNodes()
	c2.NumEdges()
	c2.NeighborClusters()

	g.mergeClusters("C2", "C3")
	h.clusters["C2"].members.Add("F", h.nodes["F"])
	h.nodes["F"].cluster = h.clusters["C2"]
	h.clusters["C3"].members.Delete("F")
	if !g.Equal(h) {
		t.Errorf("Expected graphs to be equal; were not: \ng:\n%vh:\n%v", g, h)
	}

	// Check to make sure caches
	// were flushed
	c2 = g.clusters["C2"]
	if c2.NumBorderEdges() != 1 {
		t.Errorf("Expected 1 border edge; got %v", c2.NumBorderEdges())
	}
	if c2.NumBorderNodes() != 1 {
		t.Errorf("Expected 1 border node; got %v", c2.NumBorderNodes())
	}
	if c2.NumEdges() != 3 {
		t.Errorf("Expected 3 edges; got %v", c2.NumEdges())
	}
	neighbors := c2.NeighborClusters()
	if len(neighbors) != 1 || neighbors[0] != "C1" {
		t.Errorf("Expected neighbor C1; got %v", neighbors)
	}
}

func TestCluster(t *testing.T) {
	g := makeTestGraphNoClusters()
	fmt.Println(g)
	g.Merge()
	fmt.Println(g)
}
