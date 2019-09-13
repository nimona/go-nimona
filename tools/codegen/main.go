package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"golang.org/x/tools/imports"
)

var (
	input  = flag.String("i", "", "input file")
	output = flag.String("o", "-", "output file (default is stdout)")
)

func main() {
	flag.Parse()

	if *input == "" || *output == "" {
		fmt.Fprintf(os.Stdout, "usage: %v <input> <output>\n", os.Args[0])
		os.Exit(1)
	}

	rf, err := os.Open(*input)
	defer rf.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error reading file", err)
		os.Exit(1)
	}

	doc, err := NewParser(bufio.NewReader(rf)).Parse()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error parsing file", err)
		os.Exit(1)
	}

	of, err := os.Create(*output)
	defer of.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error creating file", err)
		os.Exit(1)
	}

	code, err := Generate(doc, *output)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error generating code", err)
		os.Exit(1)
	}

	opt := &imports.Options{
		Comments:  true,
		TabIndent: true,
		TabWidth:  4,
	}
	code, err = imports.Process(*output, []byte(code), opt)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error formatting code", err)
		os.Exit(1)
	}

	if _, err := of.Write(code); err != nil {
		fmt.Fprintln(os.Stderr, "error writing to file", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "complete (%v -> %v)\n", *input, *output)
}
