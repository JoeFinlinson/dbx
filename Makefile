.PHONY: test build clean example

# Build the package
build:
	go build ./...

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -cover ./...

# Clean build artifacts
clean:
	go clean
	rm -f dbx

# Run the example (requires PostgreSQL setup)
example:
	@echo "Make sure you have PostgreSQL running and have run example/schema.sql"
	@echo "Update the connection string in example/main.go first"
	cd example && go run main.go

# Install dependencies
deps:
	go mod tidy
	go mod download

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Generate documentation
docs:
	godoc -http=:6060

# Run benchmarks
bench:
	go test -bench=. ./...

# Check for security vulnerabilities
security:
	gosec ./...

# Build for different platforms
build-all: build
	GOOS=linux GOARCH=amd64 go build -o dbx-linux-amd64 ./...
	GOOS=darwin GOARCH=amd64 go build -o dbx-darwin-amd64 ./...
	GOOS=windows GOARCH=amd64 go build -o dbx-windows-amd64.exe ./... 