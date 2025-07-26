# Makefile for AthenaX

# Binary name
BINARY_NAME=athenax
LAMBDA_BINARY_NAME=bootstrap

# Build directory
BUILD_DIR=_bin

# Build target
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) cmd/main.go
	@echo "Binary built successfully: $(BUILD_DIR)/$(BINARY_NAME)"

# Build Lambda function
.PHONY: build-lambda
build-lambda:
	@echo "Building Lambda function..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(LAMBDA_BINARY_NAME) cmd/lambda/main.go
	@echo "Lambda binary built successfully: $(BUILD_DIR)/$(LAMBDA_BINARY_NAME)"

# Package Lambda function
.PHONY: package-lambda
package-lambda: build-lambda
	@echo "Packaging Lambda function..."
	@cd $(BUILD_DIR) && zip -j $(LAMBDA_BINARY_NAME).zip $(LAMBDA_BINARY_NAME)
	@echo "Lambda package created: $(BUILD_DIR)/$(LAMBDA_BINARY_NAME).zip"

# Clean target
.PHONY: clean
clean:
	@echo "Cleaning build directory..."
	@rm -rf $(BUILD_DIR)
	@echo "Build directory cleaned"

# Default target
.DEFAULT_GOAL := build 