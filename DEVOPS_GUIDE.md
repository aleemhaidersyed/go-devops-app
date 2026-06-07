# 📘 DevOps From Zero to Production
### A Complete Step-by-Step Implementation Guide

> **By Aleem Haider** | Built with Go · Docker · GitHub Actions · Terraform · AWS · Prometheus · Grafana

---

## 📋 Table of Contents

1. [Introduction & Architecture](#introduction)
2. [Phase 1 — Environment Setup](#phase-1)
3. [Phase 2 — Go REST API](#phase-2)
4. [Phase 3 — Git & GitHub](#phase-3)
5. [Phase 4 — Docker](#phase-4)
6. [Phase 5 — GitHub Actions CI/CD](#phase-5)
7. [Phase 6 — AWS Infrastructure with Terraform](#phase-6)
8. [Phase 7 — Monitoring with Prometheus & Grafana](#phase-7)
9. [Phase 8 & 9 — Security & Production Hardening](#phase-8-9)
10. [Phase 10 — Demo & Cleanup](#phase-10)
11. [Quick Reference Card](#quick-reference)

---

<a name="introduction"></a>
## 🎯 Introduction

This guide takes you from a blank Windows laptop to a **fully deployed, monitored, and secured production application on AWS** — step by step, command by command.

### What You Will Build

```
Internet
    │
    ▼
GitHub Actions (CI/CD Pipeline)
    │  lint → test → scan → build → push
    ▼
Docker Hub / Amazon ECR (Container Registry)
    │
    ▼
AWS Application Load Balancer (public entry point)
    │
    ▼
AWS ECS Fargate (your Go container — auto-managed)
    │
    ▼
AWS RDS PostgreSQL (database — private)
    │
Prometheus + Grafana (monitoring dashboard)
```

### Technology Stack

| Tool | Purpose |
|------|---------|
| **Go 1.24** | Programming language for the REST API |
| **Docker** | Package the app into a portable container |
| **GitHub Actions** | Automate testing, scanning, and deployment |
| **Terraform** | Define AWS infrastructure as code |
| **AWS ECS Fargate** | Run Docker containers without managing servers |
| **AWS RDS** | Managed PostgreSQL database |
| **AWS ALB** | Load balancer that routes traffic to containers |
| **Prometheus** | Collect and store application metrics |
| **Grafana** | Visualize metrics in dashboards |

### Prerequisites

- Windows 11 with WSL2 (Ubuntu 24.04) installed
- GitHub account (free)
- Docker Hub account (free)
- AWS account (free tier)
- Basic familiarity with terminal commands

---

<a name="phase-1"></a>
## ⚙️ PHASE 1 — Environment Setup

> **Goal:** Install all the tools you need on WSL2 Ubuntu so you can develop, build, and deploy a production application.

### 1.1 — Install Go 1.24

Go is the programming language we use to build our API.

```bash
# Download Go 1.24
wget https://go.dev/dl/go1.24.0.linux-amd64.tar.gz

# Remove any old Go installation and extract the new one
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.24.0.linux-amd64.tar.gz

# Add Go to your PATH so you can run 'go' commands anywhere
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verify installation
go version
# Expected: go version go1.24.0 linux/amd64
```

### 1.2 — Install Docker Engine

Docker lets you build and run containers.

```bash
# Install prerequisites
sudo apt-get update
sudo apt-get install -y ca-certificates curl gnupg lsb-release

# Add Docker's official GPG key
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | \
  sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg

# Add Docker repository
echo "deb [arch=$(dpkg --print-architecture) \
  signed-by=/etc/apt/keyrings/docker.gpg] \
  https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# Install Docker
sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io \
  docker-buildx-plugin docker-compose-plugin

# Allow running Docker without sudo
sudo usermod -aG docker $USER
newgrp docker

# Start Docker service
sudo service docker start

# Verify
docker --version
docker compose version
```

### 1.3 — Install AWS CLI

AWS CLI lets you control your AWS account from the terminal.

```bash
# Download and install
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
sudo ./aws/install

# Verify
aws --version
# Expected: aws-cli/2.x.x
```

### 1.4 — Install Terraform

Terraform creates and manages AWS infrastructure from code files.

```bash
# Add HashiCorp GPG key and repository
wget -O- https://apt.releases.hashicorp.com/gpg | \
  sudo gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg

echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] \
  https://apt.releases.hashicorp.com $(lsb_release -cs) main" | \
  sudo tee /etc/apt/sources.list.d/hashicorp.list

sudo apt update && sudo apt install -y terraform

# Verify
terraform version
```

### 1.5 — Install golangci-lint

A code quality tool that runs many linters at once.

```bash
# Install via official script
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
  sh -s -- -b $(go env GOPATH)/bin latest

# Add GOPATH to PATH
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
source ~/.bashrc

# Verify
golangci-lint version
```

### 1.6 — Configure Git

```bash
git config --global user.name "Your Name"
git config --global user.email "your@email.com"
git config --global init.defaultBranch main
```

### 1.7 — Increase WSL2 Memory (Recommended)

Open **PowerShell on Windows** (not WSL) and run:

```powershell
notepad "$env:USERPROFILE\.wslconfig"
```

Add this content and save:
```ini
[wsl2]
memory=4GB
processors=4
swap=2GB
```

Restart WSL2:
```powershell
wsl --shutdown
```

Then reopen Ubuntu and start Docker:
```bash
sudo service docker start
```

---

<a name="phase-2"></a>
## 🔨 PHASE 2 — Go REST API

> **Goal:** Build a production-quality REST API with health checks, task management, metrics, structured logging, and unit tests.

### 2.1 — Create Project Structure

```bash
mkdir -p ~/projects/go-devops-app
cd ~/projects/go-devops-app

# Create the standard Go project directory layout
mkdir -p cmd/server
mkdir -p internal/handlers
mkdir -p internal/middleware
mkdir -p monitoring
```

### 2.2 — Initialize Go Module

```bash
go mod init github.com/YOUR_GITHUB_USERNAME/go-devops-app
```

> Replace `YOUR_GITHUB_USERNAME` with your actual GitHub username.

### 2.3 — Install Dependencies

```bash
# chi — lightweight HTTP router
go get github.com/go-chi/chi/v5

# zerolog — fast, structured JSON logger
go get github.com/rs/zerolog

# prometheus — exposes /metrics endpoint
go get github.com/prometheus/client_golang/prometheus/promhttp
```

### 2.4 — Create the Health Handler

**File:** `internal/handlers/health.go`

```go
package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

// HealthResponse is what we send back when someone hits /health
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Service   string `json:"service"`
}

// HealthHandler responds to GET /health
// Used by the load balancer to check if our app is alive
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Service:   "go-devops-app",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
```

### 2.5 — Create the Tasks Handler

**File:** `internal/handlers/tasks.go`

```go
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
)

// Task represents a single to-do item
type Task struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	CreatedAt string `json:"created_at"`
}

// TaskStore holds all tasks in memory
// sync.Mutex prevents data corruption when multiple requests come at once
type TaskStore struct {
	mu     sync.Mutex
	tasks  []Task
	nextID int
}

// NewTaskStore creates a new empty task store
func NewTaskStore() *TaskStore {
	return &TaskStore{nextID: 1, tasks: []Task{}}
}

// GetTasks handles GET /tasks — returns all tasks as JSON
func (s *TaskStore) GetTasks(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(s.tasks); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// CreateTask handles POST /tasks — creates a new task
func (s *TaskStore) CreateTask(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title string `json:"title"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	task := Task{
		ID:        s.nextID,
		Title:     input.Title,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	s.tasks = append(s.tasks, task)
	s.nextID++
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(task); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// DeleteTask handles DELETE /tasks/{id} — removes a task by ID
func (s *TaskStore) DeleteTask(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid task ID", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for i, t := range s.tasks {
		if t.ID == id {
			s.tasks = append(s.tasks[:i], s.tasks[i+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	http.Error(w, "task not found", http.StatusNotFound)
}
```

### 2.6 — Create the Logger Middleware

**File:** `internal/middleware/logger.go`

```go
package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// responseWriter wraps http.ResponseWriter to capture the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Logger is an HTTP middleware that logs every request
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := newResponseWriter(w)
		next.ServeHTTP(rw, r)

		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", rw.statusCode).
			Dur("duration", time.Since(start)).
			Msg("request")
	})
}
```

### 2.7 — Create the Main Server

**File:** `cmd/server/main.go`

```go
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/YOUR_GITHUB_USERNAME/go-devops-app/internal/handlers"
	"github.com/YOUR_GITHUB_USERNAME/go-devops-app/internal/middleware"
)

func main() {
	// Configure Logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Info().Msg("starting go-devops-app")

	// Create Router
	r := chi.NewRouter()

	// Register Global Middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.Logger)

	// Create Shared State
	store := handlers.NewTaskStore()

	// Register Routes
	r.Get("/health", handlers.HealthHandler)
	r.Route("/tasks", func(r chi.Router) {
		r.Get("/", store.GetTasks)
		r.Post("/", store.CreateTask)
		r.Delete("/{id}", store.DeleteTask)
	})
	r.Handle("/metrics", promhttp.Handler())

	// Server with explicit timeouts
	port := ":8080"
	srv := &http.Server{
		Addr:         port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in background
	go func() {
		log.Info().Str("port", port).Msg("server listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server failed")
		}
	}()

	// Graceful Shutdown — catches Ctrl+C and ECS SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("shutdown signal received — draining connections")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("forced shutdown after timeout")
	}
	log.Info().Msg("server exited cleanly")
}
```

### 2.8 — Create Unit Tests

**File:** `internal/handlers/tasks_test.go`

```go
package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestGetTasksEmpty(t *testing.T) {
	store := NewTaskStore()
	r := chi.NewRouter()
	r.Get("/tasks", store.GetTasks)

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var tasks []Task
	if err := json.NewDecoder(rr.Body).Decode(&tasks); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("expected empty list, got %d tasks", len(tasks))
	}
}

func TestCreateTask(t *testing.T) {
	store := NewTaskStore()
	r := chi.NewRouter()
	r.Post("/tasks", store.CreateTask)

	body := bytes.NewBufferString(`{"title":"test task"}`)
	req := httptest.NewRequest(http.MethodPost, "/tasks", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rr.Code)
	}

	var task Task
	if err := json.NewDecoder(rr.Body).Decode(&task); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if task.Title != "test task" {
		t.Errorf("expected 'test task', got '%s'", task.Title)
	}
}

func TestCreateTaskMissingTitle(t *testing.T) {
	store := NewTaskStore()
	r := chi.NewRouter()
	r.Post("/tasks", store.CreateTask)

	body := bytes.NewBufferString(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/tasks", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestDeleteTask(t *testing.T) {
	store := NewTaskStore()
	store.tasks = []Task{{ID: 1, Title: "to delete"}}
	store.nextID = 2

	r := chi.NewRouter()
	r.Delete("/tasks/{id}", store.DeleteTask)

	req := httptest.NewRequest(http.MethodDelete, "/tasks/1", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
}

func TestDeleteTaskNotFound(t *testing.T) {
	store := NewTaskStore()
	r := chi.NewRouter()
	r.Delete("/tasks/{id}", store.DeleteTask)

	req := httptest.NewRequest(http.MethodDelete, "/tasks/999", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}
```

### 2.9 — Create the Makefile

**File:** `Makefile`

```makefile
.PHONY: build run test lint check clean docker-build docker-run

## build: Compile the Go binary
build:
	go build -o server ./cmd/server/

## run: Run the server locally
run:
	go run ./cmd/server/

## test: Run all unit tests with race detector
test:
	go test ./... -v -count=1

## lint: Run golangci-lint (use --concurrency=1 for WSL2)
lint:
	golangci-lint run ./... --concurrency=1

## check: Run lint + tests together
check: lint test

## clean: Remove compiled binary
clean:
	rm -f server

## docker-build: Build Docker image
docker-build:
	docker build -t go-devops-app:latest .

## docker-run: Run container locally
docker-run:
	docker run -p 8080:8080 go-devops-app:latest
```

### 2.10 — Run and Test Locally

```bash
# Download all dependencies
go mod tidy

# Run tests
make test

# Start server (press Ctrl+C to stop)
make run

# In another terminal, test the API
curl http://localhost:8080/health
curl -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -d '{"title": "my first task"}'
curl http://localhost:8080/tasks
```

---

<a name="phase-3"></a>
## 🗂️ PHASE 3 — Git & GitHub

> **Goal:** Version control your code and set up professional GitHub workflows with branch protection.

### 3.1 — Create .gitignore

**File:** `.gitignore`

```gitignore
# Compiled binary (leading / = only root level, NOT cmd/server/)
/server
*.exe
*.out

# Go workspace files
go.work
go.work.sum

# Test artifacts
*.test
*.prof
coverage.out

# Environment files (contain secrets)
.env
*.env

# Terraform files (contain secrets + auto-generated)
terraform/.terraform/
terraform/*.tfstate*
terraform/*.tfstate.backup
terraform/.terraform.lock.hcl
terraform/terraform.tfvars

# IDE files
.vscode/
.idea/
*.swp
```

### 3.2 — Initialize Git Repository

```bash
cd ~/projects/go-devops-app
git init
git add .
git commit -m "feat: initial project setup with Go REST API"
```

### 3.3 — Connect to GitHub

**Generate SSH key (if you don't have one):**
```bash
ssh-keygen -t ed25519 -C "your@email.com"
# Press Enter for all prompts
```

**Display your public key:**
```bash
cat ~/.ssh/id_ed25519.pub
```

**Add to GitHub:**
1. Go to github.com → Settings → SSH and GPG keys
2. Click "New SSH key"
3. Paste the key → Save

**Test connection:**
```bash
ssh -T git@github.com
# Expected: Hi username! You've successfully authenticated.
```

### 3.4 — Push to GitHub

```bash
# Create a new EMPTY repo on github.com first, then:
git remote add origin git@github.com:YOUR_USERNAME/go-devops-app.git
git push -u origin main
```

### 3.5 — Enable Branch Protection

On GitHub:
1. **Settings** → **Branches** → **Add branch ruleset**
2. Target: `main`
3. Check: **"Require a pull request before merging"**
4. Save

> From now on, **never push directly to main**. Always use feature branches and PRs.

### 3.6 — The Branch Workflow (Use Every Time)

```bash
# 1. Create a feature branch
git checkout -b feat/your-feature-name

# 2. Make your changes, then commit
git add .
git commit -m "feat: description of what you did"

# 3. Push the branch
git push origin feat/your-feature-name

# 4. Go to GitHub → create Pull Request → wait for CI → merge

# 5. Sync local main after merge
git checkout main
git pull origin main
```

---

<a name="phase-4"></a>
## 🐳 PHASE 4 — Docker

> **Goal:** Package your Go app into a Docker container and run the full stack (app + database + monitoring) locally.

### 4.1 — Create the Dockerfile

**File:** `Dockerfile`

```dockerfile
# ════════════════════════════════════════════════════════════════
# STAGE 1: Builder
# Uses the full Go SDK to compile the application
# ════════════════════════════════════════════════════════════════
FROM golang:1.24-alpine AS builder

# Install git (required for downloading Go modules)
RUN apk add --no-cache git

WORKDIR /app

# Copy dependency files first (better layer caching)
# Docker only re-downloads modules when go.mod/go.sum change
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Compile the binary
# CGO_ENABLED=0  = no C dependencies (fully static binary)
# GOOS=linux     = target Linux OS
# GOARCH=amd64   = target 64-bit Intel/AMD
# -ldflags="-w -s" = strip debug info (smaller binary)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags="-w -s" \
    -o /app/server \
    ./cmd/server/

# ════════════════════════════════════════════════════════════════
# STAGE 2: Runner
# Minimal Alpine image — no Go SDK, no source code, no build tools
# ════════════════════════════════════════════════════════════════
FROM alpine:3.19 AS runner

# Install security certificates and timezone data
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user for security
# If the app is compromised, attacker has no root access
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy only the compiled binary from the builder stage
COPY --from=builder /app/server ./server

# Set correct ownership
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Document which port the app listens on
EXPOSE 8080

# Health check — Docker will monitor this
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget -q --spider http://localhost:8080/health || exit 1

# Start the server
CMD ["./server"]
```

### 4.2 — Create .dockerignore

**File:** `.dockerignore`

```
# Git history — not needed in container
.git/
.gitignore

# Compiled binary — we recompile inside Docker
# Leading / = only match root-level 'server', NOT cmd/server/
/server

# Terraform infrastructure files
terraform/

# Docker files themselves
Dockerfile
.dockerignore
docker-compose.yml

# Monitoring config
monitoring/

# IDE files
.vscode/
.idea/
```

### 4.3 — Create docker-compose.yml

**File:** `docker-compose.yml`

```yaml
services:

  # ── Our Go Application ────────────────────────────────────────
  app:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: devops-app
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
    restart: unless-stopped
    depends_on:
      db:
        condition: service_healthy
    networks:
      - devops-network
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:8080/health"]
      interval: 30s
      timeout: 5s
      retries: 3

  # ── PostgreSQL Database ───────────────────────────────────────
  db:
    image: postgres:16-alpine
    container_name: devops-db
    environment:
      POSTGRES_USER: devops
      POSTGRES_PASSWORD: devopspass
      POSTGRES_DB: tasksdb
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - devops-network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U devops -d tasksdb"]
      interval: 10s
      timeout: 5s
      retries: 5

  # ── Prometheus ────────────────────────────────────────────────
  prometheus:
    image: prom/prometheus:latest
    container_name: devops-prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
      - ./monitoring/alert_rules.yml:/etc/prometheus/alert_rules.yml
    networks:
      - devops-network
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'

  # ── Grafana ───────────────────────────────────────────────────
  grafana:
    image: grafana/grafana:latest
    container_name: devops-grafana
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin123
      - GF_SECURITY_ADMIN_USER=admin
      - GF_USERS_ALLOW_SIGN_UP=false
    volumes:
      - grafana_data:/var/lib/grafana
      - ./monitoring/grafana/provisioning:/etc/grafana/provisioning
      - ./monitoring/grafana/dashboards:/var/lib/grafana/dashboards
    networks:
      - devops-network
    depends_on:
      - prometheus

volumes:
  postgres_data:
  grafana_data:

networks:
  devops-network:
    driver: bridge
```

### 4.4 — Build and Run

```bash
# Build the Docker image
docker build -t go-devops-app:latest .

# Check image size (should be ~41MB)
docker images go-devops-app

# Start full stack
docker compose up -d

# Check all containers are running
docker compose ps

# Test the app inside Docker
curl http://localhost:8080/health

# View logs
docker compose logs -f app

# Stop everything
docker compose down
```

---

<a name="phase-5"></a>
## 🔄 PHASE 5 — GitHub Actions CI/CD Pipeline

> **Goal:** Automate code quality, security, testing, and Docker image publishing on every commit.

### 5.1 — Create linting config

**File:** `.golangci.yml`

```yaml
run:
  timeout: 5m
  go: '1.24'

linters:
  enable:
    - govet        # Go's built-in bug detector
    - errcheck     # Ensures all errors are handled
    - staticcheck  # Advanced static analysis
    - gosimple     # Suggests code simplifications
    - ineffassign  # Finds useless variable assignments
    - unused       # Finds unused code
    - gofmt        # Enforces Go formatting standard

linters-settings:
  errcheck:
    check-type-assertions: false
```

### 5.2 — Set Up GitHub Secrets

Go to your GitHub repo → **Settings** → **Secrets and variables** → **Actions** → **New repository secret**

Add these two secrets:

| Secret Name | Value |
|------------|-------|
| `DOCKERHUB_USERNAME` | Your Docker Hub username |
| `DOCKERHUB_TOKEN` | Docker Hub access token (created in Docker Hub → Account Settings → Security → New Access Token) |

### 5.3 — Create CI Workflow

**File:** `.github/workflows/ci.yml`

```yaml
name: CI

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

env:
  GO_VERSION: '1.24'
  APP_NAME: go-devops-app
  FORCE_JAVASCRIPT_ACTIONS_TO_NODE24: 'true'

jobs:

  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: latest
          args: --timeout=5m

  test:
    name: Test
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: Download dependencies
        run: go mod download

      - name: Run tests
        run: go test -race -cover -count=1 ./...

      - name: Generate coverage report
        run: |
          go test -coverprofile=coverage.out ./...
          go tool cover -func=coverage.out

      - name: Upload coverage report
        uses: actions/upload-artifact@v4
        with:
          name: coverage-report
          path: coverage.out
          retention-days: 7

  security-scan:
    name: Security Scan
    runs-on: ubuntu-latest
    needs: [lint, test]
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: Build Docker image for scanning
        run: docker build -t ${{ env.APP_NAME }}:scan .

      - name: Scan image with Trivy
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: ${{ env.APP_NAME }}:scan
          format: table
          exit-code: '1'
          ignore-unfixed: true
          severity: CRITICAL

  build-verify:
    name: Build Verify
    runs-on: ubuntu-latest
    needs: [lint, test]
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: Download dependencies
        run: go mod download

      - name: Build binary
        run: |
          CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
          go build -ldflags="-w -s" -o ./server ./cmd/server/

      - name: Verify binary exists
        run: |
          ls -lh ./server
          file ./server
```

### 5.4 — Create CD Workflow

**File:** `.github/workflows/cd.yml`

```yaml
name: CD

on:
  push:
    branches: [ main ]

env:
  GO_VERSION: '1.24'
  APP_NAME: go-devops-app
  FORCE_JAVASCRIPT_ACTIONS_TO_NODE24: 'true'

jobs:

  build-and-push:
    name: Build and Push Docker Image
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Extract Docker metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ secrets.DOCKERHUB_USERNAME }}/${{ env.APP_NAME }}
          tags: |
            type=raw,value=latest,enable={{is_default_branch}}
            type=sha,prefix=sha-,format=short

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Login to Docker Hub
        uses: docker/login-action@v3
        with:
          username: ${{ secrets.DOCKERHUB_USERNAME }}
          password: ${{ secrets.DOCKERHUB_TOKEN }}

      - name: Build and push Docker image
        uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

      - name: Print image digest
        run: |
          echo "Image pushed successfully!"
          echo "Tags: ${{ steps.meta.outputs.tags }}"
```

### 5.5 — Push and Watch Pipeline

```bash
git checkout -b feat/github-actions-cicd
git add .github/workflows/ci.yml .github/workflows/cd.yml .golangci.yml
git commit -m "feat: add GitHub Actions CI/CD pipeline"
git push origin feat/github-actions-cicd
```

Go to GitHub → **Actions** tab → watch 4 jobs run automatically!

---

<a name="phase-6"></a>
## ☁️ PHASE 6 — AWS Infrastructure with Terraform

> **Goal:** Create all AWS infrastructure using code so it's reproducible, version-controlled, and destroyable with one command.

### 6.1 — Create IAM User for Terraform

1. AWS Console → **IAM** → **Users** → **Create user**
2. Username: `terraform-devops`
3. Attach these managed policies:
   - `AmazonEC2FullAccess`
   - `AmazonECS_FullAccess`
   - `AmazonRDSFullAccess`
   - `ElasticLoadBalancingFullAccess`
   - `AmazonVPCFullAccess`
   - `EC2ContainerRegistryFullAccess`
   - `IAMFullAccess`
   - `AmazonS3FullAccess`
   - `AmazonDynamoDBFullAccess`
4. Create user → **Security credentials** → **Create access key** → CLI → copy both keys

**Add inline policies** (to avoid the 10-policy limit):

```bash
# CloudWatch Logs
aws iam put-user-policy \
  --user-name terraform-devops \
  --policy-name cloudwatch-logs-extra \
  --policy-document '{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"logs:*","Resource":"*"}]}'

# Secrets Manager
aws iam put-user-policy \
  --user-name terraform-devops \
  --policy-name secrets-manager-extra \
  --policy-document '{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"secretsmanager:*","Resource":"*"}]}'

# CloudTrail
aws iam put-user-policy \
  --user-name terraform-devops \
  --policy-name cloudtrail-extra \
  --policy-document '{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"cloudtrail:*","Resource":"*"}]}'
```

### 6.2 — Configure AWS CLI

```bash
aws configure
# AWS Access Key ID: paste your key
# AWS Secret Access Key: paste your secret
# Default region name: us-east-1
# Default output format: json

# Verify
aws sts get-caller-identity
```

### 6.3 — Create Remote State Backend (One-Time Setup)

```bash
# Get your account ID
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
BUCKET="go-devops-terraform-state-${ACCOUNT_ID}"

# Create S3 bucket for Terraform state
aws s3api create-bucket --bucket $BUCKET --region us-east-1

# Enable versioning (protects against accidental deletion)
aws s3api put-bucket-versioning \
  --bucket $BUCKET \
  --versioning-configuration Status=Enabled

# Enable encryption
aws s3api put-bucket-encryption \
  --bucket $BUCKET \
  --server-side-encryption-configuration \
    '{"Rules":[{"ApplyServerSideEncryptionByDefault":{"SSEAlgorithm":"AES256"}}]}'

# Block all public access
aws s3api put-public-access-block \
  --bucket $BUCKET \
  --public-access-block-configuration \
    "BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true"

# Create DynamoDB table for state locking
aws dynamodb create-table \
  --table-name terraform-state-lock \
  --attribute-definitions AttributeName=LockID,AttributeType=S \
  --key-schema AttributeName=LockID,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST \
  --region us-east-1

echo "Bucket name: $BUCKET"
```

### 6.4 — Create Terraform Directory Structure

```bash
cd ~/projects/go-devops-app
mkdir -p terraform/modules/{networking,ecr,compute,database,loadbalancer}
```

### 6.5 — Root Terraform Files

**File:** `terraform/providers.tf`

```hcl
terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = "go-devops-app"
      Environment = var.environment
      ManagedBy   = "terraform"
    }
  }
}
```

**File:** `terraform/backend.tf`
> Replace `YOUR_BUCKET_NAME` with your actual bucket name from Step 6.3

```hcl
terraform {
  backend "s3" {
    bucket       = "YOUR_BUCKET_NAME"
    key          = "go-devops-app/terraform.tfstate"
    region       = "us-east-1"
    use_lockfile = true
    encrypt      = true
  }
}
```

**File:** `terraform/variables.tf`

```hcl
variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "dev"
}

variable "app_name" {
  description = "Application name"
  type        = string
  default     = "go-devops-app"
}

variable "docker_image" {
  description = "Docker image URI to deploy"
  type        = string
}

variable "db_password" {
  description = "Database password"
  type        = string
  sensitive   = true
}
```

**File:** `terraform/outputs.tf`

```hcl
output "alb_dns_name" {
  description = "Public URL of the Application Load Balancer"
  value       = module.loadbalancer.alb_dns_name
}

output "ecr_repository_url" {
  description = "ECR repository URL"
  value       = module.ecr.repository_url
}

output "rds_endpoint" {
  description = "RDS connection endpoint"
  value       = module.database.db_endpoint
  sensitive   = true
}

output "ecs_cluster_name" {
  description = "ECS cluster name"
  value       = module.compute.cluster_name
}
```

**File:** `terraform/main.tf`

```hcl
module "ecr" {
  source          = "./modules/ecr"
  repository_name = var.app_name
}

module "networking" {
  source   = "./modules/networking"
  app_name = var.app_name
  vpc_cidr = "10.0.0.0/16"
}

module "loadbalancer" {
  source                = "./modules/loadbalancer"
  app_name              = var.app_name
  vpc_id                = module.networking.vpc_id
  public_subnet_ids     = module.networking.public_subnet_ids
  alb_security_group_id = module.networking.alb_security_group_id
}

module "database" {
  source                = "./modules/database"
  app_name              = var.app_name
  private_subnet_ids    = module.networking.private_subnet_ids
  rds_security_group_id = module.networking.rds_security_group_id
  db_password           = var.db_password
}

module "compute" {
  source                = "./modules/compute"
  app_name              = var.app_name
  aws_region            = var.aws_region
  docker_image          = var.docker_image
  private_subnet_ids    = module.networking.private_subnet_ids
  public_subnet_ids     = module.networking.public_subnet_ids
  ecs_security_group_id = module.networking.ecs_security_group_id
  target_group_arn      = module.loadbalancer.target_group_arn
  alb_listener_arn      = module.loadbalancer.alb_arn
  db_endpoint           = module.database.db_endpoint
  db_name               = module.database.db_name
  db_username           = module.database.db_username
}
```

**File:** `terraform/terraform.tfvars` *(never commit this file!)*

```hcl
aws_region   = "us-east-1"
environment  = "dev"
app_name     = "go-devops-app"
docker_image = "YOUR_DOCKERHUB_USERNAME/go-devops-app:latest"
db_password  = "DevOpsSecure2024"
```

### 6.6 — ECR Module

**File:** `terraform/modules/ecr/main.tf`

```hcl
resource "aws_ecr_repository" "app" {
  name                 = var.repository_name
  image_tag_mutability = "MUTABLE"
  force_delete         = true

  image_scanning_configuration {
    scan_on_push = true
  }
}

resource "aws_ecr_lifecycle_policy" "app" {
  repository = aws_ecr_repository.app.name

  policy = jsonencode({
    rules = [{
      rulePriority = 1
      description  = "Keep last 10 images"
      selection = {
        tagStatus   = "any"
        countType   = "imageCountMoreThan"
        countNumber = 10
      }
      action = { type = "expire" }
    }]
  })
}
```

**File:** `terraform/modules/ecr/variables.tf`

```hcl
variable "repository_name" {
  type = string
}
```

**File:** `terraform/modules/ecr/outputs.tf`

```hcl
output "repository_url" {
  value = aws_ecr_repository.app.repository_url
}

output "repository_arn" {
  value = aws_ecr_repository.app.arn
}
```

### 6.7 — Networking Module

**File:** `terraform/modules/networking/main.tf`

```hcl
resource "aws_vpc" "main" {
  cidr_block           = var.vpc_cidr
  enable_dns_hostnames = true
  enable_dns_support   = true
  tags = { Name = "${var.app_name}-vpc" }
}

resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id
  tags   = { Name = "${var.app_name}-igw" }
}

resource "aws_subnet" "public" {
  count                   = 2
  vpc_id                  = aws_vpc.main.id
  cidr_block              = cidrsubnet(var.vpc_cidr, 8, count.index)
  availability_zone       = data.aws_availability_zones.available.names[count.index]
  map_public_ip_on_launch = true
  tags = {
    Name = "${var.app_name}-public-${count.index + 1}"
    Tier = "public"
  }
}

resource "aws_subnet" "private" {
  count             = 2
  vpc_id            = aws_vpc.main.id
  cidr_block        = cidrsubnet(var.vpc_cidr, 8, count.index + 10)
  availability_zone = data.aws_availability_zones.available.names[count.index]
  tags = {
    Name = "${var.app_name}-private-${count.index + 1}"
    Tier = "private"
  }
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id
  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }
  tags = { Name = "${var.app_name}-public-rt" }
}

resource "aws_route_table_association" "public" {
  count          = 2
  subnet_id      = aws_subnet.public[count.index].id
  route_table_id = aws_route_table.public.id
}

resource "aws_security_group" "alb" {
  name        = "${var.app_name}-alb-sg"
  description = "Security group for ALB"
  vpc_id      = aws_vpc.main.id

  ingress { from_port = 80;  to_port = 80;  protocol = "tcp"; cidr_blocks = ["0.0.0.0/0"] }
  ingress { from_port = 443; to_port = 443; protocol = "tcp"; cidr_blocks = ["0.0.0.0/0"] }
  egress  { from_port = 0;   to_port = 0;   protocol = "-1";  cidr_blocks = ["0.0.0.0/0"] }
  tags = { Name = "${var.app_name}-alb-sg" }
}

resource "aws_security_group" "ecs" {
  name        = "${var.app_name}-ecs-sg"
  description = "Security group for ECS tasks"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port       = 8080
    to_port         = 8080
    protocol        = "tcp"
    security_groups = [aws_security_group.alb.id]
  }
  egress { from_port = 0; to_port = 0; protocol = "-1"; cidr_blocks = ["0.0.0.0/0"] }
  tags = { Name = "${var.app_name}-ecs-sg" }
}

resource "aws_security_group" "rds" {
  name        = "${var.app_name}-rds-sg"
  description = "Security group for RDS"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [aws_security_group.ecs.id]
  }
  tags = { Name = "${var.app_name}-rds-sg" }
}

data "aws_availability_zones" "available" {
  state = "available"
}
```

**File:** `terraform/modules/networking/variables.tf`

```hcl
variable "app_name" { type = string }
variable "vpc_cidr" { type = string; default = "10.0.0.0/16" }
```

**File:** `terraform/modules/networking/outputs.tf`

```hcl
output "vpc_id"                  { value = aws_vpc.main.id }
output "public_subnet_ids"       { value = aws_subnet.public[*].id }
output "private_subnet_ids"      { value = aws_subnet.private[*].id }
output "alb_security_group_id"   { value = aws_security_group.alb.id }
output "ecs_security_group_id"   { value = aws_security_group.ecs.id }
output "rds_security_group_id"   { value = aws_security_group.rds.id }
```

### 6.8 — Load Balancer Module

**File:** `terraform/modules/loadbalancer/main.tf`

```hcl
resource "aws_lb" "main" {
  name               = "${var.app_name}-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [var.alb_security_group_id]
  subnets            = var.public_subnet_ids
  tags               = { Name = "${var.app_name}-alb" }
}

resource "aws_lb_target_group" "app" {
  name        = "${var.app_name}-tg"
  port        = 8080
  protocol    = "HTTP"
  vpc_id      = var.vpc_id
  target_type = "ip"

  health_check {
    enabled             = true
    healthy_threshold   = 2
    unhealthy_threshold = 2
    timeout             = 5
    interval            = 30
    path                = "/health"
    matcher             = "200"
  }
}

resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.main.arn
  port              = 80
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.app.arn
  }
}
```

**File:** `terraform/modules/loadbalancer/variables.tf`

```hcl
variable "app_name"              { type = string }
variable "vpc_id"                { type = string }
variable "public_subnet_ids"     { type = list(string) }
variable "alb_security_group_id" { type = string }
```

**File:** `terraform/modules/loadbalancer/outputs.tf`

```hcl
output "alb_dns_name"      { value = aws_lb.main.dns_name }
output "target_group_arn"  { value = aws_lb_target_group.app.arn }
output "alb_arn"           { value = aws_lb.main.arn }
```

### 6.9 — Database Module

**File:** `terraform/modules/database/main.tf`

```hcl
resource "aws_db_subnet_group" "main" {
  name       = "${var.app_name}-db-subnet-group"
  subnet_ids = var.private_subnet_ids
  tags       = { Name = "${var.app_name}-db-subnet-group" }
}

resource "aws_db_instance" "main" {
  identifier        = "${var.app_name}-db"
  engine            = "postgres"
  engine_version    = "16.3"
  instance_class    = "db.t3.micro"
  allocated_storage = 20
  storage_type      = "gp2"

  db_name  = "tasksdb"
  username = "devops"
  password = var.db_password

  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids = [var.rds_security_group_id]
  publicly_accessible    = false

  backup_retention_period = 0
  skip_final_snapshot     = true
  deletion_protection     = false

  tags = { Name = "${var.app_name}-db" }
}
```

**File:** `terraform/modules/database/variables.tf`

```hcl
variable "app_name"              { type = string }
variable "private_subnet_ids"    { type = list(string) }
variable "rds_security_group_id" { type = string }
variable "db_password"           { type = string; sensitive = true }
```

**File:** `terraform/modules/database/outputs.tf`

```hcl
output "db_endpoint" { value = aws_db_instance.main.endpoint; sensitive = true }
output "db_name"     { value = aws_db_instance.main.db_name }
output "db_username" { value = aws_db_instance.main.username }
```

### 6.10 — Compute Module (ECS Fargate)

**File:** `terraform/modules/compute/main.tf`

```hcl
resource "aws_ecs_cluster" "main" {
  name = "${var.app_name}-cluster"
  setting { name = "containerInsights"; value = "disabled" }
}

resource "aws_iam_role" "ecs_execution" {
  name = "${var.app_name}-ecs-execution-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "ecs-tasks.amazonaws.com" }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "ecs_execution" {
  role       = aws_iam_role.ecs_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_iam_role_policy" "ecs_ecr" {
  name = "${var.app_name}-ecs-ecr-policy"
  role = aws_iam_role.ecs_execution.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = [
        "ecr:GetAuthorizationToken",
        "ecr:BatchCheckLayerAvailability",
        "ecr:GetDownloadUrlForLayer",
        "ecr:BatchGetImage"
      ]
      Resource = "*"
    }]
  })
}

resource "aws_cloudwatch_log_group" "app" {
  name              = "/ecs/${var.app_name}"
  retention_in_days = 7
}

resource "aws_ecs_task_definition" "app" {
  family                   = var.app_name
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.ecs_execution.arn

  container_definitions = jsonencode([{
    name      = var.app_name
    image     = var.docker_image
    essential = true

    portMappings = [{ containerPort = 8080; protocol = "tcp" }]

    environment = [
      { name = "PORT",    value = "8080" },
      { name = "DB_HOST", value = var.db_endpoint },
      { name = "DB_NAME", value = var.db_name },
      { name = "DB_USER", value = var.db_username }
    ]

    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.app.name
        "awslogs-region"        = var.aws_region
        "awslogs-stream-prefix" = "ecs"
      }
    }

    healthCheck = {
      command     = ["CMD-SHELL", "wget -q --spider http://localhost:8080/health || exit 1"]
      interval    = 30
      timeout     = 5
      retries     = 3
      startPeriod = 10
    }
  }])
}

resource "aws_ecs_service" "app" {
  name            = "${var.app_name}-service"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.app.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = var.public_subnet_ids
    security_groups  = [var.ecs_security_group_id]
    assign_public_ip = true
  }

  load_balancer {
    target_group_arn = var.target_group_arn
    container_name   = var.app_name
    container_port   = 8080
  }

  depends_on = [var.alb_listener_arn]

  lifecycle {
    ignore_changes = [task_definition]
  }
}
```

**File:** `terraform/modules/compute/variables.tf`

```hcl
variable "app_name"              { type = string }
variable "aws_region"            { type = string }
variable "docker_image"          { type = string }
variable "private_subnet_ids"    { type = list(string) }
variable "public_subnet_ids"     { type = list(string) }
variable "ecs_security_group_id" { type = string }
variable "target_group_arn"      { type = string }
variable "alb_listener_arn"      { type = string }
variable "db_endpoint"           { type = string }
variable "db_name"               { type = string }
variable "db_username"           { type = string }
```

**File:** `terraform/modules/compute/outputs.tf`

```hcl
output "cluster_name"        { value = aws_ecs_cluster.main.name }
output "service_name"        { value = aws_ecs_service.app.name }
output "task_definition_arn" { value = aws_ecs_task_definition.app.arn }
```

### 6.11 — Deploy the Infrastructure

```bash
cd ~/projects/go-devops-app/terraform

# Initialize Terraform (downloads AWS provider, connects to S3 backend)
terraform init

# Validate all .tf files for syntax errors
terraform validate

# Preview what will be created (nothing created yet)
terraform plan -var-file=terraform.tfvars

# CREATE EVERYTHING on AWS (takes ~10 minutes, costs ~$1-2/day)
terraform apply -var-file=terraform.tfvars
# Type: yes
```

### 6.12 — Push Docker Image to ECR

```bash
# Get your AWS account ID
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)

# Login to ECR
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS \
  --password-stdin ${ACCOUNT_ID}.dkr.ecr.us-east-1.amazonaws.com

# Build and tag for ECR
cd ~/projects/go-devops-app
docker build -t go-devops-app:latest .
docker tag go-devops-app:latest \
  ${ACCOUNT_ID}.dkr.ecr.us-east-1.amazonaws.com/go-devops-app:latest

# Push to ECR
docker push ${ACCOUNT_ID}.dkr.ecr.us-east-1.amazonaws.com/go-devops-app:latest
```

### 6.13 — Test the Live App

```bash
# Get ALB URL from Terraform outputs
cd ~/projects/go-devops-app/terraform
ALB=$(terraform output -raw alb_dns_name)

# Test health endpoint
curl http://$ALB/health

# Create a task
curl -X POST http://$ALB/tasks \
  -H "Content-Type: application/json" \
  -d '{"title": "First task on AWS!"}'

# List all tasks
curl http://$ALB/tasks
```

### 6.14 — Useful AWS Commands

```bash
# Check ECS service status
aws ecs describe-services \
  --cluster go-devops-app-cluster \
  --services go-devops-app-service \
  --query 'services[0].{Status:status,Running:runningCount,Desired:desiredCount}'

# View live container logs
aws logs tail /ecs/go-devops-app --follow

# Force a new deployment (rolling update, zero downtime)
aws ecs update-service \
  --cluster go-devops-app-cluster \
  --service go-devops-app-service \
  --force-new-deployment
```

---

<a name="phase-7"></a>
## 📊 PHASE 7 — Monitoring with Prometheus & Grafana

> **Goal:** Collect and visualize real-time metrics from your production AWS app.

### 7.1 — Prometheus Configuration

**File:** `monitoring/prometheus.yml`

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "alert_rules.yml"

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

  - job_name: 'go-devops-app-local'
    static_configs:
      - targets: ['app:8080']
    metrics_path: '/metrics'

  # Replace with your actual ALB DNS name
  - job_name: 'go-devops-app-aws'
    static_configs:
      - targets: ['YOUR-ALB-DNS.us-east-1.elb.amazonaws.com']
    metrics_path: '/metrics'
    scheme: http
```

### 7.2 — Alert Rules

**File:** `monitoring/alert_rules.yml`

```yaml
groups:
  - name: go-devops-app
    rules:
      - alert: AppDown
        expr: up{job="go-devops-app-aws"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "go-devops-app is DOWN on AWS"

      - alert: HighGCPause
        expr: go_gc_duration_seconds{quantile="0.99"} > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High GC pause time detected"

      - alert: TooManyGoroutines
        expr: go_goroutines > 100
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Unusually high goroutine count"
```

### 7.3 — Grafana Auto-Provisioning

```bash
mkdir -p monitoring/grafana/provisioning/datasources
mkdir -p monitoring/grafana/provisioning/dashboards
mkdir -p monitoring/grafana/dashboards
```

**File:** `monitoring/grafana/provisioning/datasources/prometheus.yml`

```yaml
apiVersion: 1
datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: false
```

**File:** `monitoring/grafana/provisioning/dashboards/default.yml`

```yaml
apiVersion: 1
providers:
  - name: 'default'
    orgId: 1
    folder: ''
    type: file
    disableDeletion: false
    updateIntervalSeconds: 10
    options:
      path: /var/lib/grafana/dashboards
```

### 7.4 — Start Monitoring Stack

```bash
cd ~/projects/go-devops-app
docker compose up -d prometheus grafana
```

### 7.5 — Access Dashboards

| URL | What |
|-----|------|
| `http://localhost:9090/targets` | Prometheus — verify AWS app is being scraped |
| `http://localhost:3000` | Grafana — admin / admin123 |

**Import Go dashboard in Grafana:**
1. Left sidebar → **+** → **Import dashboard**
2. Enter ID: **`6671`** → Load
3. Select Prometheus datasource → Import

### 7.6 — Generate Test Traffic

```bash
ALB="YOUR-ALB-DNS.us-east-1.elb.amazonaws.com"

for i in {1..20}; do
  curl -s http://$ALB/health > /dev/null
  curl -s -X POST http://$ALB/tasks \
    -H "Content-Type: application/json" \
    -d "{\"title\": \"Task $i\"}" > /dev/null
  echo "Request $i sent"
  sleep 1
done
```

Watch metrics move in Grafana in real time!

---

<a name="phase-8-9"></a>
## 🔐 PHASE 8 & 9 — Security & Production Hardening

### 9.1 — Store Secrets in AWS Secrets Manager

```bash
aws secretsmanager create-secret \
  --name "go-devops-app/db-password" \
  --description "RDS PostgreSQL password for go-devops-app" \
  --secret-string "DevOpsSecure2024" \
  --region us-east-1

# Verify
aws secretsmanager get-secret-value \
  --secret-id "go-devops-app/db-password" \
  --query SecretString \
  --output text
```

### 9.2 — Enable CloudTrail Audit Logging

```bash
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
BUCKET="go-devops-terraform-state-${ACCOUNT_ID}"

# Add bucket policy for CloudTrail
aws s3api put-bucket-policy \
  --bucket $BUCKET \
  --policy "{
    \"Version\": \"2012-10-17\",
    \"Statement\": [
      {
        \"Sid\": \"AWSCloudTrailAclCheck\",
        \"Effect\": \"Allow\",
        \"Principal\": {\"Service\": \"cloudtrail.amazonaws.com\"},
        \"Action\": \"s3:GetBucketAcl\",
        \"Resource\": \"arn:aws:s3:::${BUCKET}\"
      },
      {
        \"Sid\": \"AWSCloudTrailWrite\",
        \"Effect\": \"Allow\",
        \"Principal\": {\"Service\": \"cloudtrail.amazonaws.com\"},
        \"Action\": \"s3:PutObject\",
        \"Resource\": \"arn:aws:s3:::${BUCKET}/cloudtrail/AWSLogs/${ACCOUNT_ID}/*\",
        \"Condition\": {
          \"StringEquals\": {\"s3:x-amz-acl\": \"bucket-owner-full-control\"}
        }
      }
    ]
  }"

# Create trail
aws cloudtrail create-trail \
  --name go-devops-audit-trail \
  --s3-bucket-name $BUCKET \
  --s3-key-prefix cloudtrail \
  --include-global-service-events \
  --is-multi-region-trail

# Start logging
aws cloudtrail start-logging --name go-devops-audit-trail

# Verify
aws cloudtrail get-trail-status \
  --name go-devops-audit-trail \
  --query '{Logging:IsLogging}'
```

### 9.3 — Create CloudWatch Alarms

```bash
# Alarm 1: ECS CPU too high
aws cloudwatch put-metric-alarm \
  --alarm-name "go-devops-app-high-cpu" \
  --alarm-description "ECS CPU above 80% for 5 minutes" \
  --metric-name CPUUtilization \
  --namespace AWS/ECS \
  --dimensions Name=ClusterName,Value=go-devops-app-cluster \
              Name=ServiceName,Value=go-devops-app-service \
  --statistic Average \
  --period 300 \
  --threshold 80 \
  --comparison-operator GreaterThanThreshold \
  --evaluation-periods 2 \
  --treat-missing-data notBreaching

# Alarm 2: ALB 5xx errors
aws cloudwatch put-metric-alarm \
  --alarm-name "go-devops-app-5xx-errors" \
  --alarm-description "ALB returning 5xx errors" \
  --metric-name HTTPCode_Target_5XX_Count \
  --namespace AWS/ApplicationELB \
  --dimensions Name=LoadBalancer,Value=$(aws elbv2 describe-load-balancers \
    --names go-devops-app-alb \
    --query 'LoadBalancers[0].LoadBalancerArn' \
    --output text | cut -d'/' -f2-) \
  --statistic Sum \
  --period 60 \
  --threshold 10 \
  --comparison-operator GreaterThanThreshold \
  --evaluation-periods 1 \
  --treat-missing-data notBreaching

# Alarm 3: RDS connections
aws cloudwatch put-metric-alarm \
  --alarm-name "go-devops-app-rds-connections" \
  --alarm-description "RDS connections above 50" \
  --metric-name DatabaseConnections \
  --namespace AWS/RDS \
  --dimensions Name=DBInstanceIdentifier,Value=go-devops-app-db \
  --statistic Average \
  --period 300 \
  --threshold 50 \
  --comparison-operator GreaterThanThreshold \
  --evaluation-periods 1 \
  --treat-missing-data notBreaching

# Verify all 3 alarms
aws cloudwatch describe-alarms \
  --alarm-names \
    "go-devops-app-high-cpu" \
    "go-devops-app-5xx-errors" \
    "go-devops-app-rds-connections" \
  --query 'MetricAlarms[*].{Name:AlarmName,State:StateValue}'
```

---

<a name="phase-10"></a>
## 🎬 PHASE 10 — Demo & Cleanup

### 10.1 — Full API Demo

```bash
ALB="YOUR-ALB-DNS.us-east-1.elb.amazonaws.com"

# Health check
curl http://$ALB/health

# Create tasks
curl -s -X POST http://$ALB/tasks \
  -H "Content-Type: application/json" \
  -d '{"title": "Learn DevOps"}' | python3 -m json.tool

curl -s -X POST http://$ALB/tasks \
  -H "Content-Type: application/json" \
  -d '{"title": "Deploy to AWS"}' | python3 -m json.tool

# List all tasks
curl -s http://$ALB/tasks | python3 -m json.tool

# Delete a task (replace 1 with actual ID)
curl -X DELETE http://$ALB/tasks/1

# View raw Prometheus metrics
curl -s http://$ALB/metrics | grep -E "^go_goroutines|^go_memstats_alloc_bytes "
```

### 10.2 — Zero-Downtime Deployment

```bash
# Force ECS to pull latest image and restart (rolling update)
aws ecs update-service \
  --cluster go-devops-app-cluster \
  --service go-devops-app-service \
  --force-new-deployment

# Monitor the rolling update
aws ecs describe-services \
  --cluster go-devops-app-cluster \
  --services go-devops-app-service \
  --query 'services[0].{Status:status,Running:runningCount,Desired:desiredCount,Pending:pendingCount}'
```

### 10.3 — Destroy All AWS Resources (STOP BILLING)

> ⚠️ **Run this when you're done to avoid charges!**

```bash
# Step 1: Delete ECR images and repo first
aws ecr delete-repository --repository-name go-devops-app --force

# Step 2: Destroy all Terraform-managed resources
cd ~/projects/go-devops-app/terraform
terraform destroy -var-file=terraform.tfvars
# Type: yes (takes ~5 minutes)

# Step 3: Delete CloudWatch alarms
aws cloudwatch delete-alarms --alarm-names \
  "go-devops-app-high-cpu" \
  "go-devops-app-5xx-errors" \
  "go-devops-app-rds-connections"

# Step 4: Stop and delete CloudTrail
aws cloudtrail stop-logging --name go-devops-audit-trail
aws cloudtrail delete-trail --name go-devops-audit-trail

# Step 5: Verify everything is gone
aws ecs list-clusters --query 'clusterArns'
aws rds describe-db-instances --query 'DBInstances[*].DBInstanceIdentifier'
aws elbv2 describe-load-balancers --query 'LoadBalancers[*].LoadBalancerName'
# All should return []
```

### 10.4 — Stop Local Docker

```bash
cd ~/projects/go-devops-app
docker compose down
```

---

<a name="quick-reference"></a>
## 📌 Quick Reference Card

### Daily Development Workflow

```bash
# Start coding
sudo service docker start          # Start Docker
cd ~/projects/go-devops-app

# Work on a feature
git checkout -b feat/my-feature
# ... make changes ...
gofmt -w ./...                     # Format all Go files
make test                          # Run tests
git add . && git commit -m "feat: ..."
git push origin feat/my-feature
# Create PR on GitHub → CI runs → merge
git checkout main && git pull origin main
```

### Common Commands

```bash
# Local development
make run          # Start server at :8080
make test         # Run all unit tests
make lint         # Run linter (slow on WSL2)
make docker-build # Build Docker image

# Docker
docker compose up -d               # Start full stack
docker compose down                # Stop everything
docker compose logs -f app         # Follow app logs
docker compose ps                  # List containers

# AWS
aws ecs describe-services --cluster go-devops-app-cluster \
  --services go-devops-app-service \
  --query 'services[0].{Running:runningCount,Status:status}'
aws logs tail /ecs/go-devops-app --follow
aws ecs update-service --cluster go-devops-app-cluster \
  --service go-devops-app-service --force-new-deployment

# Terraform
cd terraform/
terraform plan -var-file=terraform.tfvars    # Preview changes
terraform apply -var-file=terraform.tfvars   # Apply changes
terraform destroy -var-file=terraform.tfvars # Delete everything
terraform output                             # Show outputs
```

### Project File Map

```
go-devops-app/
├── cmd/server/main.go              ← App entry point + graceful shutdown
├── internal/
│   ├── handlers/
│   │   ├── health.go               ← GET /health
│   │   ├── tasks.go                ← GET/POST/DELETE /tasks
│   │   └── tasks_test.go           ← Unit tests
│   └── middleware/
│       └── logger.go               ← Request logging middleware
├── terraform/
│   ├── main.tf                     ← Root module (wires all modules)
│   ├── variables.tf                ← Input variables
│   ├── outputs.tf                  ← Output values
│   ├── providers.tf                ← AWS provider config
│   ├── backend.tf                  ← Remote state (S3)
│   ├── terraform.tfvars            ← Secret values (NEVER commit!)
│   └── modules/
│       ├── ecr/                    ← Container registry
│       ├── networking/             ← VPC, subnets, security groups
│       ├── loadbalancer/           ← ALB, target group, listener
│       ├── database/               ← RDS PostgreSQL
│       └── compute/                ← ECS Fargate cluster + service
├── monitoring/
│   ├── prometheus.yml              ← Scrape config
│   ├── alert_rules.yml             ← Alert conditions
│   └── grafana/
│       ├── provisioning/           ← Auto-configure Grafana
│       └── dashboards/             ← Dashboard JSON files
├── .github/workflows/
│   ├── ci.yml                      ← Lint + Test + Scan + Build
│   └── cd.yml                      ← Build + Push to Docker Hub
├── Dockerfile                      ← Multi-stage container build
├── docker-compose.yml              ← Local dev stack
├── .golangci.yml                   ← Linter configuration
├── .gitignore                      ← Files to exclude from git
├── .dockerignore                   ← Files to exclude from Docker
├── Makefile                        ← Developer shortcuts
├── go.mod                          ← Go module definition
└── go.sum                          ← Dependency checksums
```

### Security Checklist

```
✅ Non-root container user in Dockerfile
✅ Multi-stage Docker build (no SDK in production image)
✅ Private subnets for RDS
✅ Security groups (least-privilege firewall rules)
✅ No secrets in code (terraform.tfvars in .gitignore)
✅ Secrets in AWS Secrets Manager
✅ S3 state bucket encrypted + public access blocked
✅ CloudTrail audit logging enabled
✅ Trivy CVE scanning in CI pipeline
✅ Branch protection (all changes via PR)
✅ Go 1.24 (patched CVE-2025-68121)
✅ HTTP server timeouts (read: 15s, write: 15s, idle: 60s)
✅ Graceful shutdown (30s drain window)
✅ CloudWatch alarms (CPU, errors, DB connections)
```

### What to Learn Next

| Topic | What to build |
|-------|--------------|
| **HTTPS/SSL** | Add ACM certificate to ALB — free with AWS Certificate Manager |
| **Auto-scaling** | Scale ECS tasks 1→N based on CPU/memory |
| **GitHub Actions CD to AWS** | Deploy to ECS automatically on merge |
| **Kubernetes (EKS)** | Replace ECS with Kubernetes |
| **Helm Charts** | Package app for Kubernetes |
| **Terraform Modules Registry** | Publish your modules publicly |
| **AWS WAF** | Web Application Firewall for ALB |
| **VPC Endpoints** | Private ECR access without NAT Gateway |

---

*Built with ❤️ through hands-on practice — the only way to truly learn DevOps.*
