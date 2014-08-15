package graph

type Cost func(g *Graph, c ClusterID) int

var (
	// MaxCost defines cluster cost as the maximum of the
	// remote LSDB size and the local LSDB size.
	MaxCost Cost = maxCost
)

// The cost of merging c and d
func (g *Graph) mergeCost(c, d ClusterID) int {
	if c == d {
		return g.cost(c)
	}
	f := func() interface{} {
		// Make sure to use the right
		// id (the other cluster will
		// not exist during this function)
		newC := c
		if c > d {
			newC = d
		}
		return g.cost(newC)
	}
	return g.mergeComputeUnmerge(c, d, f).(int)
}

func maxCost(g *Graph, c ClusterID) int {
	rem := g.RemoteLSDBSize(c)
	loc := g.Cluster(c).LocalLSDBSize()
	if rem > loc {
		return rem
	}
	return loc
}
