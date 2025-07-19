# Makefile for the pricing microservice
# Provides commands to build, run, test, benchmark, and clean the project

# Variables
BINARY_NAME=pricing-service
GO=go
GOFLAGS=-v

# Default target
.PHONY: all
all: build

# Build the project and generate the binary
.PHONY: build
build:
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) cmd/main.go

# Run the project directly
.PHONY: run
run:
	$(GO) run $(GOFLAGS) cmd/main.go

# Run all tests with verbose output
.PHONY: test
test:
	$(GO) test ./... $(GOFLAGS) > test_result.out

# Run all benchmarks with memory profiling
.PHONY: bench
bench:
	$(GO) test -bench=. -benchmem ./... > bench_result.out

# Clean up generated files
.PHONY: clean
clean:
	rm -f $(BINARY_NAME)

# Ensure dependencies are installed
.PHONY: deps
deps:
	$(GO) mod tidy