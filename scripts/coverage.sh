#!/bin/bash
# Coverage Report Generator for FreeStealer (Linux/Mac)
# Run this script to generate comprehensive coverage reports

echo "=== FreeStealer Coverage Report ==="
echo ""

# Clean old coverage files
echo "Cleaning old coverage files..."
rm -f coverage.out coverage.html

# Run tests with coverage
echo "Running tests with coverage..."
go test ./... -coverprofile=coverage.out -covermode=atomic

if [ $? -ne 0 ]; then
    echo "Tests failed!"
    exit 1
fi

echo ""
echo "=== Coverage Summary ==="

# Display total coverage
go tool cover -func=coverage.out | tail -1

echo ""
echo "=== Per-Package Coverage ==="

# Get coverage per package
for pkg in $(go list ./...); do
    pkgName=${pkg#freestealer/}
    coverage=$(go test -cover $pkg 2>&1 | grep "coverage:")
    if [ ! -z "$coverage" ]; then
        echo "$pkgName : $coverage"
    fi
done

# Generate HTML report
echo ""
echo "Generating HTML coverage report..."
go tool cover -html=coverage.out -o coverage.html

echo ""
echo "=== Coverage Report Generated ==="
echo "HTML Report: coverage.html"
echo "Profile Data: coverage.out"
echo ""
echo "To view HTML report, run: open coverage.html (Mac) or xdg-open coverage.html (Linux)"
