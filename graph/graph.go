package graph

import (
	"fmt"
)

/*
	###########################
	########## GRAPH ##########
	###########################
*/

type Graph struct {
	nodes    graphNodeMap
	clusters graphClusterMap
	costFn   Cost
}

// Cluster returns the cluster with the given cluster ID,
// or nil if no such cluster exists.
func (g *Graph) Cluster(c ClusterID) *Cluster {
	clst, _ := g.clusters.Get(c)
	return clst
}

// Node returns the node with the given node ID,
// or nil of no such node exists.
func (g *Graph) Node(n NodeID) *Node {
	nd, _ := g.nodes.Get(n)
	return nd
}

// Clusters returns g's clusters. The returned map
// is not used internally, so the caller may feel
// free to mutate it.
func (g *Graph) Clusters() map[ClusterID]*Cluster {
	cs := g.clusters.Copy()
	for cid, c := range g.clusters {
		// Empty clusters don't exist semantically
		if c.members.Len() == 0 {
			delete(cs, cid)
		}
	}
	return cs
}

// NumClusters returns the number of clusters in g.
func (g *Graph) NumClusters() int {
	n := g.clusters.Len()
	for _, c := range g.clusters {
		// Empty clusters don't exist semantically
		if c.members.Len() == 0 {
			n--
		}
	}
	return n
}

// NumNodes returns the number of nodes in g.
func (g *Graph) NumNodes() int {
	return g.nodes.Len()
}

// Nodes returns g's nodes. The returned map
// is not used internally, so the caller may feel
// free to mutate it.
func (g *Graph) Nodes() map[NodeID]*Node {
	return g.nodes.Copy()
}

// OverlayLSDBSize returns the number of links
// in g's overlay LSDB.
func (g *Graph) OverlayLSDBSize() int {
	border := 0
	for _, c := range g.clusters {
		border += c.NumBorderEdges()
	}
	border /= 2 // Don't double-count

	virtual := 0
	for _, c := range g.clusters {
		virtual += c.NumVirtEdges()
	}
	return border + virtual
}

// RemoteLSDBSize returns the number of
// links in c's remote LSDB
func (g *Graph) RemoteLSDBSize(c ClusterID) int {
	return g.OverlayLSDBSize() - g.clusters[c].NumVirtEdges()
}

// Equal tests whether g and h are equal as defined
// by having the same topology with the same cluster
// IDs, node IDs, and edge weights. Empty clusters
// are ignored, as, semantically, they don't exist.
func (g *Graph) Equal(h *Graph) bool {
	return g.clusters.Equal(h.clusters)
}

func (g *Graph) stringPrefix(base, prefix string) string {
	newPrefix := prefix + base
	cStrings := make([]interface{}, 0)
	fmtString := prefix + "{\n"
	for _, c := range g.clusters {
		cStrings = append(cStrings, c.id.String(), fmt.Sprint(g.cost(c.id)), c.stringPrefix(base, newPrefix))
		fmtString += newPrefix + "%v (%v):\n%v"
	}
	fmtString += prefix + "}\n"
	return fmt.Sprintf(fmtString, cStrings...)
}

func (g *Graph) String() string {
	return g.stringPrefix("  ", "")
}

// Assumes that clusters are disjoint
func (g *Graph) mergeComputeUnmerge(c1, c2 ClusterID, f func() interface{}) interface{} {
	oldC1, _ := g.clusters.Get(c1)
	oldC2, _ := g.clusters.Get(c2)

	g.mergeClusters(c1, c2)

	res := f()

	newC := c1
	if c1 > c2 {
		newC = c2
	}

	// NOTE: Must happen in this order!
	g.clusters.Delete(newC)
	g.clusters.Add(c1, oldC1)
	g.clusters.Add(c2, oldC2)
	oldC1.resetMemberClusterPointers()
	oldC2.resetMemberClusterPointers()
	oldC1.flushCache()
	oldC2.flushCache()

	return res
}

// When merging two clusters, the lexically
// smaller cluster remains, adopting the nodes
// of the lexically larger cluster.
func (g *Graph) mergeClusters(c1, c2 ClusterID) {
	// Enforce the lexical ordering
	if c2 < c1 {
		c1, c2 = c2, c1
	}

	// Make an entirley new cluster
	// (instead of adding to c1)
	// to make sure that all caches
	// are invalidated (could do this
	// manually, but this code and
	// the caches which are kept
	// could easily get out of sync).
	newC := newCluster(c1)
	c, _ := g.clusters.Get(c1)
	d, _ := g.clusters.Get(c2)
	for _, n := range c.members {
		newC.add(n)
	}
	for _, n := range d.members {
		newC.add(n)
	}

	g.clusters.Delete(c1)
	g.clusters.Delete(c2)
	g.clusters.Add(c1, newC)
}

