package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bmatcuk/doublestar"
	"golang.org/x/tools/imports"
)

var (
	all    = flag.String("a", "", "*.idl")
	input  = flag.String("i", "", "input file")
	output = flag.String("o", "-", "output file (default is stdout)")
)

func codegen(in, out string) {
	fmt.Fprintf(os.Stderr, "starting (%v -> %v)\n", in, out)

	rf, err := ioutil.ReadFile(in)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error reading file", err)
		os.Exit(1)
	}

	re := regexp.MustCompile("(?s)//.*?\n|/\\*.*?\\*/")
	rf = re.ReplaceAll(rf, nil)

	doc, err := NewParser(bytes.NewReader(rf)).Parse()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error parsing file", err)
		os.Exit(1)
	}

	of, err := os.Create(out)
	defer of.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error creating file", err)
		os.Exit(1)
	}

	code, err := Generate(doc, out)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error generating code", err)
		os.Exit(1)
	}

	opt := &imports.Options{
		Comments:  true,
		TabIndent: true,
		TabWidth:  4,
	}
	ccode, err := imports.Process(out, []byte(code), opt)
	if err != nil {
		fmt.Println(string(code))
		fmt.Fprintln(os.Stderr, "error formatting code", err)
		os.Exit(1)
	}

	if _, err := of.Write(ccode); err != nil {
		fmt.Fprintln(os.Stderr, "error writing to file", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "complete (%v -> %v)\n", in, out)
}

func main() {
	flag.Parse()

	if *all != "" {
		// d, _ := filepath.Abs(*all)
		d := *all
		fmt.Println(d)
		fs, err := doublestar.Glob(d + "/**/*.ndl")
		if err != nil {
			log.Fatal("could not list files", err)
		}
		if len(fs) == 0 {
			log.Fatal("no files found")
		}
		for _, f := range fs {
			o := strings.Replace(filepath.Base(f), ".ndl", "", 1)
			codegen(f, filepath.Join(filepath.Dir(f), o+"_generated.go"))
		}
	} else if *input != "" || *output != "" {
		codegen(*input, *output)
	}
}
