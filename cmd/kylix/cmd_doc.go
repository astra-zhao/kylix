package main

import (
	"flag"
	"fmt"
	"kylix/pkg/docgen"
	"kylix/pkg/openapi"
	"os"
	"path/filepath"
	"strings"
)

func cmdDoc(args []string) {
	fs := flag.NewFlagSet("doc", flag.ExitOnError)
	outDir := fs.String("out", "docs/api", "Output directory for generated Markdown files")
	stdout := fs.Bool("stdout", false, "Print to stdout instead of writing files")
	genOpenAPI := fs.Bool("openapi", false, "Generate OpenAPI 3.1 YAML instead of Markdown")
	apiTitle := fs.String("title", "", "API title for OpenAPI output (default: derived from project)")
	apiVersion := fs.String("api-version", "1.0.0", "API version for OpenAPI output")
	fs.Usage = func() {
		fmt.Printf(`USAGE: kylix doc [options] [file.klx...]

Generate documentation from Kylix source files.

Without --openapi: generates Markdown from doc comments.
With    --openapi: generates OpenAPI 3.1 YAML from KylixBoot annotations.

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

	if *genOpenAPI {
		cmdDocOpenAPI(files, *stdout, *outDir, *apiTitle, *apiVersion)
		return
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

func cmdDocOpenAPI(files []string, stdout bool, outDir, title, version string) {
	doc, err := openapi.Generate(files, title, version)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	yaml := openapi.RenderYAML(doc)

	if stdout {
		fmt.Print(yaml)
		return
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output dir: %v\n", err)
		os.Exit(1)
	}

	outFile := filepath.Join(outDir, "openapi.yaml")
	if err := os.WriteFile(outFile, []byte(yaml), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", outFile, err)
		os.Exit(1)
	}
	fmt.Printf("✓ OpenAPI 3.1 spec written to %s\n", outFile)
}
