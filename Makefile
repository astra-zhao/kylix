#!/bin/bash
# Kylix Compiler Build Script

echo "Building Kylix Compiler..."
go build -o kylix main.go

if [ $? -eq 0 ]; then
    echo "✓ Build successful!"
    echo ""
    echo "Usage:"
    echo "  ./kylix <source.klx>           # Compile to Go"
    echo "  ./kylix -run <source.klx>      # Compile and run"
    echo "  ./kylix -tokens <source.klx>   # Show tokens"
    echo "  ./kylix -ast <source.klx>      # Show AST"
    echo ""
    echo "Examples:"
    echo "  ./kylix examples/hello.klx"
    echo "  ./kylix -run examples/hello.klx"
else
    echo "✗ Build failed!"
    exit 1
fi
