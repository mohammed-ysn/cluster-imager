.PHONY: all build test lint clean run docker-build docker-run

# Variables
BINARY_NAME=cluster-imager
DOCKER_IMAGE=cluster-imager
GO=go
GOFLAGS=-v
LDFLAGS=-w -s

all: test build

build:
	@echo "Building binary..."
	CGO_ENABLED=0 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) .

test:
	@echo "Running tests..."
	$(GO) test $(GOFLAGS) -race -cover ./...

test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test $(GOFLAGS) -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

benchmark:
	@echo "Running benchmarks..."
	$(GO) test -bench=. -benchmem ./...

lint:
	@echo "Running linter..."
	golangci-lint run

fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	$(GO) mod tidy

clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

run: build
	@echo "Running application..."
	./$(BINARY_NAME)

docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):latest .

docker-run: docker-build
	@echo "Running Docker container..."
	docker run -p 8080:8080 $(DOCKER_IMAGE):latest

# Development helpers
dev:
	@echo "Starting development server with hot reload..."
	air

install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/cosmtrek/air@latest

# CI/CD helpers
ci-test:
	$(GO) test -v -race -coverprofile=coverage.out ./...

ci-lint:
	golangci-lint run --deadline=5m

# Help
help:
	@echo "Available targets:"
	@echo "  make build          - Build the binary"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make benchmark      - Run benchmarks"
	@echo "  make lint           - Run linter"
	@echo "  make fmt            - Format code"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make run            - Build and run the application"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make docker-run     - Build and run Docker container"
	@echo "  make install-tools  - Install development tools"
	@echo "  make help           - Show this help message"