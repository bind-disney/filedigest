package main

import (
	"flag"
	"fmt"
	"os"
)

func showUsage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	flag.PrintDefaults()
}

func initFlags() {
	flag.StringVar(&inputFileName, "input-file", "", "Path to the input file")
	flag.StringVar(&outputFileName, "output-file", "", "Path to the output file")
	flag.IntVar(&blockSize, "block-size", defaultBlockSize, "Block size in bytes")
	flag.Usage = showUsage

	flag.Parse()
}

func checkFlags() {
	if flag.NFlag() < 2 {
		flag.Usage()
		os.Exit(1)
	}
}
