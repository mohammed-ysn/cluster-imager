# Variables
GO_CMD=go
GO_BUILD_CMD=$(GO_CMD) build
GO_TEST_CMD=$(GO_CMD) test
BINARY_NAME=cluster-imager
BUILD_DIR=build

# Targets
all: build

build:
	$(GO_BUILD_CMD) -o $(BUILD_DIR)/$(BINARY_NAME) main.go

test:
	$(GO_TEST_CMD) -v ./...

clean:
	rm -rf $(BUILD_DIR)

.PHONY: all build test clean
