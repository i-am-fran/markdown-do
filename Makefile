.PHONY: build install clean test run dev lint

# Binary name
BINARY=mdd

# Build directory
BUILD_DIR=./build

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Version
VERSION=1.1.0
LDFLAGS=-ldflags "-s -w"

# Main package
MAIN=./cmd/mdd

all: build

# Build the binary
build:
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) $(MAIN)

# Install to GOPATH/bin (defaults to ~/go/bin if GOPATH not set)
INSTALL_DIR ?= $(or $(GOPATH),$(HOME)/go)/bin
install:
	$(GOBUILD) $(LDFLAGS) -o $(INSTALL_DIR)/$(BINARY) $(MAIN)

# Run the application
run:
	$(GORUN) $(MAIN)

# Run in development mode
dev:
	$(GORUN) $(MAIN)

# Run tests
test:
	$(GOTEST) -v ./...

# Format code
fmt:
	$(GOFMT) ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f $(BINARY)

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Cross-compile for multiple platforms
build-all: build-darwin build-linux build-windows

build-darwin:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-darwin-amd64 $(MAIN)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-darwin-arm64 $(MAIN)

build-linux:
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-linux-amd64 $(MAIN)
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-linux-arm64 $(MAIN)

build-windows:
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-windows-amd64.exe $(MAIN)

# Help
help:
	@echo "Available targets:"
	@echo "  build       - Build the binary"
	@echo "  install     - Install to GOPATH/bin"
	@echo "  run         - Run the application"
	@echo "  dev         - Run in development mode"
	@echo "  test        - Run tests"
	@echo "  fmt         - Format code"
	@echo "  lint        - Lint code"
	@echo "  clean       - Clean build artifacts"
	@echo "  deps        - Download dependencies"
	@echo "  build-all   - Cross-compile for all platforms"
	@echo "  help        - Show this help"
