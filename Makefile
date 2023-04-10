# Variables
GO=go
BINARY_NAME=cluster-imager
BUILD_DIR=build

# Targets
all: test build
build:
	$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) -v .
test:
	$(GO) test -v ./...
clean:
	$(GO) clean
	rm -rf $(BUILD_DIR)
run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

# Install dependencies
deps:
	$(GO) get -v ./...

# PHONY targets
.PHONY: all build test clean run deps
