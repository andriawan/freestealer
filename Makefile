.PHONY: test coverage coverage-html coverage-report clean build run help

# Default target
help:
	@echo "Available commands:"
	@echo "  make test           - Run all tests"
	@echo "  make coverage       - Run tests with coverage report"
	@echo "  make coverage-html  - Generate HTML coverage report and open it"
	@echo "  make coverage-report- Show detailed coverage report in terminal"
	@echo "  make clean          - Clean build artifacts and coverage files"
	@echo "  make build          - Build the application"
	@echo "  make run            - Run the application"
	@echo "  make dev            - Run with hot reload (air)"

# Run all tests
test:
	go test ./... -v

# Run tests with coverage
coverage:
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -func=coverage.out

# Generate and open HTML coverage report
coverage-html:
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Detailed coverage report
coverage-report:
	go test ./... -coverprofile=coverage.out -covermode=atomic
	@echo "\n=== Total Coverage ==="
	@go tool cover -func=coverage.out | grep total
	@echo "\n=== Per-File Coverage ==="
	@go tool cover -func=coverage.out

# Clean build artifacts
clean:
	rm -f coverage.out coverage.html
	rm -f freestealer freestealer.exe
	rm -f kasir.db

# Build the application
build:
	go build -o freestealer .

# Run the application
run:
	go run .

# Run with hot reload
dev:
	air
