package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

// TestGetTasksEmpty verifies GET /tasks returns empty list when no tasks exist
func TestGetTasksEmpty(t *testing.T) {
	// Arrange: fresh store with no tasks
	store := NewTaskStore()

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	rr := httptest.NewRecorder()

	// Act: call handler directly
	store.GetTasks(rr, req)

	// Assert: status must be 200
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	// Assert: body must be empty array
	var tasks []Task
	json.NewDecoder(rr.Body).Decode(&tasks)
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(tasks))
	}
}

// TestCreateTask verifies POST /tasks creates a task and returns 201
func TestCreateTask(t *testing.T) {
	store := NewTaskStore()

	body := bytes.NewBufferString(`{"title": "Buy groceries"}`)
	req := httptest.NewRequest(http.MethodPost, "/tasks", body)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	store.CreateTask(rr, req)

	// Check 201 Created
	if rr.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rr.Code)
	}

	// Decode and verify fields
	var task Task
	json.NewDecoder(rr.Body).Decode(&task)

	if task.Title != "Buy groceries" {
		t.Errorf("expected title 'Buy groceries', got '%s'", task.Title)
	}
	if task.ID != 1 {
		t.Errorf("expected ID 1, got %d", task.ID)
	}
	if task.Done != false {
		t.Errorf("expected Done=false, got true")
	}
	if task.CreatedAt == "" {
		t.Error("expected CreatedAt to be set, got empty string")
	}
}

// TestCreateTaskMissingTitle verifies empty title returns 400
func TestCreateTaskMissingTitle(t *testing.T) {
	store := NewTaskStore()

	body := bytes.NewBufferString(`{"title": ""}`)
	req := httptest.NewRequest(http.MethodPost, "/tasks", body)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	store.CreateTask(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

// TestDeleteTask verifies DELETE /tasks/{id} removes a task and returns 204
func TestDeleteTask(t *testing.T) {
	store := NewTaskStore()

	// Step 1: create a task first so we have something to delete
	body := bytes.NewBufferString(`{"title": "Task to delete"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/tasks", body)
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	store.CreateTask(createRR, createReq)

	// Step 2: use a chi router to handle the delete
	// chi automatically extracts {id} from the URL and puts it in request context
	r := chi.NewRouter()
	r.Delete("/{id}", store.DeleteTask)

	deleteRR := httptest.NewRecorder()
	r.ServeHTTP(deleteRR, httptest.NewRequest(http.MethodDelete, "/1", nil))

	// 204 No Content = deleted successfully
	if deleteRR.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", deleteRR.Code)
	}
}

// TestDeleteTaskNotFound verifies 404 when deleting non-existent task
func TestDeleteTaskNotFound(t *testing.T) {
	store := NewTaskStore()

	r := chi.NewRouter()
	r.Delete("/{id}", store.DeleteTask)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, httptest.NewRequest(http.MethodDelete, "/999", nil))

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}
}
