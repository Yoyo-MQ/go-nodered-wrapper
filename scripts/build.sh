#!/bin/bash

# Build script for yoyo-nodered-wrapper

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed or not in PATH"
    exit 1
fi

print_status "Building yoyo-nodered-wrapper..."

# Get version from git tag or use default
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
print_status "Version: $VERSION"

# Build flags
LDFLAGS="-X main.version=$VERSION -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)"

# Create build directory
mkdir -p build

# Build the library
print_status "Building library..."
go build -ldflags "$LDFLAGS" -o build/lib.a ./pkg/...

# Build examples
print_status "Building examples..."
for example in examples/*/; do
    if [ -f "$example/main.go" ]; then
        example_name=$(basename "$example")
        print_status "Building example: $example_name"
        go build -ldflags "$LDFLAGS" -o "build/example-$example_name" "$example"
    fi
done

# Build CLI tool
print_status "Building CLI tool..."
go build -ldflags "$LDFLAGS" -o build/yoyo-nodered-cli ./cmd/example

# Run tests
print_status "Running tests..."
go test -v ./...

# Run linting
if command -v golangci-lint &> /dev/null; then
    print_status "Running linter..."
    golangci-lint run
else
    print_warning "golangci-lint not found, skipping linting"
fi

# Generate documentation
print_status "Generating documentation..."
if command -v godoc &> /dev/null; then
    godoc -http=:6060 &
    DOC_PID=$!
    sleep 2
    print_status "Documentation available at http://localhost:6060"
    print_status "Press Ctrl+C to stop the documentation server"
    wait $DOC_PID
else
    print_warning "godoc not found, skipping documentation generation"
fi

print_status "Build completed successfully!"
print_status "Build artifacts:"
ls -la build/
