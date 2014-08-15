package graph

import "fmt"

type NodeID string

// Equal returns whether n and m are equal.
func (n NodeID) Equal(m NodeID) bool {
	return n == m
}

func (n NodeID) String() string { return string(n) }

type ClusterID string

// Equal returns whether c and d are equal.
func (c ClusterID) Equal(d ClusterID) bool {
	return c == d
}

// Comparison function for ClusterID
// so that ClusterIDs may be ordered
// (defines an ordering on (ClusterID, ClusterID);
// returns (a, b) < (b, c))
func clusterIDLess(a, b, c, d ClusterID) bool {
	var ab, cd ClusterID
	if a < b {
		ab = a + b
	} else {
		ab = b + a
	}
	if c < d {
		cd = c + d
	} else {
		cd = d + c
	}
	return ab < cd
}

func (c ClusterID) String() string { return string(c) }

/*
	GRAPH NODE MAP
*/

type graphNodeMap map[NodeID]*Node

func newGraphNodeMap() graphNodeMap { return make(graphNodeMap) }

func (g graphNodeMap) Add(nid NodeID, n *Node) { g[nid] = n }

func (g graphNodeMap) Get(nid NodeID) (*Node, bool) { n, ok := g[nid]; return n, ok }

func (g graphNodeMap) Len() int { return len(g) }

func (g graphNodeMap) Copy() graphNodeMap {
	h := newGraphNodeMap()
	for nid, n := range g {
		h.Add(nid, n)
	}
	return h
}

/*
	MEMBER NODE MAP
*/

type memberNodeMap map[NodeID]*Node

func newMemberNodeMap() memberNodeMap { return make(memberNodeMap) }

func (m memberNodeMap) Add(nid NodeID, n *Node) { m[nid] = n }

func (m memberNodeMap) Get(nid NodeID) (*Node, bool) { n, ok := m[nid]; return n, ok }

func (m memberNodeMap) Delete(nid NodeID) { delete(m, nid) }

func (m memberNodeMap) Len() int { return len(m) }

func (m memberNodeMap) Copy() memberNodeMap {
	n := newMemberNodeMap()
	for nid, nd := range m {
		n.Add(nid, nd)
	}
	return n
}

func (m memberNodeMap) Equal(n memberNodeMap) bool {
	for k, v1 := range m {
		v2, ok := n[k]
		if !ok {
			return false
		}
		if !v1.Equal(v2) {
			return false
		}
	}
	return true
}

/*
	GRAPH CLUSTER MAP
*/

type graphClusterMap map[ClusterID]*Cluster

func newGraphClusterMap() graphClusterMap { return make(graphClusterMap) }

func (g graphClusterMap) Add(cid ClusterID, c *Cluster) { g[cid] = c }

func (g graphClusterMap) Get(cid ClusterID) (*Cluster, bool) { c, ok := g[cid]; return c, ok }

func (g graphClusterMap) Delete(cid ClusterID) { delete(g, cid) }

func (g graphClusterMap) Len() int { return len(g) }

func (g graphClusterMap) Copy() graphClusterMap {
	h := newGraphClusterMap()
	for cid, c := range g {
		h.Add(cid, c)
	}
	return h
}

func (g graphClusterMap) Equal(h graphClusterMap) bool {
	htmp := make(graphClusterMap)
	for cid, c := range h {
		htmp.Add(cid, c)
	}
	for k, v1 := range g {
		v2, ok := h[k]
		if !ok {
			// If there are clusters in g that are not
			// in h, that's ok so long as they're all
			// empty (in which case, semantically,
			// they don't exist).
			if v1.members.Len() != 0 {
				return false
			}
			continue
		}
		htmp.Delete(k)
		if !v1.Equal(v2) {
			return false
		}
	}
	// If there are clusters in h that are not
	// in g, that's ok so long as they're all
	// empty (in which case, semantically,
	// they don't exist).
	for _, c := range htmp {
		if c.members.Len() != 0 {
			fmt.Println(c.id, c.members)
			return false
		}
	}
	return true
}

/*
	EDGE MAP
*/

type edgeMap map[NodeID]edge

func newEdgeMap() edgeMap { return make(edgeMap) }

func (e edgeMap) Add(nid NodeID, edge edge) { e[nid] = edge }

func (e edgeMap) Len() int { return len(e) }

// Only check for id equality. If we checked
// for deep node equality, we'd get infinite
// recursion since deep node equality calls
// this method.
func (e edgeMap) Equal(f edgeMap) bool {
	for k, v1 := range e {
		v2, ok := f[k]
		if !ok {
			return false
		}
		if !v1.dst.id.Equal(v2.dst.id) {
			return false
		}
		if v1.cost != v2.cost {
			return false
		}
	}
	return true
}

type edge struct {
	cost uint64
	dst  *Node
}
