# Build
build:
	go build -o ./tmp/main ./cmd/api

# Run
run:
	go run ./cmd/api

# Dev (hot reload with air)
dev:
	air

# Test
test:
	go test ./...

# Lint
lint:
	golangci-lint run

# Format
fmt:
	go fmt ./...

# Format with goimports
fmt-imports:
	goimports -l -w .

# Format + Lint
check: fmt-imports lint

# Clean build artifacts
clean:
	rm -rf ./tmp

# Install dependencies
deps:
	go mod download
	go mod tidy

# Migration commands
migration-list:
	go run ./cmd/schema -action=list

migration-generate:
	go run ./cmd/schema/main.go -action=generate -name=$(name)

migration-apply:
	go run ./cmd/schema -action=apply
