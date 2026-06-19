package main

import (
	"flag"
	"fmt"
	"kylix/pkg/testrunner"
	"os"
)

func cmdTest(args []string) {
	fs := flag.NewFlagSet("test", flag.ExitOnError)
	verbose := fs.Bool("v", false, "Verbose output")
	tap := fs.Bool("tap", false, "Output TAP format")
	dir := fs.String("dir", ".", "Directory to search for *_test.klx files")
	filter := fs.String("filter", "", "Run only tests whose name contains this substring")
	fs.Usage = func() {
		fmt.Printf(`USAGE: kylix test [options] [file_test.klx...]

Discover and run Kylix tests in *_test.klx files.
Test procedures must be named Test<Something>() with no parameters.
Use Assert(condition, message) to check expectations.

Optional fixtures (per file):
  procedure Setup;     — runs before each test
  procedure Teardown;  — runs after each test (deferred)

OPTIONS:
`)
		fs.PrintDefaults()
		fmt.Println()
	}
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	runner := testrunner.New(*verbose)
	if *filter != "" {
		runner.SetFilter(*filter)
	}

	var cases []testrunner.TestCase

	if fs.NArg() > 0 {
		// Explicit files provided
		for _, f := range fs.Args() {
			found, err := runner.DiscoverFile(f)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error reading %s: %v\n", f, err)
				os.Exit(1)
			}
			cases = append(cases, found...)
		}
	} else {
		// Auto-discover in dir
		var err error
		cases, err = runner.Discover(*dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error discovering tests: %v\n", err)
			os.Exit(1)
		}
	}

	// Apply filter (if set).
	if *filter != "" {
		cases = runner.FilterCases(cases)
	}

	if len(cases) == 0 {
		if *filter != "" {
			fmt.Printf("no tests matching filter %q\n", *filter)
		} else {
			fmt.Println("no tests found")
		}
		return
	}

	results := runner.Run(cases)

	if *tap {
		testrunner.PrintTAP(results)
	} else {
		passed, failed := testrunner.Summary(results)
		for _, r := range results {
			if r.Passed {
				fmt.Printf("  ok  %s\n", r.Name)
			} else {
				fmt.Printf("  FAIL %s\n", r.Name)
				if r.Message != "" {
					fmt.Printf("       %s\n", r.Message)
				}
			}
		}
		fmt.Printf("\n%d passed, %d failed", passed, failed)
		if *filter != "" {
			fmt.Printf(" (filter: %q)", *filter)
		}
		fmt.Println()
		if failed > 0 {
			os.Exit(1)
		}
	}
}
