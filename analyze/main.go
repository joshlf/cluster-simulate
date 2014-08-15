package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/synful/cluster-simulate/graph"
	"github.com/synful/cluster-simulate/graph/encoding"
)

const (
	ERR_IO = 3 + iota
	ERR_PARSE
)

var (
	graphFilename = flag.String("graph", "", "a file containing the graph to cluster")
	basic         = flag.Bool("basic", false, "show basic whole-graph statistics such as number of clusters and nodes")
	advanced      = flag.Bool("advanced", false, "show advanced whole-graph statistics such as average cost metrics")
	clusters      = flag.String("clusters", "", "comma-separated list of clusters to perform cluster-specific analysis on, or \"all\"")
	basicCluster  = flag.Bool("basicCluster", false, "show basic statistics for each cluster such as number of nodes and border nodes")
	maxCost       = flag.Bool("maxCost", false, "compute the MaxCost metric for each cluster and the advanced statistics")
)

var (
	clusterAnalyzers = []func(*graph.Graph, *graph.Cluster){runBasicCluster, runMaxCost}
	overallAnalyzers = []func(*graph.Graph){runOverallMaxCost}
)

var maxInt int

func init() {
	u := uint(0)
	u--
	u >>= 1
	maxInt = int(u)
}

func main() {
	flag.Parse()

	f, err := os.Open(*graphFilename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening graph file: %v\n", err)
		os.Exit(ERR_IO)
	}

	gd, err := ioutil.ReadAll(f)
	def, err := encoding.Unmarshal(gd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing graph file: %v\n", err)
		os.Exit(ERR_PARSE)
	}

	g := graph.NewGraph(def, graph.MaxCost)

	var clusterList map[graph.ClusterID]*graph.Cluster
	if *clusters == "all" {
		clusterList = g.Clusters()
	} else {
		clusterList = make(map[graph.ClusterID]*graph.Cluster)
		arr := strings.Split(*clusters, ",")
		for _, str := range arr {
			// Allow for "--clusters=all,"
			// (in case there's a cluster
			// called "all")
			if str == "" {
				continue
			}
			cid := graph.ClusterID(str)
			c := g.Cluster(cid)
			if c == nil {
				fmt.Fprintf(os.Stderr, "Nonexistant cluster ID: %v\n", str)
				continue
			}
			clusterList[cid] = c
		}
	}

	printedYet := false
	if *basic {
		fmt.Println("BASIC STATISTICS")
		runBasic(g)
		printedYet = true
	}

	if *advanced {
		if printedYet {
			fmt.Println()
		}
		printedYet = true
		fmt.Println("ADVANCED STATISTICS")
		for _, f := range overallAnalyzers {
			f(g)
		}
	}

	if len(clusterList) > 0 {
		if printedYet {
			fmt.Println()
		}

		fmt.Println("PER-CLUSTER ANALYSIS")
		for cid, c := range clusterList {
			fmt.Printf("  %v:\n", cid)
			for _, f := range clusterAnalyzers {
				f(g, c)
			}
		}
	}
}

func runBasicCluster(g *graph.Graph, c *graph.Cluster) {
	if !*basicCluster {
		return
	}
}

func runMaxCost(g *graph.Graph, c *graph.Cluster) {
	if !*maxCost {
		return
	}
	fmt.Printf("    MaxCost: %v\n", graph.MaxCost(g, c.ClusterID()))
}

func runOverallMaxCost(g *graph.Graph) {
	if !*maxCost {
		return
	}
	cost := 0
	highest, hcost := graph.ClusterID(""), 0
	lowest, lcost := graph.ClusterID(""), maxInt
	for _, c := range g.Clusters() {
		tmpcost := graph.MaxCost(g, c.ClusterID())
		cost += tmpcost
		if tmpcost > hcost {
			highest, hcost = c.ClusterID(), tmpcost
		}
		if tmpcost < lcost {
			lowest, lcost = c.ClusterID(), tmpcost
		}
	}

	fmt.Printf("    Average MaxCost: %v\n", float64(cost)/float64(g.NumClusters()))
	fmt.Printf("    Total MaxCost: %v\n", cost)
	fmt.Printf("    Cluster with highest MaxCost: %v (%v)\n", highest, hcost)
	fmt.Printf("    Cluster with lowest MaxCost: %v (%v)\n", lowest, lcost)
}

func runBasic(g *graph.Graph) {
	fmt.Printf("  Number of nodes: %v\n", g.NumNodes())
	fmt.Printf("  Number of clusters: %v\n", g.NumClusters())
	fmt.Printf("  Average nodes per cluster: %.2f\n", float64(g.NumNodes())/float64(g.NumClusters()))
	borderNodes := 0
	for _, n := range g.Nodes() {
		if n.IsBorderNode() {
			borderNodes++
		}
	}
	fmt.Printf("  Number of border nodes: %v\n", borderNodes)
	fmt.Printf("  Average border nodes per cluster: %.2f\n", float64(borderNodes)/float64(g.NumClusters()))
	edges := 0
	for _, n := range g.Nodes() {
		edges += n.NumEdges()
	}
	edges /= 2
	fmt.Printf("  Number of edges: %v\n", edges)
	fmt.Printf("  Average edges per node: %.2f\n", float64(2*edges)/float64(g.NumNodes()))
	borderEdges := 0
	for _, n := range g.Nodes() {
		borderEdges += n.NumEdgesOutCluster()
	}
	borderEdges /= 2
	fmt.Printf("  Number of border edges: %v\n", borderEdges)
	fmt.Printf("  Average border edges per cluster: %.2f\n", float64(2*borderEdges)/float64(g.NumClusters()))
	fmt.Println()
	biggest, bsize := "", 0
	smallest, ssize := "", maxInt
	borderBiggest, borderBsize := "", 0
	borderSmallest, borderSsize := "", maxInt
	for _, c := range g.Clusters() {
		cid := string(c.ClusterID())
		n := c.NumNodes()
		if n > bsize {
			biggest, bsize = cid, n
		}
		if n < ssize {
			smallest, ssize = cid, n
		}

		n = c.NumBorderNodes()
		if n > borderBsize {
			borderBiggest, borderBsize = cid, n
		}
		if n < borderSsize {
			borderSmallest, borderSsize = cid, n
		}
	}
	fmt.Printf("  Cluster with most nodes: %v (%v)\n", biggest, bsize)
	fmt.Printf("  Cluster with fewest nodes: %v (%v)\n", smallest, ssize)
	fmt.Printf("  Cluster with most border nodes: %v (%v)\n", borderBiggest, borderBsize)
	fmt.Printf("  Cluster with fewest border nodes: %v (%v)\n", borderSmallest, borderSsize)
}
