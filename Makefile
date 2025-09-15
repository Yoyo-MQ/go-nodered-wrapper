# Makefile for yoyo-nodered-wrapper

.PHONY: help build test clean lint docs install

# Default target
help:
	@echo "Available targets:"
	@echo "  build     - Build the library and examples"
	@echo "  test      - Run tests"
	@echo "  clean     - Clean build artifacts"
	@echo "  lint      - Run linter"
	@echo "  docs      - Generate documentation"
	@echo "  install   - Install the library"
	@echo "  example   - Run basic example"

# Build the library and examples
build:
	@echo "Building yoyo-nodered-wrapper..."
	@mkdir -p build
	@go build -o build/lib.a ./pkg/...
	@go build -o build/yoyo-nodered-cli ./cmd/example
	@for example in examples/*/; do \
		if [ -f "$$example/main.go" ]; then \
			example_name=$$(basename "$$example"); \
			echo "Building example: $$example_name"; \
			go build -o "build/example-$$example_name" "$$example"; \
		fi; \
	done
	@echo "Build completed!"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf build/
	@go clean

# Run linter
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run; \
	fi

# Generate documentation
docs:
	@echo "Generating documentation..."
	@if command -v godoc >/dev/null 2>&1; then \
		echo "Documentation available at http://localhost:6060"; \
		godoc -http=:6060; \
	else \
		echo "godoc not found, installing..."; \
		go install golang.org/x/tools/cmd/godoc@latest; \
		echo "Documentation available at http://localhost:6060"; \
		godoc -http=:6060; \
	fi

# Install the library
install:
	@echo "Installing library..."
	@go install ./...

# Run basic example
example: build
	@echo "Running basic example..."
	@./build/example-basic

# Run CLI tool
cli: build
	@echo "Running CLI tool..."
	@./build/yoyo-nodered-cli -action=health

# Development setup
dev-setup:
	@echo "Setting up development environment..."
	@go mod tidy
	@go mod download
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@if ! command -v godoc >/dev/null 2>&1; then \
		echo "Installing godoc..."; \
		go install golang.org/x/tools/cmd/godoc@latest; \
	fi
	@echo "Development setup completed!"

# Check dependencies
deps:
	@echo "Checking dependencies..."
	@go mod verify
	@go list -m all

# Update dependencies
deps-update:
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy
