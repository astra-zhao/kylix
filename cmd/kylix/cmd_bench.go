package main

import (
	"flag"
	"fmt"
	"kylix/pkg/testrunner"
	"os"
	"path/filepath"
	"strings"
)

func cmdBench(args []string) {
	fs := flag.NewFlagSet("bench", flag.ExitOnError)
	count := fs.Int("count", 5, "Number of iterations per benchmark")
	verbose := fs.Bool("v", false, "Verbose output")
	fs.Usage = func() {
		fmt.Printf(`USAGE: kylix bench [options] [file_bench.klx...]

Discover and run Kylix benchmarks in *_bench.klx files.
Benchmark procedures must be named Bench<Something>() with no parameters.
Results show average time per iteration.

OPTIONS:
`)
		fs.PrintDefaults()
		fmt.Println()
	}
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	runner := testrunner.New(*verbose)

	var files []string
	if fs.NArg() > 0 {
		for _, f := range fs.Args() {
			if strings.HasSuffix(f, "_bench.klx") {
				files = append(files, f)
			} else {
				fmt.Fprintf(os.Stderr, "Warning: %s is not a bench file (must end with _bench.klx)\n", f)
			}
		}
	} else {
		// Auto-discover *_bench.klx in current directory
		entries, err := os.ReadDir(".")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), "_bench.klx") {
				files = append(files, e.Name())
			}
		}
	}

	if len(files) == 0 {
		fmt.Println("No bench files found.")
		return
	}

	// Discover bench procedures (named Bench*)
	var cases []testrunner.TestCase
	for _, f := range files {
		abs, _ := filepath.Abs(f)
		found, err := runner.DiscoverBenches(abs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", f, err)
			continue
		}
		cases = append(cases, found...)
	}

	if len(cases) == 0 {
		fmt.Println("No Bench* procedures found.")
		return
	}

	// Run each benchmark N times and report average
	fmt.Printf("Running %d benchmark(s), %d iteration(s) each...\n\n", len(cases), *count)
	results := runner.RunBench(cases, *count)

	// Print results in Go-bench-compatible format
	for _, r := range results {
		if !r.Passed {
			fmt.Printf("FAIL  %-40s  error: %s\n", r.Name, r.Message)
		} else {
			fmt.Printf("ok    %-40s  %s\n", r.Name, r.BenchResult)
		}
	}
}
