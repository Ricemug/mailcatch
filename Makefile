# FakeSMTP Build Configuration

APP_NAME = fakesmtp
VERSION = v1.0.0
BUILD_DIR = build
CMD_DIR = cmd/server

# Go build settings
GO_FLAGS = -ldflags "-s -w -X main.version=$(VERSION)"

# Platforms to build for
PLATFORMS = \
	linux/amd64 \
	linux/arm64 \
	darwin/amd64 \
	darwin/arm64 \
	windows/amd64

.PHONY: all clean build build-all test run dev docker docker-build docker-push

# Default target
all: clean build

# Clean build directory
clean:
	rm -rf $(BUILD_DIR)
	mkdir -p $(BUILD_DIR)

# Build for current platform
build:
	go mod tidy
	go build $(GO_FLAGS) -o $(BUILD_DIR)/$(APP_NAME) ./$(CMD_DIR)

# Build for all platforms
build-all: clean
	@echo "Building for all platforms..."
	@for platform in $(PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d'/' -f1); \
		GOARCH=$$(echo $$platform | cut -d'/' -f2); \
		EXT=""; \
		if [ "$$GOOS" = "windows" ]; then EXT=".exe"; fi; \
		OUTPUT="$(BUILD_DIR)/$(APP_NAME)-$$GOOS-$$GOARCH$$EXT"; \
		echo "Building $$OUTPUT..."; \
		GOOS=$$GOOS GOARCH=$$GOARCH go build $(GO_FLAGS) -o $$OUTPUT ./$(CMD_DIR) || exit 1; \
	done
	@echo "Build complete! Binaries are in $(BUILD_DIR)/"

# Run tests
test:
	go test -v ./...

# Run development server
run: build
	./$(BUILD_DIR)/$(APP_NAME)

# Development mode with hot reload
dev:
	go run ./$(CMD_DIR)

# Docker targets
docker: docker-build

# Build Docker image
docker-build:
	docker build -t $(APP_NAME):$(VERSION) -t $(APP_NAME):latest .

# Build multi-platform Docker images
docker-buildx:
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		-t $(APP_NAME):$(VERSION) \
		-t $(APP_NAME):latest \
		--push .

# Push Docker image (customize registry as needed)
docker-push:
	docker push $(APP_NAME):$(VERSION)
	docker push $(APP_NAME):latest

# Create release archives
release: build-all
	@echo "Creating release archives..."
	@cd $(BUILD_DIR) && \
	for binary in $(APP_NAME)-*; do \
		if [[ "$$binary" == *".exe" ]]; then \
			zip "$$binary.zip" "$$binary"; \
		else \
			tar -czf "$$binary.tar.gz" "$$binary"; \
		fi; \
	done
	@echo "Release archives created in $(BUILD_DIR)/"

# Show build info
info:
	@echo "App Name: $(APP_NAME)"
	@echo "Version: $(VERSION)"
	@echo "Build Dir: $(BUILD_DIR)"
	@echo "Supported Platforms:"
	@for platform in $(PLATFORMS); do echo "  - $$platform"; done

# Install dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Show help
help:
	@echo "Available targets:"
	@echo "  all       - Clean and build for current platform"
	@echo "  build     - Build for current platform"
	@echo "  build-all - Build for all supported platforms"
	@echo "  clean     - Clean build directory"
	@echo "  test      - Run tests"
	@echo "  run       - Build and run"
	@echo "  dev       - Run in development mode"
	@echo "  docker    - Build Docker image"
	@echo "  release   - Create release archives"
	@echo "  deps      - Install dependencies"
	@echo "  fmt       - Format code"
	@echo "  lint      - Lint code"
	@echo "  info      - Show build information"
	@echo "  help      - Show this help"