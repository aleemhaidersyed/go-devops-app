package main

import (
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/aleemhaider/go-devops-app/internal/handlers"
	"github.com/aleemhaider/go-devops-app/internal/middleware"
)

func main() {
	// ── Configure Logger ──────────────────────────────────────────
	// zerolog outputs JSON by default (great for production log aggregators)
	// During development, we use ConsoleWriter for human-readable output
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	log.Info().Msg("starting go-devops-app")

	// ── Create Router ─────────────────────────────────────────────
	// chi is a lightweight HTTP router — it maps URL patterns to handler functions
	r := chi.NewRouter()

	// ── Register Global Middleware ────────────────────────────────
	// These run for EVERY request, in order
	r.Use(chimiddleware.RequestID) // Assigns a unique ID to every request
	r.Use(chimiddleware.Recoverer) // If a handler panics, recover gracefully (don't crash)
	r.Use(middleware.Logger)       // Our custom request logger

	// ── Create Shared State ───────────────────────────────────────
	// TaskStore holds all tasks in memory — shared across all requests
	store := handlers.NewTaskStore()

	// ── Register Routes ───────────────────────────────────────────
	// GET  /health       → HealthHandler
	// GET  /tasks        → store.GetTasks
	// POST /tasks        → store.CreateTask
	// DELETE /tasks/{id} → store.DeleteTask
	// GET  /metrics      → Prometheus metrics (auto-generated)

	r.Get("/health", handlers.HealthHandler)

	r.Route("/tasks", func(r chi.Router) {
		r.Get("/", store.GetTasks)
		r.Post("/", store.CreateTask)
		r.Delete("/{id}", store.DeleteTask)
	})

	// Prometheus metrics endpoint — exposes all metrics at /metrics
	r.Handle("/metrics", promhttp.Handler())

	// ── Start the Server ──────────────────────────────────────────
	port := ":8080"
	log.Info().Str("port", port).Msg("server listening")

	// ListenAndServe blocks forever, handling incoming requests
	// If it returns, something went wrong — log and exit
	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatal().Err(err).Msg("server failed to start")
	}
}