func (g *Graph) cost(c ClusterID) int {
	return g.costFn(g, c)
}

/*
	##########################
	########## NODE ##########
	##########################
*/

type Node struct {
	id      NodeID
	cluster *Cluster
	edges   edgeMap
}

// NodeID returns n's node ID.
func (n *Node) NodeID() NodeID {
	return n.id
}

// ClusterID returns the id of n's cluster.
func (n *Node) ClusterID() ClusterID {
	return n.cluster.id
}

// Equal returns whether n and m are equal as defined
// by having equal node IDs, and equal edges of equal
// lengths to nodes with the same node IDs.
func (n *Node) Equal(m *Node) bool {
	return n.id.Equal(m.id) && n.cluster.id.Equal(m.cluster.id) && n.edges.Equal(m.edges)
}

// NumEdges returns the number of edges from n
func (n *Node) NumEdges() int {
	return n.edges.Len()
}

// NumEdgesInCluster returns the number of edges
// from n whose destination node is in the same
// cluster as n.
func (n *Node) NumEdgesInCluster() int {
	num := len(n.edges)
	cid := n.cluster.id
	for _, e := range n.edges {
		if e.dst.cluster.id != cid {
			num--
		}
	}
	return num
}

// NumEdgesOutCluster returns the number of edges
// from n whose destination node is in a different
// cluster from n.
func (n *Node) NumEdgesOutCluster() int {
	num := 0
	cid := n.cluster.id
	for _, e := range n.edges {
		if e.dst.cluster.id != cid {
			num++
		}
	}
	return num
}

// IsBorderNode returns whether any of n's neighbors
// are members of a different cluster from n.
func (n *Node) IsBorderNode() bool {
	cid := n.cluster.id
	for _, e := range n.edges {
		if e.dst.cluster.id != cid {
			return true
		}
	}
	return false
}

func (n *Node) String() string {
	cStrings := make([]interface{}, 0)
	fmtString := "{ "
	for _, e := range n.edges {
		cStrings = append(cStrings, e.dst.stringNeighbor(), e.cost)
		fmtString += "%v(%v), "
	}
	fmtString += "}"
	return fmt.Sprintf(fmtString, cStrings...)
}

func (n *Node) stringNeighbor() string {
	return "{" + n.id.String() + "/" + n.cluster.id.String() + "}"
}

/*
	#############################
	########## CLUSTER ##########
	#############################
*/

type Cluster struct {
	id      ClusterID
	members memberNodeMap

	// Cached methods
	cachedNumEdges         func() int
	cachedNumBorderNodes   func() int
	cachedNumBorderEdges   func() int
	cachedNeighborClusters func() []ClusterID
	cachedLocalLSDBSize    func() int
}

func newCluster(id ClusterID) *Cluster {
	return &Cluster{
		id:      id,
		members: newMemberNodeMap(),
	}
}

// ClusterID returns c's cluster ID.
func (c *Cluster) ClusterID() ClusterID {
	return c.id
}

// Node returns the node with the given node ID,
// or nil of no such node exists.
func (c *Cluster) Node(n NodeID) *Node {
	nd, _ := c.members.Get(n)
	return nd
}

// NumNodes returns the number of nodes in c.
func (c *Cluster) NumNodes() int {
	return c.members.Len()
}

// Nodes returns c's nodes. The returned map
// is not used internally, so the caller may feel
// free to mutate it.
func (c *Cluster) Nodes() map[NodeID]*Node {
	return c.members.Copy()
}

// Equal returns whether c and d are equal as defined
// by having equal cluster IDs and equal sets of nodes
// (where node equality is defined as (*Node).Equal()).
func (c *Cluster) Equal(d *Cluster) bool {
	return c.id.Equal(d.id) && c.members.Equal(d.members)
}

// Set the cluster pointer in each of c's
// members to point to c. This is useful
// if another temporary cluster has been
// created, and we wish to revert the change.
func (c *Cluster) resetMemberClusterPointers() {
	for _, n := range c.members {
		n.cluster = c
	}
}

// Remove all cached pre-computed values
func (c *Cluster) flushCache() {
	c.cachedNumEdges = nil
	c.cachedNumBorderNodes = nil
	c.cachedNumBorderEdges = nil
	c.cachedNeighborClusters = nil
}

