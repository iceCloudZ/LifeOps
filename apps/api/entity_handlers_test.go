package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateAndGetTask(t *testing.T) {
	srv := testServer(t)

	req := authRequest(http.MethodPost, "/api/tasks/", `{"title":"交水费","description":"提醒交水费"}`)
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

	req = authRequest(http.MethodGet, "/api/tasks/"+created.ID, "")
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
		req := authRequest(http.MethodPost, "/api/tasks/", `{"title":"`+title+`"}`)
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("create task failed: %d", w.Code)
		}
	}

	req := authRequest(http.MethodGet, "/api/tasks/", "")
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

	req := authRequest(http.MethodPost, "/api/tasks/", `{"title":"old"}`)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	var created FamilyTask
	json.Unmarshal(w.Body.Bytes(), &created)

	req = authRequest(http.MethodPut, "/api/tasks/"+created.ID, `{"title":"new","description":"updated","status":"done"}`)
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

	req := authRequest(http.MethodPost, "/api/tasks/", `{"title":"to-delete"}`)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	var created FamilyTask
	json.Unmarshal(w.Body.Bytes(), &created)

	req = authRequest(http.MethodDelete, "/api/tasks/"+created.ID, "")
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}

	req = authRequest(http.MethodGet, "/api/tasks/"+created.ID, "")
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", w.Code)
	}
}

func TestCreateEntityMissingTitle(t *testing.T) {
	srv := testServer(t)

	for _, endpoint := range []string{"/api/tasks/", "/api/events/", "/api/notes/"} {
		req := authRequest(http.MethodPost, endpoint, `{"description":"no title"}`)
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for %s, got %d", endpoint, w.Code)
		}
	}

	req := authRequest(http.MethodPost, "/api/shopping-items/", `{"quantity":"2"}`)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for shopping-items, got %d", w.Code)
	}
}

func TestCreateShoppingItemHandler(t *testing.T) {
	srv := testServer(t)

	req := authRequest(http.MethodPost, "/api/shopping-items/", `{"name":"牛奶","quantity":"2盒"}`)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var item ShoppingItem
	json.Unmarshal(w.Body.Bytes(), &item)
	if item.Name != "牛奶" || item.Quantity != "2盒" || item.Status != "open" {
		t.Fatalf("unexpected item: %+v", item)
	}
}

func TestCreateEventHandler(t *testing.T) {
	srv := testServer(t)

	req := authRequest(http.MethodPost, "/api/events/", `{"title":"家长会","description":"下周二","starts_at":"2026-06-16T18:30:00Z"}`)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateNoteHandler(t *testing.T) {
	srv := testServer(t)

	req := authRequest(http.MethodPost, "/api/notes/", `{"title":"晚餐偏好","content":"少油少辣","topic":"food"}`)
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
		req := authRequest(http.MethodGet, endpoint, "")
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 for %s, got %d", endpoint, w.Code)
		}
	}
}

func TestEntityRequiresAuth(t *testing.T) {
	srv := testServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks/", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without auth, got %d", w.Code)
	}
}
