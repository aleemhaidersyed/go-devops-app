package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
)

// Task represents a single task item
type Task struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Done      bool   `json:"done"`
	CreatedAt string `json:"created_at"`
}

// TaskStore is our in-memory "database"
// sync.Mutex ensures only one goroutine accesses it at a time (thread-safe)
type TaskStore struct {
	mu      sync.Mutex
	tasks   []Task
	counter int
}

// NewTaskStore creates and returns a new TaskStore
func NewTaskStore() *TaskStore {
	return &TaskStore{}
}

// GetTasks handles GET /tasks — returns all tasks as JSON
func (s *TaskStore) GetTasks(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()         // Lock: no other goroutine can touch tasks now
	defer s.mu.Unlock() // Unlock when this function returns (no matter what)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(s.tasks); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// CreateTask handles POST /tasks — creates a new task from JSON body
func (s *TaskStore) CreateTask(w http.ResponseWriter, r *http.Request) {
	// Decode the JSON body into a Task struct
	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate: title must not be empty
	if task.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Assign auto-incrementing ID
	s.counter++
	task.ID = s.counter
	task.Done = false
	task.CreatedAt = time.Now().UTC().Format(time.RFC3339)

	// Add to our slice
	s.tasks = append(s.tasks, task)

	// Respond with 201 Created and the new task
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(task); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// DeleteTask handles DELETE /tasks/{id} — removes a task by ID
func (s *TaskStore) DeleteTask(w http.ResponseWriter, r *http.Request) {
	// Extract the {id} from the URL using chi's URL param reader
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr) // Convert string "3" to integer 3
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Find and remove the task
	for i, t := range s.tasks {
		if t.ID == id {
			// Remove element at index i by slicing around it
			s.tasks = append(s.tasks[:i], s.tasks[i+1:]...)
			w.WriteHeader(http.StatusNoContent) // 204 = deleted successfully, no body
			return
		}
	}

	// If we get here, the ID wasn't found
	http.Error(w, "Task not found", http.StatusNotFound)
}
