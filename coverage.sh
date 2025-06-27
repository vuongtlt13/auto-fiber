#!/bin/bash

# Coverage file names
COVERAGE_OUT="coverage.out"
COVERAGE_HTML="coverage.html"

echo "=== AutoFiber Code Coverage Report ==="
echo

# Check if cleanup is requested
if [ "$1" = "clean" ]; then
    echo "Cleaning up coverage files..."
    rm -f "$COVERAGE_OUT" "$COVERAGE_HTML"
    echo "Coverage files cleaned."
    exit 0
fi

# Basic coverage (only main package, excluding example)
echo "1. Basic Coverage:"
go test -cover ./... -coverpkg=github.com/vuongtlt13/auto-fiber
echo

# Detailed function coverage (only main package, excluding example)
echo "2. Function Coverage:"
go test -coverprofile="$COVERAGE_OUT" ./... -coverpkg=github.com/vuongtlt13/auto-fiber
go tool cover -func="$COVERAGE_OUT" | tail -1
echo

# Generate HTML report
echo "3. Generating HTML coverage report..."
go tool cover -html="$COVERAGE_OUT" -o "$COVERAGE_HTML"
echo "HTML report saved to: $COVERAGE_HTML"
echo

# Show uncovered functions (0% coverage)
echo "4. Functions with 0% coverage:"
go tool cover -func="$COVERAGE_OUT" | grep "0.0%"
echo

# Show functions with low coverage (< 50%)
echo "5. Functions with low coverage (< 50%):"
go tool cover -func="$COVERAGE_OUT" | awk '$3 < 50.0 {print}'
echo

echo "=== Coverage Summary ==="
echo "Total coverage: $(go tool cover -func="$COVERAGE_OUT" | tail -1 | awk '{print $3}')"
echo "HTML report: $COVERAGE_HTML"
echo "Raw data: $COVERAGE_OUT"
echo
echo "Note: Example directory is excluded from coverage analysis"
echo "To clean up coverage files: ./coverage.sh clean" 