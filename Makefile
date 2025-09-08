.PHONY: build test clean install lint fmt

# Build the binary
build:
	go build -o gitops-validator ./main.go

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -f gitops-validator

# Install dependencies
install:
	go mod download
	go mod tidy

# Run linter
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...

# Run validation on this repository
validate-self:
	./gitops-validator --path . --verbose

# Build and test
all: fmt lint test build

# Create release binary
release:
	GOOS=linux GOARCH=amd64 go build -o gitops-validator-linux-amd64 ./main.go
	GOOS=darwin GOARCH=amd64 go build -o gitops-validator-darwin-amd64 ./main.go
	GOOS=windows GOARCH=amd64 go build -o gitops-validator-windows-amd64.exe ./main.go