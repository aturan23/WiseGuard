.PHONY: all build test clean docker-build docker-run lint test-unit test-integration test-coverage fmt

# Envs
BINARY_SERVER=bin/server
BINARY_CLIENT=bin/client
DOCKER_COMPOSE=docker-compose.yml
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

# Main commands
all: clean build test

build:
	@echo "Building..."
	@mkdir -p bin
	go build -o $(BINARY_SERVER) ./cmd/server
	go build -o $(BINARY_CLIENT) ./cmd/client

# Testing
test: test-unit test-integration test-coverage

test-unit:
	@echo "Running unit tests..."
	go test -v -race -count=1 ./pkg/...

test-integration:
	@echo "Running integration tests..."
	go test -v -race -count=1 ./integration/...

test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report generated in $(COVERAGE_HTML)"

# Cleaning
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f $(COVERAGE_FILE)
	rm -f $(COVERAGE_HTML)

# Docker commands
docker-build:
	@echo "Building Docker images..."
	docker-compose build

docker-run:
	@echo "Running Docker containers..."
	docker-compose up

docker-stop:
	@echo "Stopping Docker containers..."
	docker-compose down

# Local run
run-server: build
	@echo "Running server locally..."
	./$(BINARY_SERVER)

run-client: build
	@echo "Running client locally..."
	./$(BINARY_CLIENT)

# Linters
lint:
	@echo "Running linters..."
	golangci-lint run ./...

fmt:
	@echo "Formatting code..."
	go fmt ./cmd/... ./pkg/... ./integration/...

# Benchmarks
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Race detector
race:
	@echo "Running race detector..."
	go test -race ./...

# DDoS testing tool
build-ddos:
	@echo "Building DDoS testing tool..."
	go build -o bin/ddos ./tools/ddos

run-ddos: build-ddos
	@echo "Running DDoS simulation..."
	./bin/ddos

run-ddos-heavy: build-ddos
	@echo "Running heavy DDoS simulation..."
	./bin/ddos -clients 1000 -duration 60s

## Protection testing tool
build-protection:
	@echo "Running protection testing tool..."
	go build -o bin/protection ./tools/protection/main.go

run-protection: build-protection
	./bin/protection -attack invalid_pow -clients 50 -duration 10s
	./bin/protection -attack connection_limit -clients 50 -duration 10s
	./bin/protection -attack failed_attempts -clients 50 -duration 10s
	./bin/protection -attack slowloris -clients 50 -duration 10s

# Help
help:
	@echo "Available commands:"
	@echo " make build          - Build the project"
	@echo " make test          - Run all tests with coverage"
	@echo " make test-unit     - Run only unit tests"
	@echo " make test-integration - Run only integration tests"
	@echo " make test-coverage - Generate test coverage report"
	@echo " make clean         - Clean build artifacts"
	@echo " make docker-build  - Build Docker images"
	@echo " make docker-run    - Run Docker containers"
	@echo " make docker-stop   - Stop Docker containers"
	@echo " make run-server    - Run server locally"
	@echo " make run-client    - Run client locally"
	@echo " make lint          - Run linters"
	@echo " make fmt           - Format code"
	@echo " make bench         - Run benchmarks"
	@echo " make race          - Run race detector"
	@echo " make build-ddos     - Build DDoS testing tool"
	@echo " make run-ddos      - Run DDoS simulation"
	@echo " make run-ddos-heavy - Run heavy DDoS simulation"
	@echo " make run-protection - Run protection testing tool"