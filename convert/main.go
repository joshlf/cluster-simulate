package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/synful/cluster-simulate/graph"
	"github.com/synful/cluster-simulate/graph/encoding"
)

const (
	ERR_FORMAT = 3 + iota
	ERR_IO
	ERR_PARSE
)

var (
	inputFilename  = flag.String("input", "-", "the file to convert")
	outputFilename = flag.String("output", "-", "the destination to write the converted graph")
	inputFormat    = flag.String("format", "edge-list", "the format of the input file; currently only supports \"edge-list\".")
)

var converters = map[string]func(*bytes.Buffer) (graph.GraphDef, error){
	"edge-list": edgeListConverter,
}

func main() {
	flag.Parse()

	converter, ok := converters[*inputFormat]
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown format: %v\n", *inputFormat)
		os.Exit(ERR_FORMAT)
	}

	var input, output *os.File
	if *inputFilename == "-" {
		input = os.Stdin
	} else {
		var err error
		input, err = os.Open(*inputFilename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening input file: %v\n", err)
			os.Exit(ERR_IO)
		}
	}
	if *outputFilename == "-" {
		output = os.Stdout
	} else {
		var err error
		output, err = os.Create(*outputFilename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening output file: %v\n", err)
			os.Exit(ERR_IO)
		}
	}

	data, err := ioutil.ReadAll(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
		os.Exit(ERR_IO)
	}

	def, err := converter(bytes.NewBuffer(data))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing input file: %v\n", err)
		os.Exit(ERR_PARSE)
	}

	data, err = encoding.Marshal(def)

	_, err = output.Write(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		os.Exit(ERR_IO)
	}
}

func edgeListConverter(data *bytes.Buffer) (graph.GraphDef, error) {
	var def graph.GraphDef

	nodes := make(map[graph.NodeID]graph.NodeDef)
	links := make(map[graph.LinkDef]struct{})

	r := bufio.NewReader(data)
	s, err := r.ReadString('\n')
	for ; err == nil || err == io.EOF; s, err = r.ReadString('\n') {
		if err == io.EOF && len(s) == 0 {
			break
		}

		// Slice before newline character
		s = s[:len(s)-1]
		if len(s) == 0 {
			continue
		}

		arr := strings.Split(s, "\t")
		if len(arr) != 2 || arr[0] == "" || arr[1] == "" {
			return def, fmt.Errorf("Malformed line: %q", s)
		}

		a, b := graph.NodeID(arr[0]), graph.NodeID(arr[1])
		if a == b {
			fmt.Fprintf(os.Stderr, "Omitting illegal self-link for node %v\n", a)
			// return def, fmt.Errorf("Illegal self-link for node %v", a)
		}
		nodes[a] = graph.NodeDef{a, graph.ClusterID(a)}
		nodes[b] = graph.NodeDef{b, graph.ClusterID(b)}
		// Make sure only one link is created per edge
		if a < b {
			links[graph.LinkDef{a, b, 1}] = struct{}{}
		}

		if err == io.EOF {
			break
		}
	}
	if err != io.EOF {
		panic(fmt.Sprintf("Got non-EOF error: %v", err))
	}

	def.Nodes = make([]graph.NodeDef, 0, len(nodes))
	def.Links = make([]graph.LinkDef, 0, len(links))
	for _, n := range nodes {
		def.Nodes = append(def.Nodes, n)
	}
	for l, _ := range links {
		def.Links = append(def.Links, l)
	}
	return def, nil
}
