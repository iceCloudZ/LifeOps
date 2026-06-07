package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateAndGetTask(t *testing.T) {
	srv := testServer(t)

	body := bytes.NewBufferString(`{"title":"交水费","description":"提醒交水费"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/tasks/", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var created FamilyTask
	json.Unmarshal(w.Body.Bytes(), &created)
	if created.Title != "交水费" {
		t.Fatalf("expected title 交水费, got %q", created.Title)
	}
	if created.ID == "" {
		t.Fatal("expected non-empty id")
	}

	req = httptest.NewRequest(http.MethodGet, "/api/tasks/"+created.ID, nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var got FamilyTask
	json.Unmarshal(w.Body.Bytes(), &got)
	if got.Title != "交水费" {
		t.Fatalf("expected title 交水费, got %q", got.Title)
	}
}

func TestListTasks(t *testing.T) {
	srv := testServer(t)

	for _, title := range []string{"task1", "task2"} {
		body := bytes.NewBufferString(`{"title":"` + title + `"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/tasks/", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("create task failed: %d", w.Code)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/api/tasks/", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var tasks []FamilyTask
	json.Unmarshal(w.Body.Bytes(), &tasks)
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
}

func TestUpdateTask(t *testing.T) {
	srv := testServer(t)

	body := bytes.NewBufferString(`{"title":"old"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/tasks/", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	var created FamilyTask
	json.Unmarshal(w.Body.Bytes(), &created)

	body = bytes.NewBufferString(`{"title":"new","description":"updated","status":"done"}`)
	req = httptest.NewRequest(http.MethodPut, "/api/tasks/"+created.ID, body)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var updated FamilyTask
	json.Unmarshal(w.Body.Bytes(), &updated)
	if updated.Title != "new" {
		t.Fatalf("expected title new, got %q", updated.Title)
	}
}

func TestDeleteTask(t *testing.T) {
	srv := testServer(t)

	body := bytes.NewBufferString(`{"title":"to-delete"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/tasks/", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	var created FamilyTask
	json.Unmarshal(w.Body.Bytes(), &created)

	req = httptest.NewRequest(http.MethodDelete, "/api/tasks/"+created.ID, nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/tasks/"+created.ID, nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", w.Code)
	}
}

func TestCreateEntityMissingTitle(t *testing.T) {
	srv := testServer(t)

	for _, endpoint := range []string{"/api/tasks/", "/api/events/", "/api/notes/"} {
		body := bytes.NewBufferString(`{"description":"no title"}`)
		req := httptest.NewRequest(http.MethodPost, endpoint, body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for %s, got %d", endpoint, w.Code)
		}
	}

	body := bytes.NewBufferString(`{"quantity":"2"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/shopping-items/", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for shopping-items, got %d", w.Code)
	}
}

func TestCreateShoppingItemHandler(t *testing.T) {
	srv := testServer(t)

	body := bytes.NewBufferString(`{"name":"牛奶","quantity":"2盒"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/shopping-items/", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var item ShoppingItem
	json.Unmarshal(w.Body.Bytes(), &item)
	if item.Name != "牛奶" {
		t.Fatalf("expected name 牛奶, got %q", item.Name)
	}
	if item.Quantity != "2盒" {
		t.Fatalf("expected quantity 2盒, got %q", item.Quantity)
	}
	if item.Status != "open" {
		t.Fatalf("expected status open, got %q", item.Status)
	}
}

func TestCreateEventHandler(t *testing.T) {
	srv := testServer(t)

	body := bytes.NewBufferString(`{"title":"家长会","description":"下周二","starts_at":"2026-06-16T18:30:00Z"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/events/", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateNoteHandler(t *testing.T) {
	srv := testServer(t)

	body := bytes.NewBufferString(`{"title":"晚餐偏好","content":"少油少辣","topic":"food"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/notes/", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var note FamilyNote
	json.Unmarshal(w.Body.Bytes(), &note)
	if note.Topic != "food" {
		t.Fatalf("expected topic food, got %q", note.Topic)
	}
}

func TestGetEntityNotFound(t *testing.T) {
	srv := testServer(t)

	for _, endpoint := range []string{"/api/tasks/nonexistent", "/api/events/nonexistent", "/api/shopping-items/nonexistent", "/api/notes/nonexistent"} {
		req := httptest.NewRequest(http.MethodGet, endpoint, nil)
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 for %s, got %d", endpoint, w.Code)
		}
	}
}
