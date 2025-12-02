.PHONY: build run test clean docker help

# Build variables
BINARY_NAME=dudu-proxy
BUILD_DIR=build
MAIN=main.go

# Go variables
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOMOD=$(GOCMD) mod

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

run: ## Run the application
	$(GORUN) $(MAIN_PATH) -config configs/config.example.json

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
	$(GORUN) $(MAIN_PATH) -config configs/config.example.json

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
