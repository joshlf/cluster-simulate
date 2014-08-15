package graph

import "sort"

type mergeCost func(c, d ClusterID) int

// MergeRound attempts to perform a round of merging.
// If no merging is required (that is, the graph
// has stabilized), it does not perform a round.
// It returns whether a round was performed.
//
// Additionally, if merge is not nil, it will be
// called every time a merge is performed. If c.Equal(d),
// c merged with itself (that is, decided that the best
// option was not to merge).
func (g *Graph) MergeRound(merge func(c, d ClusterID)) bool {
	// Indicates whether a change was made
	// as far as round stability is concerned
	// (that is, if changedOverall == false,
	// the graph is considered to have stabilized)
	changedOverall := false

	preferences := make(map[ClusterID][]ClusterID)

	// Make copy of cluster list since
	// g.clusters will be modified during
	// computation of proposals (during
	// mergeComputeUnmerge), and this
	// makes iteration over g.clusters
	// undefined.
	clusters := make(map[ClusterID]struct{})
	for c := range g.clusters {
		clusters[c] = struct{}{}
	}

	for c := range clusters {
		preferences[c] = g.proposeMerge(c)
	}

	merged := make(map[ClusterID]struct{})
	for {
		changed := false
		for c, p := range preferences {
			if _, ok := merged[c]; ok {
				// c has already merged
				continue
			}
			switch {
			case c == p[0]:
				// c proposed to merge with itself
				merged[c] = struct{}{}
				if merge != nil {
					merge(c, c)
				}

				// This isn't considered an overall change
				// (a stable round of clustering is one in
				// which everyone merges with themselves)
				changed = true
			case c == preferences[p[0]][0]:
				// It's a match!

				// We only want to merge once,
				// so impose this arbitrary
				// ordering. Note that this
				// condition will be true exactly
				// once per pair.
				if c < p[0] {

					merged[c] = struct{}{}
					merged[p[0]] = struct{}{}
					changed, changedOverall = true, true
					if merge != nil {
						merge(c, p[0])
					}

					g.mergeClusters(c, p[0])
				}
			}
		}

		for c := range clusters {
			for {
				_, firstChoiceMerged := merged[preferences[c][0]]
				_, cMerged := merged[c]
				if firstChoiceMerged && !cMerged {
					preferences[c] = preferences[c][1:]
				} else {
					break
				}
			}
		}

		if !changed {
			break
		}
	}

	return changedOverall
}

// MergeCallback performs rounds of merging until the graph stabilizes,
// and returns the number of rounds performed. Before each round,
// round is called unless it is nil.
//
// Additionally, if merge is not nil, it will be
// called every time a merge is performed. If c.Equal(d),
// c merged with itself (that is, decided that the best
// option was not to merge).
func (g *Graph) MergeCallback(round func(), merge func(c, d ClusterID)) int {
	n := 0
	if round != nil {
		round()
	}
	for g.MergeRound(merge) {
		n++
		if round != nil {
			round()
		}
	}
	return n
}

// Merge performs rounds of merging until the graph stabilizes,
// and returns the number of rounds performed.
func (g *Graph) Merge() int {
	return g.MergeCallback(nil, nil)
	// n := 0
	// t := time.Now()
	// // fmt.Printf("%v (%v)\n", t, t.UnixNano())
	// // fmt.Printf("Round %v\n", n)
	// // fmt.Println()
	// // writeState(g, n)
	// for g.MergeRound() {
	// 	n++
	// 	// t := time.Now()
	// 	// fmt.Printf("%v (%v)\n", t, t.UnixNano())
	// 	// fmt.Printf("Round %v\n", n)
	// 	// fmt.Println()
	// 	// writeState(g, n)
	// }
	// return n
}

type preferenceList struct {
	list []ClusterID
	g    *Graph
	c    ClusterID
	f    mergeCost
}

func (p preferenceList) Len() int      { return len(p.list) }
func (p preferenceList) Swap(i, j int) { p.list[i], p.list[j] = p.list[j], p.list[i] }
func (p preferenceList) Less(i, j int) bool {
	icost := p.f(p.c, p.list[i])
	jcost := p.f(p.c, p.list[j])
	switch {
	case icost < jcost:
		return true
	case jcost < icost:
		return false
	default:
		return clusterIDLess(p.c, p.list[i], p.c, p.list[j])
	}
	panic("Reached unreachable code")
}

// Return order in which c would prefer
// to merge with other clusters, ending
// with c itself.
func (g *Graph) proposeMerge(c ClusterID) []ClusterID {
	p := preferenceList{
		list: append(g.clusters[c].NeighborClusters(), c),
		g:    g,
		c:    c,
		f:    g.mergeCost,
	}
	sort.Sort(p)
	for i, id := range p.list {
		if id.Equal(c) {
			// fmt.Println(c, p.list[:i+1])
			return p.list[:i+1]
		}
	}
	panic("Reached unreachable code")
}
