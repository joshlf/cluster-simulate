package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	format  = flag.String("format", "edge-list", "the format of the input files")
	input   = flag.String("input", ".", "the directory to search for input files")
	output  = flag.String("output", ".", "the directory to write converted files")
	verbose = flag.Bool("verbose", false, "")
)

const (
	ERR_USAGE = 2 + iota
	ERR_IO
	ERR_SUBCOMMAND
)

var exit_code = 0

func main() {
	flag.Parse()

	inFi, err := os.Stat(*input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error statting input directory: %v\n", err)
		os.Exit(ERR_IO)
	}
	if !inFi.IsDir() {
		fmt.Fprintf(os.Stderr, "Input directory is not a directory\n")
		os.Exit(ERR_IO)
	}

	dir, err := ioutil.ReadDir(*input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing input directory: %v\n", err)
		os.Exit(ERR_IO)
	}
	outFi, err := os.Stat(*output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error statting output directory: %v\n", err)
		os.Exit(ERR_IO)
	}

	if !inFi.IsDir() {
		fmt.Fprintf(os.Stderr, "Output directory is not a directory\n")
		os.Exit(ERR_IO)
	}

	if os.SameFile(inFi, outFi) {
		fmt.Fprintf(os.Stderr, "Output and input directories cannot be the same\n")
		os.Exit(ERR_USAGE)
	}

	_, err = exec.LookPath("convert")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running convert: %v\n", err)
		os.Exit(ERR_IO)
	}
	for _, f := range dir {
		if !f.IsDir() {
			from := filepath.Join(*input, f.Name())
			to := filepath.Join(*output, sanitize(f.Name()))
			if *verbose {
				fmt.Println(from, "->", to)
			}
			convert(from, to)
		}
	}
	os.Exit(exit_code)
}

func convert(from, to string) {
	cmd := exec.Command("convert", "--input", from, "--output", to)
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running \"convert\" subcommand: %v\n", err)
		exit_code = ERR_SUBCOMMAND
	}
}

// assumes filename is a base
// (that is, does not include
// other path components)
func sanitize(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) > 1 {
		parts = parts[:len(parts)-1]
	}
	return strings.Join(append(parts, "def"), ".")
}
