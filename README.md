# 🚀 go-devops-app

A production-style Go REST API built as an end-to-end DevOps learning project.
Demonstrates a complete pipeline from local development to AWS deployment.

## Tech Stack
- **Language:** Go 1.22
- **Router:** chi
- **Logging:** zerolog (structured JSON)
- **Metrics:** Prometheus
- **CI/CD:** GitHub Actions
- **Infrastructure:** Terraform + AWS ECS Fargate
- **Monitoring:** Prometheus + Grafana

## API Endpoints

| Method | Endpoint    | Description         |
|--------|-------------|---------------------|
| GET    | /health     | Health check        |
| GET    | /tasks      | List all tasks      |
| POST   | /tasks      | Create a new task   |
| DELETE | /tasks/{id} | Delete a task by ID |
| GET    | /metrics    | Prometheus metrics  |

## Quick Start

### Prerequisites
- Go 1.22+
- Docker
- Make

### Run locally
```bash
make run
