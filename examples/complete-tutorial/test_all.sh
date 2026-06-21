#!/bin/bash
# Test all Kylix tutorial examples

KYLIX=${KYLIX:-kylix}
PASS=0
FAIL=0
TOTAL=0

echo "Testing Kylix Tutorial Examples"
echo "================================"
echo ""

for dir in 01_basics 02_control_flow 03_functions 05_generics 06_advanced_types 07_stdlib_core 10_exceptions; do
    echo "Testing $dir..."
    cd "$dir" 2>/dev/null || continue
    
    for f in example*.klx; do
        [ -f "$f" ] || continue
        TOTAL=$((TOTAL + 1))
        
        if $KYLIX build "$f" 2>&1 | grep -q "✓ Compiled"; then
            GOFILE="${f%.klx}.go"
            if [ -f "$GOFILE" ] && go run "$GOFILE" >/dev/null 2>&1; then
                echo "  ✓ $f"
                PASS=$((PASS + 1))
            else
                echo "  ✗ $f (go run failed)"
                FAIL=$((FAIL + 1))
            fi
        else
            echo "  ✗ $f (compile failed)"
            FAIL=$((FAIL + 1))
        fi
    done
    
    cd ..
done

# Test modules separately
echo "Testing 11_modules..."
cd 11_modules 2>/dev/null
if $KYLIX build math_helper.klx example33_use_module.klx 2>&1 | grep -q "✓ Compiled"; then
    if [ -f "main.go" ] && go run main.go >/dev/null 2>&1; then
        echo "  ✓ modules (2 files)"
        PASS=$((PASS + 2))
        TOTAL=$((TOTAL + 2))
    else
        echo "  ✗ modules (go run failed)"
        FAIL=$((FAIL + 2))
        TOTAL=$((TOTAL + 2))
    fi
else
    echo "  ✗ modules (compile failed)"
    FAIL=$((FAIL + 2))
    TOTAL=$((TOTAL + 2))
fi
cd ..

echo ""
echo "================================"
echo "Results: $PASS/$TOTAL passed, $FAIL failed"
echo "================================"
