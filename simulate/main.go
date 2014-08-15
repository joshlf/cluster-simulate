package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/synful/cluster-simulate/graph"
	"github.com/synful/cluster-simulate/graph/encoding"
)

const (
	ERR_USAGE = 2 + iota
	ERR_IO
	ERR_PARSE
)

var (
	graphFilename = flag.String("graph", "", "a file containing the graph to cluster")
	outputDir     = flag.String("output", ".", "a directory to write graph state files after each round")
)

func main() {
	// Let GC run in its own thread
	runtime.GOMAXPROCS(2)

	flag.Parse()

	if *graphFilename == "" {
		fmt.Fprintf(os.Stderr, "No graph file specified\n")
		os.Exit(ERR_USAGE)
	}
	f, err := os.Open(*graphFilename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening graph file: %v\n", err)
		os.Exit(ERR_IO)
	}

	fi, err := os.Stat(*outputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Bad output directory: %v\n", err)
		os.Exit(ERR_IO)
	}

	if !fi.IsDir() {
		fmt.Fprintf(os.Stderr, "Bad output directory: not a directory\n", err)
		os.Exit(ERR_IO)
	}

	gd, err := ioutil.ReadAll(f)
	def, err := encoding.Unmarshal(gd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing graph file: %v\n", err)
		os.Exit(ERR_PARSE)
	}

	g := graph.NewGraph(def, graph.MaxCost)

	var t0, tprev time.Time
	round := 0
	roundFunc := func() {
		if round == 0 {
			// This is the first time we're called

			t0 = time.Now()
			tprev = t0

			if err := writeLogfile(g, round); err != nil {
				fmt.Fprintf(os.Stderr, "%v", err)
				return
			}
			fmt.Printf("ROUND %v...\n", round)

			round++
			return
		}

		t := time.Now()
		diff := t.Sub(tprev)
		tprev = t
		fmt.Printf("Round %v took %v\n", round-1, diff)
		if err := writeLogfile(g, round); err != nil {
			fmt.Fprintf(os.Stderr, "%v", err)
			return
		}
		fmt.Println()
		fmt.Printf("ROUND %v...\n", round)
		round++
	}

	g.MergeCallback(roundFunc, merge)
	t := time.Now()
	diff := t.Sub(tprev)
	fmt.Printf("Round %v took %v\n", round-1, diff)
	if err := writeLogfile(g, round); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		return
	}

	fmt.Println()
	diff = t.Sub(t0)
	fmt.Println("Graph stabilized.")
	fmt.Printf("%v rounds completed in %v\n", round-1, diff)
	if round != 1 {
		fmt.Printf("Average time per round: %v\n", diff/time.Duration(round-1))
	}
}

func merge(c, d graph.ClusterID) {
	if c == d {
		fmt.Printf("Merged %v with itself\n", c)
	} else {
		fmt.Printf("Merged %v with %v\n", c, d)
	}
}

func writeLogfile(g *graph.Graph, round int) error {
	logfile := filepath.Join(*outputDir, fmt.Sprintf("%04d.def", round))
	f, err := os.Create(logfile)
	if err != nil {
		return fmt.Errorf("Error creating logfile: %v", err)
	}
	def := g.GraphDef()
	data, err := encoding.Marshal(def)
	if err != nil {
		return fmt.Errorf("Error marshalling graph def: %v", err)
	}
	_, err = f.Write(data)
	return err
}
