package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"kylix/generator"
	"kylix/lexer"
	"kylix/parser"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	// Command line flags
	outputFile := flag.String("o", "", "Output file name")
	run := flag.Bool("run", false, "Compile and run the program")
	showTokens := flag.Bool("tokens", false, "Show tokens")
	showAST := flag.Bool("ast", false, "Show AST")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Kylix Compiler v0.1.0")
		fmt.Println("Usage: kylix [options] <source.klx>")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		fmt.Println("\nModern Pascal features:")
		fmt.Println("  - Type inference with :=")
		fmt.Println("  - Lambda expressions with ->")
		fmt.Println("  - Generics with <>")
		fmt.Println("  - Async/await support")
		fmt.Println("  - Pattern matching with match")
		fmt.Println("  - Modern exception handling")
		fmt.Println("  - Classes and interfaces")
		fmt.Println("  - Properties with getters/setters")
		os.Exit(1)
	}

	sourceFile := flag.Arg(0)

	// Read source file
	source, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Create lexer
	l := lexer.New(string(source))

	// Show tokens if requested
	if *showTokens {
		fmt.Println("=== Tokens ===")
		for {
			tok := l.NextToken()
			fmt.Printf("%+v\n", tok)
			if tok.Type == "EOF" {
				break
			}
		}
		os.Exit(0)
	}

	// Parse
	p := parser.New(l)
	program := p.ParseProgram()

	// Check for parser errors
	if len(p.Errors()) > 0 {
		fmt.Println("Parser errors:")
		for _, err := range p.Errors() {
			fmt.Printf("  %s\n", err)
		}
		os.Exit(1)
	}

	// Show AST if requested
	if *showAST {
		fmt.Println("=== AST ===")
		fmt.Printf("%+v\n", program)
		os.Exit(0)
	}

	// Generate Go code
	gen := generator.New()
	goCode := gen.Generate(program)

	// Determine output file name
	output := *outputFile
	if output == "" {
		base := filepath.Base(sourceFile)
		output = base[:len(base)-len(filepath.Ext(base))] + ".go"
	}

	// Write Go code to file
	err = ioutil.WriteFile(output, []byte(goCode), 0644)
	if err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated: %s\n", output)

	// Run if requested
	if *run {
		fmt.Println("\n=== Running ===")
		cmd := exec.Command("go", "run", output)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			fmt.Printf("Runtime error: %v\n", err)
			os.Exit(1)
		}
	}
}
