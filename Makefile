# Makefile for AthenaX

# Binary name
BINARY_NAME=athenax

# Build directory
BUILD_DIR=_bin

# Build target
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) cmd/main.go
	@echo "Binary built successfully: $(BUILD_DIR)/$(BINARY_NAME)"

# Clean target
.PHONY: clean
clean:
	@echo "Cleaning build directory..."
	@rm -rf $(BUILD_DIR)
	@echo "Build directory cleaned"

# Default target
.DEFAULT_GOAL := build 