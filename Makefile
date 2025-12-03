.PHONY: build run test clean docker help build-all build-linux build-darwin build-windows

# Build variables
BINARY_NAME=dudu-proxy
BUILD_DIR=build
MAIN=main.go
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Go variables
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION)"
BUILD_FLAGS=-trimpath $(LDFLAGS)

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	@echo "Building $(BINARY_NAME) version $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

run: ## Run the application
	$(GORUN) $(MAIN) -config configs/config.example.json

test: ## Run tests
	$(GOTEST) -v ./...

test-coverage: ## Run tests with coverage
	$(GOTEST) -cover -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

clean: ## Clean build files
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

tidy: ## Tidy go modules
	$(GOMOD) tidy

docker: ## Build Docker image
	docker build -t dudu-proxy:latest .

docker-run: ## Run with Docker
	docker run -p 8080:8080 -p 1080:1080 -v $(PWD)/configs:/app/configs dudu-proxy:latest

dev: tidy ## Development mode (auto rebuild)
	@echo "Starting development mode..."
	$(GORUN) $(MAIN) -config configs/config.example.json

install-deps: ## Install dependencies
	$(GOMOD) download
	$(GOMOD) verify

lint: ## Run linter
	golangci-lint run ./...

fmt: ## Format code
	$(GOCMD) fmt ./...

verify: build ## Run automated verification tests
	@echo "Running automated verification tests..."
	./scripts/verify.sh

all: clean tidy build test ## Clean, tidy, build and test

# Cross-platform build targets
build-all: ## Build for all platforms
	@echo "Building for all platforms..."
	@chmod +x scripts/build.sh
	@./scripts/build.sh $(VERSION)

build-linux: ## Build for Linux (amd64 and arm64)
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	@echo "Building Linux amd64..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$(VERSION)-linux-amd64 $(MAIN)
	@cd $(BUILD_DIR) && zip -q $(BINARY_NAME)-$(VERSION)-linux-amd64.zip $(BINARY_NAME)-$(VERSION)-linux-amd64
	@echo "Building Linux arm64..."
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$(VERSION)-linux-arm64 $(MAIN)
	@cd $(BUILD_DIR) && zip -q $(BINARY_NAME)-$(VERSION)-linux-arm64.zip $(BINARY_NAME)-$(VERSION)-linux-arm64
	@echo "Linux builds complete"

build-darwin: ## Build for macOS (amd64 and arm64)
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	@echo "Building macOS amd64..."
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-amd64 $(MAIN)
	@cd $(BUILD_DIR) && zip -q $(BINARY_NAME)-$(VERSION)-darwin-amd64.zip $(BINARY_NAME)-$(VERSION)-darwin-amd64
	@echo "Building macOS arm64..."
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-arm64 $(MAIN)
	@cd $(BUILD_DIR) && zip -q $(BINARY_NAME)-$(VERSION)-darwin-arm64.zip $(BINARY_NAME)-$(VERSION)-darwin-arm64
	@echo "macOS builds complete"

build-windows: ## Build for Windows (amd64 and arm64)
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	@echo "Building Windows amd64..."
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$(VERSION)-windows-amd64.exe $(MAIN)
	@cd $(BUILD_DIR) && zip -q $(BINARY_NAME)-$(VERSION)-windows-amd64.exe.zip $(BINARY_NAME)-$(VERSION)-windows-amd64.exe
	@echo "Building Windows arm64..."
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$(VERSION)-windows-arm64.exe $(MAIN)
	@cd $(BUILD_DIR) && zip -q $(BINARY_NAME)-$(VERSION)-windows-arm64.exe.zip $(BINARY_NAME)-$(VERSION)-windows-arm64.exe
	@echo "Windows builds complete"