// Add n and set n's cluster pointer to c
func (c *Cluster) add(n *Node) {
	c.members.Add(n.id, n)
	n.cluster = c
}

/*
	CLUSTER NUM EDGES
*/

// NumEdges returns the number of edges between
// members of c.
func (c *Cluster) NumEdges() int {
	if c.cachedNumEdges == nil {
		n := c.computeNumEdges()
		c.cachedNumEdges = makeIntFunc(n)
		return n
	}
	return c.cachedNumEdges()
}

func (c *Cluster) computeNumEdges() int {
	n := 0
	for _, node := range c.members {
		n += node.NumEdgesInCluster()
	}
	return n / 2
}

/*
	CLUSTER NUM VIRTUAL EDGES
*/

// NumVirtEdges returns the number of virtual
// edges that c contributes to the overlay
// LSDB.
func (c *Cluster) NumVirtEdges() int {
	n := c.NumBorderNodes()
	return (n * (n - 1)) / 2
}

/*
	CLUSTER BORDER NODES
*/

// NumBorderNodes returns the number of
// border nodes in c.
func (c *Cluster) NumBorderNodes() int {
	if c.cachedNumBorderNodes == nil {
		n := c.computeNumBorderNodes()
		c.cachedNumBorderNodes = makeIntFunc(n)
		return n
	}
	return c.cachedNumBorderNodes()
}

func (c *Cluster) computeNumBorderNodes() int {
	n := 0
	for _, node := range c.members {
		if node.IsBorderNode() {
			n++
		}
	}
	return n
}

/*
	CLUSTER BORDER EDGES
*/

// NumBorderEdges returns the number of edges
// connecting members of c to nodes which are not
// members of c
func (c *Cluster) NumBorderEdges() int {
	if c.cachedNumBorderEdges == nil {
		n := c.computeNumBorderEdges()
		c.cachedNumBorderEdges = makeIntFunc(n)
		return n
	}
	return c.cachedNumBorderEdges()
}

func (c *Cluster) computeNumBorderEdges() int {
	n := 0
	for _, node := range c.members {
		n += node.NumEdgesOutCluster()
	}
	return n
}

/*
	CLUSTER NEIGHBOR CLUSTERS
*/

// NeighborClusters returns the cluster IDs
// of c's neighboring clusters.
func (c *Cluster) NeighborClusters() []ClusterID {
	if c.cachedNeighborClusters == nil {
		ids := c.computeNeighborClusters()
		c.cachedNeighborClusters = func() []ClusterID { return ids }
		return ids
	}
	return c.cachedNeighborClusters()
}

func (c *Cluster) computeNeighborClusters() []ClusterID {
	m := make(map[ClusterID]struct{})

	for _, n := range c.members {
		for _, o := range n.edges {
			m[o.dst.cluster.id] = struct{}{}
		}
	}
	// It's likely that there's more than
	// one node in the cluster, and thus
	// all border nodes are neighbors with
	// a node in the cluster
	delete(m, c.id)

	// NOTE: length MUST be 0 (or else
	// we'll be appending to the end of
	// a zeroed slice, and get a slice
	// of length len(m) * 2).
	ids := make([]ClusterID, 0, len(m))
	for id := range m {
		ids = append(ids, id)
	}
	return ids
}

/*
	CLUSTER LOCAL LSDB SIZE
*/

// LocalLSDBSize returns the number of links
// in c's local LSDB.
func (c *Cluster) LocalLSDBSize() int {
	if c.cachedLocalLSDBSize == nil {
		size := c.computeLocalLSDBSize()
		c.cachedLocalLSDBSize = makeIntFunc(size)
		return size
	}
	return c.cachedLocalLSDBSize()
}

func (c *Cluster) computeLocalLSDBSize() int {
	return c.NumEdges()
}

/*
	CLUSTER STRING FORMATTING
*/

func (c *Cluster) stringPrefix(base, prefix string) string {
	newPrefix := prefix + base
	cStrings := make([]interface{}, 0)
	fmtString := ""
	// fmtString := prefix + "{\n"
	for _, n := range c.members {
		cStrings = append(cStrings, n.id.String(), n.String())
		fmtString += newPrefix + "%v: %v\n"
	}
	// fmtString += prefix + "}\n"
	return fmt.Sprintf(fmtString, cStrings...)
}

func (c *Cluster) String() string {
	return c.stringPrefix("  ", "")
}
