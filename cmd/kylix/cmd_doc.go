package main

import (
	"flag"
	"fmt"
	"kylix/pkg/docgen"
	"os"
	"path/filepath"
	"strings"
)

func cmdDoc(args []string) {
	fs := flag.NewFlagSet("doc", flag.ExitOnError)
	outDir := fs.String("out", "docs/api", "Output directory for generated Markdown files")
	stdout := fs.Bool("stdout", false, "Print to stdout instead of writing files")
	fs.Usage = func() {
		fmt.Printf(`USAGE: kylix doc [options] [file.klx...]

Generate Markdown documentation from Kylix source files.
Doc comments are // lines immediately preceding declarations.

OPTIONS:
`)
		fs.PrintDefaults()
		fmt.Println()
	}
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	var files []string

	if fs.NArg() > 0 {
		files = fs.Args()
	} else {
		// Auto-discover .klx files in current directory
		entries, err := os.ReadDir(".")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".klx") {
				files = append(files, e.Name())
			}
		}
		if len(files) == 0 {
			fmt.Println("No .klx files found in current directory.")
			return
		}
	}

	if !*stdout {
		if err := os.MkdirAll(*outDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output dir: %v\n", err)
			os.Exit(1)
		}
	}

	for _, f := range files {
		doc, err := docgen.GenerateFile(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", f, err)
			continue
		}

		md := docgen.RenderMarkdown(doc)

		if *stdout {
			fmt.Printf("=== %s ===\n%s\n", f, md)
			continue
		}

		outFile := filepath.Join(*outDir, doc.Name+".md")
		if err := os.WriteFile(outFile, []byte(md), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", outFile, err)
			continue
		}
		fmt.Printf("  wrote %s\n", outFile)
	}

	if !*stdout {
		fmt.Printf("✓ Documentation generated in %s/\n", *outDir)
	}
}
