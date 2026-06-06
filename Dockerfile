# ════════════════════════════════════════════════════════════════
# STAGE 1: builder
# Use the official Go image — it has everything needed to compile
# ════════════════════════════════════════════════════════════════
FROM golang:1.23-alpine AS builder

# Why alpine? It's a minimal Linux (5MB vs 200MB for full Ubuntu)
# The builder stage is temporary — it won't be in the final image

# Install git (needed by go mod to fetch some dependencies)
RUN apk add --no-cache git

# Set the working directory inside the container
# All subsequent commands run from here
WORKDIR /app

# ── Copy dependency files FIRST (smart caching) ──────────────────
# Docker builds in layers. If go.mod/go.sum haven't changed,
# Docker reuses the cached layer and skips downloading dependencies.
# This makes rebuilds much faster during development.
COPY go.mod go.sum ./

# Download all dependencies
# This layer is cached unless go.mod or go.sum change
RUN go mod download

# ── Copy the rest of the source code ─────────────────────────────
# This is done AFTER downloading deps so a code change doesn't
# invalidate the dependency cache layer
COPY . .

# ── Build the binary ──────────────────────────────────────────────
# CGO_ENABLED=0  = disable C bindings (produces a fully static binary)
# GOOS=linux     = build for Linux (even if you're building on Windows/Mac)
# GOARCH=amd64   = build for 64-bit x86 (standard AWS server architecture)
# -ldflags="-w -s" = strip debug info and symbol table (smaller binary)
# -o /app/server = output the binary to this path
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags="-w -s" \
    -o /app/server \
    ./cmd/server/

# ════════════════════════════════════════════════════════════════
# STAGE 2: runner (the final image)
# Start completely fresh from a tiny base image
# Nothing from stage 1 carries over except what we explicitly copy
# ════════════════════════════════════════════════════════════════
FROM alpine:3.19 AS runner

# Install CA certificates so our app can make HTTPS calls
# (alpine doesn't include them by default)
RUN apk add --no-cache ca-certificates tzdata

# Create a non-root user to run the app
# NEVER run production apps as root — security best practice
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set working directory
WORKDIR /app

# ── Copy ONLY the compiled binary from the builder stage ─────────
# --from=builder = take this file from stage 1
# /app/server    = path in the builder stage
# ./server       = path in this final image
COPY --from=builder /app/server ./server

# Give ownership of the app directory to our non-root user
RUN chown -R appuser:appgroup /app

# Switch to the non-root user
USER appuser

# Tell Docker this container listens on port 8080
# (This is documentation — it doesn't actually open the port)
EXPOSE 8080

# Health check: Docker itself will ping /health every 30 seconds
# If it fails 3 times in a row, Docker marks the container as unhealthy
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# The command to run when the container starts
# Use exec form (JSON array) — not shell form — so signals work correctly
# This is important for graceful shutdown (Ctrl+C, SIGTERM from orchestrators)
ENTRYPOINT ["./server"]
