package graph

type NodeDef struct {
	ID      NodeID
	Cluster ClusterID
}

type LinkDef struct {
	A, B NodeID
	Cost uint64
}

// The GraphDef type provides a simple data structure to
// hold definitions of graphs. Canonically, links should
// not appear in both directions (that is, a->b and b->a),
// though NewGraph will still behave properly if such
// duplication exists.
type GraphDef struct {
	Nodes []NodeDef
	Links []LinkDef
}

// NewGraph constructs a native Graph data structure from
// the definition of a graph. Canonically, def should not
// include both directions of any given link (that is,
// a->b and b->a), though it will still behave properly
// if such duplication exists.
func NewGraph(def GraphDef, c Cost) *Graph {
	g := &Graph{
		nodes:    newGraphNodeMap(),
		clusters: newGraphClusterMap(),
		costFn:   c,
	}

	for _, n := range def.Nodes {
		c, ok := g.clusters.Get(n.Cluster)
		if !ok {
			c = &Cluster{
				id:      n.Cluster,
				members: newMemberNodeMap(),
			}
			g.clusters.Add(n.Cluster, c)
		}
		g.nodes.Add(n.ID, &Node{
			id:      n.ID,
			cluster: c,
			edges:   newEdgeMap(),
		})
		node, _ := g.nodes.Get(n.ID)
		c.members.Add(n.ID, node)
	}

	for _, l := range def.Links {
		a, _ := g.nodes.Get(l.A)
		b, _ := g.nodes.Get(l.B)
		a.edges.Add(l.B, edge{
			cost: l.Cost,
			dst:  b,
		})
		b.edges.Add(l.A, edge{
			cost: l.Cost,
			dst:  a,
		})
	}
	return g
}

// GraphDef creates a canonincal GraphDef describing g.
func (g *Graph) GraphDef() GraphDef {
	links := make(map[LinkDef]struct{})
	for _, n := range g.nodes {
		for _, e := range n.edges {
			if n.id < e.dst.id {
				links[LinkDef{
					A:    n.id,
					B:    e.dst.id,
					Cost: e.cost,
				}] = struct{}{}
			}
		}
	}

	gd := GraphDef{
		Nodes: make([]NodeDef, 0),
		Links: make([]LinkDef, 0),
	}

	for _, n := range g.nodes {
		gd.Nodes = append(gd.Nodes, NodeDef{
			ID:      n.id,
			Cluster: n.cluster.id,
		})
	}
	for l := range links {
		gd.Links = append(gd.Links, l)
	}
	return gd
}
