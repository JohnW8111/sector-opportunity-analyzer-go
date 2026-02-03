.PHONY: build run clean test deps frontend

# Build the binary
build: deps
	go build -o sector-analyzer .

# Run the server
run: build
	./sector-analyzer

# Download dependencies
deps:
	go mod download
	go mod tidy

# Clean build artifacts
clean:
	rm -f sector-analyzer
	rm -rf static/*

# Run tests
test:
	go test ./...

# Build frontend from original project and copy to static/
frontend:
	@echo "Building frontend..."
	cd ../sector-opportunity-analyzer/frontend && npm install && npm run build
	@echo "Copying to static/..."
	rm -rf static/*
	cp -r ../sector-opportunity-analyzer/frontend/dist/* static/
	@echo "Frontend ready for embedding"

# Full build with frontend
all: frontend build
	@echo "Build complete. Run with: ./sector-analyzer"

# Development mode - just build and run without frontend
dev: build
	PORT=8000 ./sector-analyzer
