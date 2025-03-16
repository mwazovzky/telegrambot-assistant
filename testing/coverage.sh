#!/bin/bash

# Run tests with coverage
echo "Running tests with coverage..."
go test -v -race -coverprofile=testing/coverage.out ./...

# Display coverage summary
echo -e "\nCoverage summary:"
go tool cover -func=testing/coverage.out

# Generate HTML report
echo -e "\nGenerating HTML coverage report..."
go tool cover -html=testing/coverage.out -o testing/coverage.html

echo -e "\nCoverage report generated: coverage.html"
echo "Open coverage.html in your browser to view the detailed report" 
