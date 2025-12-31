.PHONY: all build wasm install clean test help

# Default target
all: build

# Build everything
build: wasm
	@echo "Building CodeBraid MCP server..."
	go build -o codebraid ./cmd/codebraid
	@echo "✓ Build complete: ./codebraid"

# Build WASM sandbox
wasm:
	@echo "Building WASM sandbox..."
	cd pkg/wasm && npm install && npm run build
	@echo "✓ WASM build complete"

# Install the binary to GOPATH/bin
install: build
	@echo "Installing codebraid..."
	go install ./cmd/codebraid
	@echo "✓ Installed to $(shell go env GOPATH)/bin/codebraid"

# Build codegen tool
codegen:
	@echo "Building codegen tool..."
	go build -o codegen ./cmd/codegen
	@echo "✓ Build complete: ./codegen"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f codebraid codegen
	rm -rf pkg/wasm/node_modules
	rm -f pkg/wasm/dist/*.js pkg/wasm/dist/*.map
	go clean
	@echo "✓ Clean complete"

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run with default config
run: build
	./codebraid -config codebraid.json

# Run in stdio mode
run-stdio: build
	./codebraid -transport stdio -config codebraid.json

# Run in http mode
run-http: build
	./codebraid -transport http -port 3000 -config codebraid.json

# Docker targets
docker-build:
	@echo "Building Docker image..."
	docker build -t codebraid-mcp:latest .
	@echo "✓ Docker image built: codebraid-mcp:latest"

docker-run: docker-build
	@echo "Running Docker container..."
	docker run -p 3000:3000 -v $(PWD)/codebraid.json:/etc/codebraid/config.json:ro codebraid-mcp:latest

docker-up:
	@echo "Starting with docker-compose..."
	docker-compose up -d
	@echo "✓ Container started"

docker-down:
	@echo "Stopping docker-compose..."
	docker-compose down
	@echo "✓ Container stopped"

docker-logs:
	docker-compose logs -f

# Show help
help:
	@echo "CodeBraid MCP Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make                Build everything (wasm + go binary)"
	@echo "  make build          Build CodeBraid server"
	@echo "  make wasm           Build WASM sandbox only"
	@echo "  make codegen        Build codegen tool"
	@echo "  make install        Install codebraid to GOPATH/bin"
	@echo "  make clean          Remove build artifacts"
	@echo "  make test           Run tests"
	@echo "  make run            Run server with default config"
	@echo "  make run-stdio      Run in stdio mode"
	@echo "  make run-http       Run in HTTP mode on port 3000"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-build   Build Docker image"
	@echo "  make docker-run     Build and run in Docker"
	@echo "  make docker-up      Start with docker-compose"
	@echo "  make docker-down    Stop docker-compose"
	@echo "  make docker-logs    View container logs"
	@echo ""
	@echo "  make help           Show this help message"
