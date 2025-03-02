.PHONY: build clean test run run-api run-grpc docker-up docker-down proto lint

# Build variables
BINARY_NAME_API=api-server
BINARY_NAME_GRPC=grpc-server
BUILD_DIR=bin

# Go variables
GO=go
GOFLAGS=-v

# Docker variables
DOCKER_COMPOSE=docker-compose

# Build both servers
build: build-api build-grpc

# Build API server
build-api:
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME_API) cmd/api/main.go

# Build gRPC server
build-grpc:
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME_GRPC) cmd/grpc/main.go

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.txt

# Run tests
test:
	$(GO) test -v -race -cover ./...

# Run tests with coverage
test-coverage:
	$(GO) test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

# Run API server
run-api:
	$(GO) run cmd/api/main.go

# Run gRPC server
run-grpc:
	$(GO) run cmd/grpc/main.go

# Start Docker containers
docker-up:
	$(DOCKER_COMPOSE) up -d

# Stop Docker containers
docker-down:
	$(DOCKER_COMPOSE) down

# Generate Protocol Buffers code
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/task.proto

# Run go linter
lint:
	golangci-lint run

# Install dependencies
deps:
	$(GO) mod download

# Update dependencies
deps-update:
	$(GO) get -u ./...
	$(GO) mod tidy
