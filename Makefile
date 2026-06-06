# ── Variables ──────────────────────────────────────────────────────────────
APP_NAME    := go-devops-app
CMD_PATH    := ./cmd/server
BINARY      := ./server
GO          := go
PORT        := 8080

# ── Default target (runs when you just type 'make') ────────────────────────
.DEFAULT_GOAL := help

# ── Help ───────────────────────────────────────────────────────────────────
help: ## Show available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

# ── Development ────────────────────────────────────────────────────────────
run: ## Run the app directly (no build step)
	$(GO) run $(CMD_PATH)/main.go

build: ## Compile the app into a binary
	$(GO) build -o $(BINARY) $(CMD_PATH)

clean: ## Remove compiled binary
	rm -f $(BINARY)

# ── Quality ────────────────────────────────────────────────────────────────
test: ## Run all unit tests
	$(GO) test ./... -v -count=1

test-coverage: ## Run tests and show coverage percentage
	$(GO) test ./... -coverprofile=coverage.out
	$(GO) tool cover -func=coverage.out

lint: ## Run golangci-lint
	golangci-lint run ./...

fmt: ## Format all Go code
	$(GO) fmt ./...

vet: ## Run Go's built-in static analyzer
	$(GO) vet ./...

# ── Combined Checks ────────────────────────────────────────────────────────
check: fmt vet lint test ## Run all quality checks (format, vet, lint, test)

# ── Docker ─────────────────────────────────────────────────────────────────
docker-build: ## Build Docker image
	docker build -t $(APP_NAME):latest .

docker-run: ## Run app in Docker container
	docker run -p $(PORT):$(PORT) $(APP_NAME):latest

# ── Dependencies ───────────────────────────────────────────────────────────
deps: ## Download all Go dependencies
	$(GO) mod download

tidy: ## Remove unused dependencies
	$(GO) mod tidy

.PHONY: help run build clean test test-coverage lint fmt vet check docker-build docker-run deps tidy
